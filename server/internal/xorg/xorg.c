#include "xorg.h"

static clipboard_c *CLIPBOARD = NULL;
static Display *DISPLAY = NULL;
static char *NAME = ":0.0";
static int REGISTERED = 0;
static int DIRTY = 0;

typedef struct linked_list {
  KeySym keysym;
  KeyCode keycode;
  struct linked_list *next;
} node;

node *head = NULL;

void insertItem(KeySym keysym, KeyCode keycode) {
  node *temp_node = (node *) malloc(sizeof(node));

  temp_node->keysym = keysym;
  temp_node->keycode = keycode;
  temp_node->next = NULL;

  // For the 1st element
  if (head) {
    temp_node->next = head;
  }
  head = temp_node;
}

void deleteItem(KeySym keysym) {
  node *myNode = head, *previous = NULL;
  int i = 0;

  while (myNode) {
    if (myNode->keysym == keysym) {
      if (!previous)
        head = myNode->next;
      else
        previous->next = myNode->next;

      free(myNode);
      break;
    }

    previous = myNode;
    myNode = myNode->next;
    if (i++ > 120) {
      // this should lead to a panic
      printf("loop over limit");
      break;
    }
  }
}

node *searchItemNode(KeySym keysym) {
  node *searchNode = head;
  int i = 0;

  while (searchNode) {
    if (searchNode->keysym == keysym) {
      return searchNode;
    }

    searchNode = searchNode->next;

    if (i++ > 120) {
      // this should lead to a panic
      printf("loop over limit");
      break;
    }
  }

  return NULL;
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
  node *compareNode;

  // Key is released. Look it up
  if (!down) {
    compareNode = searchItemNode(key);

    // The key is known, use the known KeyCode
    if (compareNode) {
      code = compareNode->keycode;
      XTestFakeKeyEvent(display, code, down, CurrentTime);
      XSync(display, 0);

      deleteItem(key);
      return;
    }
  }

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
    insertItem(key, code);

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
