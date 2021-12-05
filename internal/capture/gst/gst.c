#include "gst.h"

void gstreamer_init(void) {
  gst_init(NULL, NULL);
}

GMainLoop *gstreamer_main_loop = NULL;
void gstreamer_loop(void) {
  gstreamer_main_loop = g_main_loop_new (NULL, FALSE);
  g_main_loop_run(gstreamer_main_loop);
}

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
  GstElement *appsink = gst_bin_get_by_name(GST_BIN(ctx->pipeline), sinkName);
  g_object_set(appsink, "emit-signals", TRUE, NULL);
  g_signal_connect(appsink, "new-sample", G_CALLBACK(gstreamer_send_new_sample_handler), ctx);
  gst_object_unref(appsink);
}

void gstreamer_pipeline_play(GstPipelineCtx *ctx) {
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_PLAYING);
}

void gstreamer_pipeline_pause(GstPipelineCtx *ctx) {
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_PAUSED);
}

void gstreamer_pipeline_destory(GstPipelineCtx *ctx) {
  gst_element_set_state(GST_ELEMENT(ctx->pipeline), GST_STATE_NULL);
  gst_object_unref(ctx->pipeline);
}

void gstreamer_pipeline_push(GstPipelineCtx *ctx, char *srcName, void *buffer, int bufferLen) {
  GstElement *src = gst_bin_get_by_name(GST_BIN(ctx->pipeline), srcName);

  if (src != NULL) {
    gpointer p = g_memdup(buffer, bufferLen);
    GstBuffer *buffer = gst_buffer_new_wrapped(p, bufferLen);
    gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
    gst_object_unref(src);
  }
}
