#pragma once

#include <stdio.h>
#include <gst/gst.h>
#include <gst/app/gstappsrc.h>

typedef struct GstPipelineCtx {
  int pipelineId;
  GstElement *pipeline;
} GstPipelineCtx;

extern void goHandlePipelineBuffer(void *buffer, int bufferLen, int samples, int pipelineId);
extern void goPipelineLog(char *level, char *msg, int pipelineId);

GstPipelineCtx *gstreamer_pipeline_create(char *pipelineStr, int pipelineId, GError **error);
void gstreamer_pipeline_attach_appsink(GstPipelineCtx *ctx, char *sinkName);
void gstreamer_pipeline_play(GstPipelineCtx *ctx);
void gstreamer_pipeline_pause(GstPipelineCtx *ctx);
void gstreamer_pipeline_destory(GstPipelineCtx *ctx);
void gstreamer_pipeline_push(GstPipelineCtx *ctx, char *srcName, void *buffer, int bufferLen);

void gstreamer_init(void);
void gstreamer_loop(void);
