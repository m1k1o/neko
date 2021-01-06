#pragma once

#include <gtk/gtk.h>

enum {
  TARGET_TYPE_TEXT,
  TARGET_TYPE_URI
};

static void drag_data_get(
    GtkWidget *widget,
    GdkDragContext *context,
    GtkSelectionData *data,
    guint target_type,
    guint time,
    gpointer user_data
);

static void drag_end(
    GtkWidget *widget,
    GdkDragContext *context,
    gpointer user_data
);

void drag_window(char **uris);

char **uris_make(int size);
void uris_set(char **uris, char *filename, int n);
void uris_free(char **uris, int size);
