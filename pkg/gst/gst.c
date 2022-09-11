#include "gst.h"

static void gstreamer_pipeline_log(GstPipelineCtx *ctx, char* level, const char* format, ...) {
  va_list argptr;
  va_start(argptr, format);
  char buffer[100];
  vsprintf(buffer, format, argptr);
  va_end(argptr);
  goPipelineLog(level, buffer, ctx->pipelineId);
}

static gboolean gstreamer_bus_call(GstBus *bus, GstMessage *msg, gpointer user_data) {
  GstPipelineCtx *ctx = (GstPipelineCtx *)user_data;

  switch (GST_MESSAGE_TYPE(msg)) {
    case GST_MESSAGE_EOS: {
      gstreamer_pipeline_log(ctx, "fatal", "end of stream");
      break;
    }

    case GST_MESSAGE_STATE_CHANGED: {
      GstState old_state, new_state;
      gst_message_parse_state_changed(msg, &old_state, &new_state, NULL);

      gstreamer_pipeline_log(ctx, "debug",
        "element %s changed state from %s to %s",
          GST_OBJECT_NAME(msg->src),
          gst_element_state_get_name(old_state),
          gst_element_state_get_name(new_state));
      break;
    }

    case GST_MESSAGE_TAG: {
      GstTagList *tags = NULL;
      gst_message_parse_tag(msg, &tags);

      gstreamer_pipeline_log(ctx, "debug",
        "got tags from element %s",
          GST_OBJECT_NAME(msg->src));

      gst_tag_list_unref(tags);
      break;
    }

    case GST_MESSAGE_ERROR: {
      GError *err = NULL;
      gchar *dbg_info = NULL;
      gst_message_parse_error(msg, &err, &dbg_info);

      gstreamer_pipeline_log(ctx, "error",
        "error from element %s: %s",
          GST_OBJECT_NAME(msg->src), err->message);
      gstreamer_pipeline_log(ctx, "warn",
        "debugging info: %s",
          (dbg_info) ? dbg_info : "none");

      g_error_free(err);
      g_free(dbg_info);
      break;
    }

    default:
      gstreamer_pipeline_log(ctx, "trace", "unknown message");
      break;
  }

  return TRUE;
}

GstPipelineCtx *gstreamer_pipeline_create(char *pipelineStr, int pipelineId, GError **error) {
  GstElement *pipeline = gst_parse_launch(pipelineStr, error);
  if (pipeline == NULL) return NULL;

  // create gstreamer pipeline context
  GstPipelineCtx *ctx = calloc(1, sizeof(GstPipelineCtx));
  ctx->pipelineId = pipelineId;
  ctx->pipeline = pipeline;

  GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
  gst_bus_add_watch(bus, gstreamer_bus_call, ctx);
  gst_object_unref(bus);

  return ctx;
}

static GstFlowReturn gstreamer_send_new_sample_handler(GstElement *object, gpointer user_data) {
  GstPipelineCtx *ctx = (GstPipelineCtx *)user_data;
  GstSample *sample = NULL;
  GstBuffer *buffer = NULL;
  gpointer copy = NULL;
  gsize copy_size = 0;

  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample) {
    buffer = gst_sample_get_buffer(sample);
    if (buffer) {
      gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);
      goHandlePipelineBuffer(copy, copy_size, GST_BUFFER_DURATION(buffer), ctx->pipelineId);
    }
    gst_sample_unref(sample);
  }

  return GST_FLOW_OK;
}

void gstreamer_pipeline_attach_appsink(GstPipelineCtx *ctx, char *sinkName) {
  ctx->appsink = gst_bin_get_by_name(GST_BIN(ctx->pipeline), sinkName);
  g_object_set(ctx->appsink, "emit-signals", TRUE, NULL);
  g_signal_connect(ctx->appsink, "new-sample", G_CALLBACK(gstreamer_send_new_sample_handler), ctx);
}

void gstreamer_pipeline_attach_appsrc(GstPipelineCtx *ctx, char *srcName) {
  ctx->appsrc = gst_bin_get_by_name(GST_BIN(ctx->pipeline), srcName);
}

void gstreamer_pipeline_play(GstPipelineCtx *ctx) {
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_PLAYING);
}

void gstreamer_pipeline_pause(GstPipelineCtx *ctx) {
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_PAUSED);
}

void gstreamer_pipeline_destory(GstPipelineCtx *ctx) {
  // end appsrc, if exists
  if (ctx->appsrc) {
    gst_app_src_end_of_stream(GST_APP_SRC(ctx->appsrc));
  }

  // send pipeline eos
  gst_element_send_event(GST_ELEMENT(ctx->pipeline), gst_event_new_eos());

  // set null state
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_NULL);

  if (ctx->appsink) {
    gst_object_unref(ctx->appsink);
    ctx->appsink = NULL;
  }

  if (ctx->appsrc) {
    gst_object_unref(ctx->appsrc);
    ctx->appsrc = NULL;
  }

  gst_object_unref(ctx->pipeline);
}

void gstreamer_pipeline_push(GstPipelineCtx *ctx, void *buffer, int bufferLen) {
  if (ctx->appsrc != NULL) {
    gpointer p = g_memdup(buffer, bufferLen);
    GstBuffer *buffer = gst_buffer_new_wrapped(p, bufferLen);
    gst_app_src_push_buffer(GST_APP_SRC(ctx->appsrc), buffer);
  }
}

gboolean gstreamer_pipeline_set_prop_int(GstPipelineCtx *ctx, char *binName, char *prop, gint value) {
  GstElement *el = gst_bin_get_by_name(GST_BIN(ctx->pipeline), binName);
  if (el == NULL) return FALSE;

  g_object_set(G_OBJECT(el),
    prop, value,
    NULL);

  gst_object_unref(el);
  return TRUE;
}

gboolean gstreamer_pipeline_set_caps_framerate(GstPipelineCtx *ctx, const gchar* binName, gint numerator, gint denominator) {
  GstElement *el = gst_bin_get_by_name(GST_BIN(ctx->pipeline), binName);
  if (el == NULL) return FALSE;

  GstCaps *caps = gst_caps_new_simple("video/x-raw",
    "framerate", GST_TYPE_FRACTION, numerator, denominator,
    NULL);

  g_object_set(G_OBJECT(el),
    "caps", caps,
    NULL);

  gst_caps_unref(caps);
  gst_object_unref(el);
  return TRUE;
}

gboolean gstreamer_pipeline_set_caps_resolution(GstPipelineCtx *ctx, const gchar* binName, gint width, gint height) {
  GstElement *el = gst_bin_get_by_name(GST_BIN(ctx->pipeline), binName);
  if (el == NULL) return FALSE;

  GstCaps *caps = gst_caps_new_simple("video/x-raw",
    "width", G_TYPE_INT, width,
    "height", G_TYPE_INT, height,
    NULL);

  g_object_set(G_OBJECT(el),
    "caps", caps,
    NULL);

  gst_caps_unref(caps);
  gst_object_unref(el);
  return TRUE;
}
