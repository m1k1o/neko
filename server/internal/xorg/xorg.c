#include "xorg.h"

static clipboard_c *CLIPBOARD = NULL;
static Display *DISPLAY = NULL;
static char *NAME = ":0.0";
static int REGISTERED = 0;
static int DIRTY = 0;

Display *getXDisplay(void) {
  /* Close the display if displayName has changed */
  if (DIRTY) {
    XDisplayClose();
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
      atexit(&XDisplayClose);
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

void XDisplayClose(void) {
  if (DISPLAY != NULL) {
    XCloseDisplay(DISPLAY);
    DISPLAY = NULL;
  }
}

void XDisplaySet(char *input) {
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
    if(code == 0) {
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

void XClipboardSet(char *src) {
  clipboard_c *cb = getClipboard();
  clipboard_set_text_ex(cb, src, strlen(src), 0);
}

char *XClipboardGet() {
  clipboard_c *cb = getClipboard();
  return clipboard_text_ex(cb, NULL, 0);
}

void XGetScreenConfigurations() {
  Display       *display = getXDisplay();
  Window        root = RootWindow(display, 0);
  XRRScreenSize *xrrs;
  int           num_sizes;

  xrrs = XRRSizes(display, 0, &num_sizes);
  for(int i = 0; i < num_sizes; i ++) {
    short *rates;
    int   num_rates;

    goCreateScreenSize(i, xrrs[i].width, xrrs[i].height, xrrs[i].mwidth, xrrs[i].mheight);
    rates = XRRRates(display, 0, i, &num_rates);
    for (int j = 0; j < num_rates; j ++) {
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
  XRRScreenConfiguration *conf  = XRRGetScreenInfo(display, RootWindow(display, 0));
  Rotation original_rotation;
  return XRRConfigCurrentConfiguration(conf, &original_rotation);
}

short XGetScreenRate() {
  Display *display = getXDisplay();
  XRRScreenConfiguration *conf  = XRRGetScreenInfo(display, RootWindow(display, 0));
  return XRRConfigCurrentRate(conf);
}

void SetKeyboard(char *layout) {
  // TOOD: refactor, use native API.
  char cmd[13] = "setxkbmap ";
  strncat(cmd, layout, 2);
  system(cmd);
}

void SetKeyboardModifiers(int num_lock, int caps_lock, int scroll_lock) {
  // TOOD: refactor, use native API.
  // https://stackoverflow.com/questions/8427817/how-to-get-a-num-lock-state-using-c-c/8429021
  Display *display = getXDisplay();
  XKeyboardState x;
  XGetKeyboardControl(display, &x);

  // set caps lock
  //printf("CapsLock is %s\n", (x.led_mask & 1) ? "On" : "Off");
  if(caps_lock != -1 && x.led_mask & 1 != caps_lock) {
    XKey(0xffe5, 1);
    XKey(0xffe5, 0);
  }

  // set num lock
  //printf("NumLock is %s\n", (x.led_mask & 2) ? "On" : "Off");
  if(num_lock != -1 && x.led_mask & 2 != num_lock) {
    XKey(0xff7f, 1);
    XKey(0xff7f, 0);
  }

  /* NOT SUPPORTED
  // set scroll lock
  //printf("ScrollLock is %s\n", (x.led_mask & 4) ? "On" : "Off");
  if(scroll_lock != -1 && x.led_mask & 4 != scroll_lock) {
    XKey(0xff14, 1);
    XKey(0xff14, 0);
  }
  */
}

