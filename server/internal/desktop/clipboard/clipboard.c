#include "clipboard.h"

static clipboard_c *CLIPBOARD = NULL;

clipboard_c *getClipboard(void) {
  if (CLIPBOARD == NULL) {
    CLIPBOARD = clipboard_new(NULL);
  }

  return CLIPBOARD;
}

void ClipboardSet(char *src) {
  clipboard_c *cb = getClipboard();
  clipboard_set_text_ex(cb, src, strlen(src), 0);
}

char *ClipboardGet() {
  clipboard_c *cb = getClipboard();
  return clipboard_text_ex(cb, NULL, 0);
}
