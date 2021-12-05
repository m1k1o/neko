#include "gst.h"

typedef struct SampleHandlerUserData {
  int pipelineId;
} SampleHandlerUserData;

void gstreamer_init(void) {
  gst_init(NULL, NULL);
}

GMainLoop *gstreamer_main_loop = NULL;
void gstreamer_loop(void) {
  gstreamer_main_loop = g_main_loop_new (NULL, FALSE);
  g_main_loop_run(gstreamer_main_loop);
}

static void gstreamer_pipeline_log(char* level, const char* format, ...) {
  va_list argptr;
  va_start(argptr, format);
  char buffer[100];
  vsprintf(buffer, format, argptr);
  va_end(argptr);
  goPipelineLog(level, buffer);
}

static gboolean gstreamer_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
  switch (GST_MESSAGE_TYPE(msg)) {
    case GST_MESSAGE_EOS: {
      gstreamer_pipeline_log("panic", "end of stream");
      exit(1);
    }

    case GST_MESSAGE_STATE_CHANGED: {
      GstState old_state, new_state;
      gst_message_parse_state_changed(msg, &old_state, &new_state, NULL);

      gstreamer_pipeline_log("debug",
        "element %s changed state from %s to %s",
          GST_OBJECT_NAME(msg->src),
          gst_element_state_get_name(old_state),
          gst_element_state_get_name(new_state));
      break;
    }

    case GST_MESSAGE_TAG: {
      GstTagList *tags = NULL;
      gst_message_parse_tag(msg, &tags);

      gstreamer_pipeline_log("debug",
        "got tags from element %s",
          GST_OBJECT_NAME(msg->src));
  
      gst_tag_list_unref(tags);
      break;
    }

    case GST_MESSAGE_ERROR: {
      GError *err = NULL;
      gchar *dbg_info = NULL;
      gst_message_parse_error(msg, &err, &dbg_info);

      gstreamer_pipeline_log("error",
        "error from element %s: %s",
          GST_OBJECT_NAME(msg->src), err->message);
      gstreamer_pipeline_log("warn",
        "debugging info: %s",
          (dbg_info) ? dbg_info : "none");

      g_error_free(err);
      g_free(dbg_info);
      break;
    }

    default:
      gstreamer_pipeline_log("trace", "unknown message");
      break;
  }

  return TRUE;
}

static GstFlowReturn gstreamer_send_new_sample_handler(GstElement *object, gpointer user_data) {
  GstSample *sample = NULL;
  GstBuffer *buffer = NULL;
  gpointer copy = NULL;
  gsize copy_size = 0;
  SampleHandlerUserData *s = (SampleHandlerUserData *)user_data;

  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample) {
    buffer = gst_sample_get_buffer(sample);
    if (buffer) {
      gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);
      goHandlePipelineBuffer(copy, copy_size, GST_BUFFER_DURATION(buffer), s->pipelineId);
    }
    gst_sample_unref(sample);
  }

  return GST_FLOW_OK;
}

void gstreamer_pipeline_attach_appsink(GstElement *pipeline, char *sinkName, int pipelineId) {
  SampleHandlerUserData *s = calloc(1, sizeof(SampleHandlerUserData));
  s->pipelineId = pipelineId;

  GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), sinkName);
  g_object_set(appsink, "emit-signals", TRUE, NULL);
  g_signal_connect(appsink, "new-sample", G_CALLBACK(gstreamer_send_new_sample_handler), s);
  gst_object_unref(appsink);
}

GstElement *gstreamer_pipeline_create(char *pipelineStr, GError **error) {
  GstElement *pipeline = gst_parse_launch(pipelineStr, error);

  if (pipeline != NULL) {
    GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, gstreamer_bus_call, NULL);
    gst_object_unref(bus);
  }

  return pipeline;
}

void gstreamer_pipeline_play(GstElement *pipeline) {
  gst_element_set_state(pipeline, GST_STATE_PLAYING);
}

void gstreamer_pipeline_stop(GstElement *pipeline) {
  gst_element_set_state(pipeline, GST_STATE_NULL);
  gst_object_unref(pipeline);
}

void gstreamer_pipeline_push(GstElement *pipeline, char *srcName, void *buffer, int bufferLen) {
  GstElement *src = gst_bin_get_by_name(GST_BIN(pipeline), srcName);

  if (src != NULL) {
    gpointer p = g_memdup(buffer, bufferLen);
    GstBuffer *buffer = gst_buffer_new_wrapped(p, bufferLen);
    gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
    gst_object_unref(src);
  }
}
