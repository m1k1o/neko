#include "hid.h"

static clipboard_c *CLIPBOARD = NULL;
static Display *DISPLAY = NULL;
static char *NAME = ":0.0";
static int REGISTERED = 0;
static int DIRTY = 0;

Display *getXDisplay(void) {
  /* Close the display if displayName has changed */
  if (DIRTY) {
    closeXDisplay();
    DIRTY = 0;
  }

  if (DISPLAY == NULL) {
    /* First try the user set displayName */
    DISPLAY = XOpenDisplay(NAME);

    /* Then try using environment variable DISPLAY */
    if (DISPLAY == NULL) {
      DISPLAY = XOpenDisplay(NULL);
    }

    if (DISPLAY == NULL) {
      fputs("Could not open main display\n", stderr);
    } else if (!REGISTERED) {
      atexit(&closeXDisplay);
      REGISTERED = 1;
    }
  }

  return DISPLAY;
}

clipboard_c *getClipboard(void) {
  if (CLIPBOARD == NULL) {
    CLIPBOARD = clipboard_new(NULL);
  }
  return CLIPBOARD;
}

void closeXDisplay(void) {
  if (DISPLAY != NULL) {
    XCloseDisplay(DISPLAY);
    DISPLAY = NULL;
  }
}

void setXDisplay(char *input) {
  NAME = strdup(input);
  DIRTY = 1;
}

void XMove(int x, int y) {
  Display *display = getXDisplay();
  XWarpPointer(display, None, DefaultRootWindow(display),  0, 0, 0, 0, x, y);
  XSync(display, 0);
}

void XScroll(int x, int y) {
  int ydir = 4; /* Button 4 is up, 5 is down. */
  int xdir = 6;

  Display *display = getXDisplay();

  if (y < 0) {
    ydir = 5;
  }

  if (x < 0) {
    xdir = 7;
  }

  int xi;
  int yi;

  for (xi = 0; xi < abs(x); xi++) {
    XTestFakeButtonEvent(display, xdir, 1, CurrentTime);
    XTestFakeButtonEvent(display, xdir, 0, CurrentTime);
  }

  for (yi = 0; yi < abs(y); yi++) {
    XTestFakeButtonEvent(display, ydir, 1, CurrentTime);
    XTestFakeButtonEvent(display, ydir, 0, CurrentTime);
  }

  XSync(display, 0);
}

void XButton(unsigned int button, int down) {
  Display *display = getXDisplay();
  XTestFakeButtonEvent(display, button, down, CurrentTime);
  XSync(display, 0);
}

void XKey(unsigned long key, int down) {
  Display *display = getXDisplay();
  KeyCode code = XKeysymToKeycode(display, key);
  XTestFakeKeyEvent(display, code, down, CurrentTime);
  XSync(display, 0);
}

void XClipboardSet(char *src) {
  clipboard_c *cb = getClipboard();
  clipboard_set_text_ex(cb, src, strlen(src), 0);
}

char *XClipboardGet() {
  clipboard_c *cb = getClipboard();
  return clipboard_text_ex(cb, NULL, 0);
}
