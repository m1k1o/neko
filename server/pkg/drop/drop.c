#include "drop.h"

GtkWidget *drag_widget = NULL;

static void dragDataGet(
  GtkWidget *widget,
  GdkDragContext *context,
  GtkSelectionData *data,
  guint target_type,
  guint time,
  gpointer user_data
) {
  gchar **uris = (gchar **) user_data;

  if (target_type == DRAG_TARGET_TYPE_URI) {
    gtk_selection_data_set_uris(data, uris);
    return;
  }

  if (target_type == DRAG_TARGET_TYPE_TEXT) {
    gtk_selection_data_set_text(data, uris[0], -1);
    return;
  }
}

static void dragEnd(
  GtkWidget *widget,
  GdkDragContext *context,
  gpointer user_data
) {
  gboolean succeeded = gdk_drag_drop_succeeded(context);
  gtk_widget_destroy(widget);
  goDragFinish(succeeded);
  drag_widget = NULL;
}

void dragWindowOpen(char **uris) {
  if (drag_widget != NULL) dragWindowClose();

  gtk_init(NULL, NULL);

  GtkWidget *widget = gtk_window_new(GTK_WINDOW_POPUP);
  GtkWindow *window = GTK_WINDOW(widget);

  gtk_window_move(window, 0, 0);
  gtk_window_set_title(window, "Neko Drag & Drop Window");
  gtk_window_set_decorated(window, FALSE);
  gtk_window_set_keep_above(window, TRUE);
  gtk_window_set_default_size(window, 100, 100);

  GtkTargetList* target_list = gtk_target_list_new(NULL, 0);
  gtk_target_list_add_uri_targets(target_list, DRAG_TARGET_TYPE_URI);
  gtk_target_list_add_text_targets(target_list, DRAG_TARGET_TYPE_TEXT);

  gtk_drag_source_set(widget, GDK_BUTTON1_MASK, NULL, 0, GDK_ACTION_COPY | GDK_ACTION_LINK | GDK_ACTION_ASK);
  gtk_drag_source_set_target_list(widget, target_list);

  g_signal_connect(widget, "map-event", G_CALLBACK(goDragCreate), NULL);
  g_signal_connect(widget, "enter-notify-event", G_CALLBACK(goDragCursorEnter), NULL);
  g_signal_connect(widget, "button-press-event", G_CALLBACK(goDragButtonPress), NULL);
  g_signal_connect(widget, "drag-begin", G_CALLBACK(goDragBegin), NULL);

  g_signal_connect(widget, "drag-data-get", G_CALLBACK(dragDataGet), uris);
  g_signal_connect(widget, "drag-end", G_CALLBACK(dragEnd), NULL);
  g_signal_connect(window, "destroy", G_CALLBACK(gtk_main_quit), NULL);

  gtk_widget_show_all(widget);
  drag_widget = widget;

  gtk_main();
}

void dragWindowClose() {
  gtk_widget_destroy(drag_widget);
  drag_widget = NULL;
}

char **dragUrisMake(int size) {
  return calloc(size + 1, sizeof(char *));
}

void dragUrisSetFile(char **uris, char *file, int n) {
  GFile *gfile = g_file_new_for_path(file);
  uris[n] = g_file_get_uri(gfile);
}

void dragUrisFree(char **uris, int size) {
  for (int i = 0; i < size; i++) {
    free(uris[i]);
  }

  free(uris);
}
