#include "gst.h"

typedef struct SampleHandlerUserData {
  int pipelineId;
} SampleHandlerUserData;

void gstreamer_init(void) {
  gst_init(NULL, NULL);
}

static gboolean gstreamer_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
  switch (GST_MESSAGE_TYPE(msg)) {

  case GST_MESSAGE_EOS:
    g_print("End of stream\n");
    exit(1);
    break;

  case GST_MESSAGE_ERROR: {
    gchar *debug;
    GError *error;

    gst_message_parse_error(msg, &error, &debug);
    g_free(debug);

    g_printerr("Error: %s\n", error->message);
    g_error_free(error);
    exit(1);
  }
  default:
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

  GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
  gst_bus_add_watch(bus, gstreamer_bus_call, NULL);
  gst_object_unref(bus);

  return pipeline;
}

void gstreamer_pipeline_play(GstElement *pipeline) {
  gst_element_set_state(pipeline, GST_STATE_PLAYING);
}

void gstreamer_pipeline_stop(GstElement *pipeline) {
  gst_element_set_state(pipeline, GST_STATE_NULL);
  gst_object_unref(pipeline);
}

void gstreamer_pipeline_push(GstElement *pipeline, char *sinkName, void *buffer, int bufferLen) {
  GstElement *src = gst_bin_get_by_name(GST_BIN(pipeline), sinkName);
  if (src != NULL) {
    gpointer p = g_memdup(buffer, bufferLen);
    GstBuffer *buffer = gst_buffer_new_wrapped(p, bufferLen);
    gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
    gst_object_unref(src);
  }
}
