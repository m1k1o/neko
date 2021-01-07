#include "gtk.h"

static void drag_data_get(
  GtkWidget *widget,
  GdkDragContext *context,
  GtkSelectionData *data,
  guint target_type,
  guint time,
  gpointer user_data
) {
  gchar **uris = (gchar **) user_data;

  if (target_type == TARGET_TYPE_URI) {
    gtk_selection_data_set_uris(data, uris);
    return;
  }

  if (target_type == TARGET_TYPE_TEXT) {
    gtk_selection_data_set_text(data, uris[0], -1);
    return;
  }
}

static void drag_end(
  GtkWidget *widget,
  GdkDragContext *context,
  gpointer user_data
) {
  gboolean succeeded = gdk_drag_drop_succeeded(context);
  GdkDragAction action = gdk_drag_context_get_selected_action(context);

  char* action_str;
  switch (action) {
    case GDK_ACTION_COPY:
      action_str = "COPY"; break;
    case GDK_ACTION_MOVE:
      action_str = "MOVE"; break;
    case GDK_ACTION_LINK:
      action_str = "LINK"; break;
    case GDK_ACTION_ASK:
      action_str = "ASK"; break;
    default:
      action_str = malloc(sizeof(char) * 20);
      snprintf(action_str, 20, "invalid (%d)", action);
      break;
  }

  fprintf(stderr, "Selected drop action: %s; Succeeded: %d\n", action_str, succeeded);
  if (action_str[0] == 'i') {
    free(action_str);
  }

  gtk_widget_destroy(widget);
}

void drag_window(char **uris) {
  gtk_init(NULL, NULL);

  GtkWidget *widget = gtk_window_new(GTK_WINDOW_TOPLEVEL);
  GtkWindow *window = GTK_WINDOW(widget);

  gtk_window_move(window, 0, 0);
  gtk_window_set_title(window, "neko-drop");
  gtk_window_set_decorated(window, FALSE);
  gtk_window_set_keep_above(window, TRUE);
  gtk_window_set_default_size(window, 100, 100);
  gtk_widget_set_opacity(widget, 0);

  GtkTargetList* target_list = gtk_target_list_new(NULL, 0);
  gtk_target_list_add_uri_targets(target_list, TARGET_TYPE_URI);
  gtk_target_list_add_text_targets(target_list, TARGET_TYPE_TEXT);

  gtk_drag_source_set(widget, GDK_BUTTON1_MASK, NULL, 0, GDK_ACTION_COPY | GDK_ACTION_LINK | GDK_ACTION_ASK);
  gtk_drag_source_set_target_list(widget, target_list);

  g_signal_connect(widget, "drag-data-get", G_CALLBACK(drag_data_get), uris);
  g_signal_connect(widget, "drag-end", G_CALLBACK(drag_end), NULL);
  g_signal_connect(window, "destroy", G_CALLBACK(gtk_main_quit), NULL);

  gtk_widget_show_all(widget);
  gtk_main();
}

char **uris_make(int size) {
  return calloc(size + 1, sizeof(char *));
}

void uris_set_file(char **uris, char *file, int n) {
  GFile *gfile = g_file_new_for_path(file);
  uris[n] = g_file_get_uri(gfile);
}

void uris_free(char **uris, int size) {
  int i;
  for (i = 0; i < size; i++) {
    free(uris[i]);
  }

  free(uris);
}
