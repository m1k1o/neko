#pragma once

#include <stdio.h>
#include <gst/gst.h>
#include <gst/app/gstappsrc.h>
#include <gst/video/video.h>

#define GLIB_CHECK_VERSION(major,minor,micro)    \
    (GLIB_MAJOR_VERSION > (major) || \
    (GLIB_MAJOR_VERSION == (major) && GLIB_MINOR_VERSION > (minor)) || \
    (GLIB_MAJOR_VERSION == (major) && GLIB_MINOR_VERSION == (minor) && \
      GLIB_MICRO_VERSION >= (micro)))

// g_memdup2 was added in glib 2.67.4, maintain compatibility with older versions
#if !GLIB_CHECK_VERSION(2, 67, 4)
#define g_memdup2 g_memdup
#endif

typedef struct GstPipelineCtx {
  int pipelineId;
  GstElement *pipeline;
  GstElement *appsink;
  GstElement *appsrc;
} GstPipelineCtx;

extern void goHandlePipelineBuffer(int pipelineId, void *buffer, int bufferLen, guint64 duration, gboolean deltaUnit);
extern void goPipelineLog(int pipelineId, char *level, char *msg);

GstPipelineCtx *gstreamer_pipeline_create(char *pipelineStr, int pipelineId, GError **error);
void gstreamer_pipeline_attach_appsink(GstPipelineCtx *ctx, char *sinkName);
void gstreamer_pipeline_attach_appsrc(GstPipelineCtx *ctx, char *srcName);
void gstreamer_pipeline_play(GstPipelineCtx *ctx);
void gstreamer_pipeline_pause(GstPipelineCtx *ctx);
void gstreamer_pipeline_destory(GstPipelineCtx *ctx);
void gstreamer_pipeline_push(GstPipelineCtx *ctx, void *buffer, int bufferLen);

gboolean gstreamer_pipeline_set_prop_int(GstPipelineCtx *ctx, char *binName, char *prop, gint value);
gboolean gstreamer_pipeline_set_caps_framerate(GstPipelineCtx *ctx, const gchar* binName, gint numerator, gint denominator);
gboolean gstreamer_pipeline_set_caps_resolution(GstPipelineCtx *ctx, const gchar* binName, gint width, gint height);
gboolean gstreamer_pipeline_emit_video_keyframe(GstPipelineCtx *ctx);
