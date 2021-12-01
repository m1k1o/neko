#pragma once

#include <gst/gst.h>
#include <gst/app/gstappsrc.h>

extern void goHandlePipelineBuffer(void *buffer, int bufferLen, int samples, int pipelineId);

void gstreamer_pipeline_attach_appsink(GstElement *pipeline, char *sinkName, int pipelineId);
GstElement *gstreamer_pipeline_create(char *pipelineStr, GError **error);
void gstreamer_pipeline_play(GstElement *pipeline);
void gstreamer_pipeline_stop(GstElement *pipeline);
void gstreamer_pipeline_push(GstElement *pipeline, char *srcName, void *buffer, int bufferLen);

void gstreamer_init(void);
