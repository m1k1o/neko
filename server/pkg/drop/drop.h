#pragma once

#include <gtk/gtk.h>

enum {
  DRAG_TARGET_TYPE_TEXT,
  DRAG_TARGET_TYPE_URI
};

extern void goDragCreate(GtkWidget *widget, GdkEvent *event, gpointer user_data);
extern void goDragCursorEnter(GtkWidget *widget, GdkEvent *event, gpointer user_data);
extern void goDragButtonPress(GtkWidget *widget, GdkEvent *event, gpointer user_data);
extern void goDragBegin(GtkWidget *widget, GdkDragContext *context, gpointer user_data);
extern void goDragFinish(gboolean succeeded);

static void dragDataGet(
  GtkWidget *widget,
  GdkDragContext *context,
  GtkSelectionData *data,
  guint target_type,
  guint time,
  gpointer user_data
);

static void dragEnd(
  GtkWidget *widget,
  GdkDragContext *context,
  gpointer user_data
);

void dragWindowOpen(char **uris);
void dragWindowClose();

char **dragUrisMake(int size);
void dragUrisSetFile(char **uris, char *file, int n);
void dragUrisFree(char **uris, int size);
