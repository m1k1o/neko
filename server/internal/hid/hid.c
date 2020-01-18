#include "hid.h"

static Display *display = NULL;
static char *name = ":0.0";
static int registered = 0;
static int dirty = 0;

Display *getXDisplay(void) {
  /* Close the display if displayName has changed */
  if (dirty) {
    closeXDisplay();
    dirty = 0;
  }

  if (display == NULL) {
    /* First try the user set displayName */
    display = XOpenDisplay(name);

    /* Then try using environment variable DISPLAY */
    if (display == NULL) {
      display = XOpenDisplay(NULL);
    }

    if (display == NULL) {
      fputs("Could not open main display\n", stderr);
    } else if (!registered) {
      atexit(&closeXDisplay);
      registered = 1;
    }
  }

  return display;
}

void closeXDisplay(void) {
  if (display != NULL) {
    XCloseDisplay(display);
    display = NULL;
  }
}

void setXDisplay(char *input) {
  name = strdup(input);
  dirty = 1;
}

void mouseMove(int x, int y) {
  Display *display = getXDisplay();
  XWarpPointer(display, None, DefaultRootWindow(display),  0, 0, 0, 0, x, y);
  XSync(display, 0);
}

void mouseScroll(int x, int y) {
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

void mouseEvent(unsigned int button, int down) {
  Display *display = getXDisplay();
  XTestFakeButtonEvent(display, button, down, CurrentTime);
  XSync(display, 0);
}

void keyEvent(unsigned long key, int down) {
  Display *display = getXDisplay();
  KeyCode code = XKeysymToKeycode(display, key);
  XTestFakeKeyEvent(display, code, down, CurrentTime);
  XSync(display, 0);
}
