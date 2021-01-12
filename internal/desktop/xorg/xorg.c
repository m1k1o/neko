#include "xorg.h"

static Display *DISPLAY = NULL;

Display *getXDisplay(void) {
  return DISPLAY;
}

int XDisplayOpen(char *name) {
  DISPLAY = XOpenDisplay(name);
  return DISPLAY == NULL;
}

void XDisplayClose(void) {
  XCloseDisplay(DISPLAY);
}

void XMove(int x, int y) {
  Display *display = getXDisplay();
  XWarpPointer(display, None, DefaultRootWindow(display), 0, 0, 0, 0, x, y);
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
  if (button != 0) {
    Display *display = getXDisplay();
    XTestFakeButtonEvent(display, button, down, CurrentTime);
    XSync(display, 0);
  }
}

void XKey(unsigned long key, int down) {
  if (key != 0) {
    Display *display = getXDisplay();
    KeyCode code = XKeysymToKeycode(display, key);

    // Map non-existing keysyms to new keycodes
    if (code == 0) {
      int min, max, numcodes;
      XDisplayKeycodes(display, &min, &max);
      XGetKeyboardMapping(display, min, max-min, &numcodes);

      code = (max-min+1)*numcodes;
      KeySym keysym_list[numcodes];
      for(int i=0;i<numcodes;i++) keysym_list[i] = key;
      XChangeKeyboardMapping(display, code, numcodes, keysym_list, 1);
    }

    XTestFakeKeyEvent(display, code, down, CurrentTime);
    XSync(display, 0);
  }
}

void XGetScreenConfigurations() {
  Display *display = getXDisplay();
  Window root = RootWindow(display, 0);
  XRRScreenSize *xrrs;
  int num_sizes;

  xrrs = XRRSizes(display, 0, &num_sizes);
  for (int i = 0; i < num_sizes; i++) {
    short *rates;
    int num_rates;

    goCreateScreenSize(i, xrrs[i].width, xrrs[i].height, xrrs[i].mwidth, xrrs[i].mheight);
    rates = XRRRates(display, 0, i, &num_rates);
    for (int j = 0; j < num_rates; j++) {
      goSetScreenRates(i, j, rates[j]);
    }
  }
}

void XSetScreenConfiguration(int index, short rate) {
  Display *display = getXDisplay();
  Window root = RootWindow(display, 0);
  XRRSetScreenConfigAndRate(display, XRRGetScreenInfo(display, root), root, index, RR_Rotate_0, rate, CurrentTime);
}

int XGetScreenSize() {
  Display *display = getXDisplay();
  XRRScreenConfiguration *conf = XRRGetScreenInfo(display, RootWindow(display, 0));
  Rotation original_rotation;
  return XRRConfigCurrentConfiguration(conf, &original_rotation);
}

short XGetScreenRate() {
  Display *display = getXDisplay();
  XRRScreenConfiguration *conf = XRRGetScreenInfo(display, RootWindow(display, 0));
  return XRRConfigCurrentRate(conf);
}

void SetKeyboardLayout(char *layout) {
  // TOOD: refactor, use native API.
  char cmd[13] = "setxkbmap ";
  strncat(cmd, layout, 2);
  system(cmd);
}

void XSetKeyboardModifier(int mod, int on) {
  Display *display = getXDisplay();
  XkbLockModifiers(display, XkbUseCoreKbd, mod, on ? mod : 0);
  XFlush(display);
}

char XGetKeyboardModifiers() {
  XkbStateRec xkbState;
  Display *display = getXDisplay();
  XkbGetState(display, XkbUseCoreKbd, &xkbState);
  return xkbState.locked_mods;
}

XFixesCursorImage *XGetCursorImage(void) {
  Display *display = getXDisplay();
  return XFixesGetCursorImage(display);
}
