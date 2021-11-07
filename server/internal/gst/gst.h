#pragma once

#include <gst/gst.h>
#include <gst/app/gstappsrc.h>

extern void goHandlePipelineBuffer(void *buffer, int bufferLen, int samples, int pipelineId);

GstElement *gstreamer_send_create_pipeline(char *pipeline, GError **error);

void gstreamer_send_start_pipeline(GstElement *pipeline, int pipelineId);
void gstreamer_send_play_pipeline(GstElement *pipeline);
void gstreamer_send_stop_pipeline(GstElement *pipeline);
void gstreamer_init(void);
