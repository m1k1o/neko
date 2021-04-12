#include "xorg.h"

static clipboard_c *CLIPBOARD = NULL;
static Display *DISPLAY = NULL;
static char *NAME = ":0.0";
static int REGISTERED = 0;
static int DIRTY = 0;

xkeys_t *xKeysHead = NULL;

void XKeysInsert(KeySym keysym, KeyCode keycode) {
  xkeys_t *node = (xkeys_t *) malloc(sizeof(xkeys_t));

  node->keysym = keysym;
  node->keycode = keycode;
  node->next = xKeysHead;
  xKeysHead = node;
}

KeyCode XKeysPop(KeySym keysym) {
  KeyCode keycode = 0;
  xkeys_t *node = xKeysHead,
          *previous = NULL;
  int i = 0;

  while (node) {
    if (node->keysym == keysym) {
      keycode = node->keycode;

      if (!previous)
        xKeysHead = node->next;
      else
        previous->next = node->next;

      free(node);
      return keycode;
    }

    previous = node;
    node = node->next;
    if (i++ > 120) {
      // this should NEVER HAPPEN
      fprintf(stderr, "[FATAL-ERROR] XKeysPop() in xorg.c: reached maximum loop limit! Something is wrong\n");
      break;
    }
  }

  return 0;
}

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
      fprintf(stderr, "[FATAL-ERROR] XKeysPop() in xorg.c: Could not open main display!");
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

// From: https://github.com/TigerVNC/tigervnc/blob/0946e298075f8f7b6d63e552297a787c5f84d27c/unix/x0vncserver/XDesktop.cxx#L343-L379
KeyCode XkbKeysymToKeycode(Display *dpy, KeySym keysym) {
  XkbDescPtr xkb;
  XkbStateRec state;
  unsigned int mods;
  unsigned keycode;

  xkb = XkbGetMap(dpy, XkbAllComponentsMask, XkbUseCoreKbd);
  if (!xkb)
    return 0;

  XkbGetState(dpy, XkbUseCoreKbd, &state);
  // XkbStateFieldFromRec() doesn't work properly because
  // state.lookup_mods isn't properly updated, so we do this manually
  mods = XkbBuildCoreState(XkbStateMods(&state), state.group);

  for (keycode = xkb->min_key_code;
       keycode <= xkb->max_key_code;
       keycode++) {
    KeySym cursym;
    unsigned int out_mods;
    XkbTranslateKeyCode(xkb, keycode, mods, &out_mods, &cursym);
    if (cursym == keysym)
      break;
  }

  if (keycode > xkb->max_key_code)
    keycode = 0;

  XkbFreeKeyboard(xkb, XkbAllComponentsMask, True);

  // Shift+Tab is usually ISO_Left_Tab, but RFB hides this fact. Do
  // another attempt if we failed the initial lookup
  if ((keycode == 0) && (keysym == XK_Tab) && (mods & ShiftMask))
    return XkbKeysymToKeycode(dpy, XK_ISO_Left_Tab);

  return keycode;
}

void XKey(KeySym key, int down) {
  Display *display = getXDisplay();
  KeyCode code = 0;

  if (!down)
    code = XKeysPop(key);

  if (!code)
    code = XkbKeysymToKeycode(display, key);

  if (!code) {
    int min, max, numcodes;
    XDisplayKeycodes(display, &min, &max);
    XGetKeyboardMapping(display, min, max-min, &numcodes);

    code = (max-min+1)*numcodes;
    KeySym keysym_list[numcodes];
    for (int i=0;i<numcodes;i++) keysym_list[i] = key;
    XChangeKeyboardMapping(display, code, numcodes, keysym_list, 1);
  }

  if (!code)
    return;

  if (down)
    XKeysInsert(key, code);

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

void SetKeyboardModifiers(int num_lock, int caps_lock, int scroll_lock) {
  Display *display = getXDisplay();

  if (num_lock != -1) {
    XkbLockModifiers(display, XkbUseCoreKbd, 16, num_lock * 16);
  }

  if (caps_lock != -1) {
    XkbLockModifiers(display, XkbUseCoreKbd, 2, caps_lock * 2);
  }

  if (scroll_lock != -1) {
    XKeyboardControl values;
    values.led_mode = scroll_lock ? LedModeOn : LedModeOff;
    values.led = 3;
    XChangeKeyboardControl(display, KBLedMode, &values);
  }

  XFlush(display);
}
