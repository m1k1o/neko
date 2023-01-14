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

void XCursorPosition(int *x, int *y) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);
  Window window;
  int i;
  unsigned mask;
  XQueryPointer(display, root, &root, &window, x, y, &i, &i, &mask);
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
  if (button == 0)
    return;

  Display *display = getXDisplay();
  XTestFakeButtonEvent(display, button, down, CurrentTime);
  XSync(display, 0);
}

static xkeyentry_t *xKeysHead = NULL;

// add keycode->keysym mapping to list
void XKeyEntryAdd(KeySym keysym, KeyCode keycode) {
  xkeyentry_t *entry = (xkeyentry_t *) malloc(sizeof(xkeyentry_t));
  if (entry == NULL)
    return;

  entry->keysym = keysym;
  entry->keycode = keycode;
  entry->next = xKeysHead;
  xKeysHead = entry;
}

// get keycode for keysym from list
KeyCode XKeyEntryGet(KeySym keysym) {
  xkeyentry_t *prev = NULL;
  xkeyentry_t *curr = xKeysHead;

  KeyCode keycode = 0;
  while (curr != NULL) {
    if (curr->keysym == keysym) {
      keycode = curr->keycode;

      if (prev == NULL) {
        xKeysHead = curr->next;
      } else {
        prev->next = curr->next;
      }

      free(curr);
      return keycode;
    }

    prev = curr;
    curr = curr->next;
  }

  return 0;
}

// From https://github.com/TigerVNC/tigervnc/blob/0946e298075f8f7b6d63e552297a787c5f84d27c/unix/x0vncserver/XDesktop.cxx#L343-L379
KeyCode XkbKeysymToKeycode(Display* dpy, KeySym keysym) {
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

static xkeycode_t *xFreeKeycodesHead = NULL;
int xNumcodes = 0;

void XFreeKeycodesInit(Display* dpy) {
  if (xFreeKeycodesHead != NULL)
    return;

  KeyCode keycode;
  KeySym *keysyms;
  int min, max, numcodes;

  // get full keyboard mapping
  XDisplayKeycodes(dpy, &min, &max);
  keysyms = XGetKeyboardMapping(dpy, min, max-min, &numcodes);
  xNumcodes = numcodes;

  // loop through all keycodes
  xkeycode_t *last = NULL;
  for (int i = min; i <= max; ++i) {
    // check if keycode is empty
    int isEmpty = 1;
    for (int j = 0; j < numcodes; ++j) {
      int symindex = ((i - min) * numcodes) + j;
      if (keysyms[symindex] != 0) {
        isEmpty = 0;
        break;
      }
    }
    if (!isEmpty) continue;

    xkeycode_t *entry = (xkeycode_t *) malloc(sizeof(xkeycode_t));
    if (entry == NULL) return;

    entry->keycode = i;
    if (last == NULL) {
      xFreeKeycodesHead = entry;
    } else {
      last->next = entry;
    } 
    last = entry;
  }

  // no free keycodes, pick last two keycodes anyway
  if (last == NULL) {
    xkeycode_t *entry1 = (xkeycode_t *) malloc(sizeof(xkeycode_t));
    if (entry1 == NULL) return;
    entry1->keycode = max-1;

    xkeycode_t *entry2 = (xkeycode_t *) malloc(sizeof(xkeycode_t));
    if (entry2 == NULL) return;
    entry2->keycode = max-2;

    xFreeKeycodesHead = entry1;
    entry1->next = entry2;
    last = entry2;
  }

  // make as circular list
  last->next = xFreeKeycodesHead;

  XFree(keysyms);
}

// get free keycode from list
KeyCode XFreeKeycodeGet() {
  if (xFreeKeycodesHead == NULL)
    return 0;

  // move head to next entry
  xkeycode_t *entry = xFreeKeycodesHead;
  xFreeKeycodesHead = entry->next;

  return entry->keycode;
}

// map keysym to new keycode
KeyCode XKeysymMapNew(Display* dpy, KeySym keysym) {
  // initialize free keycodes list
  if (xFreeKeycodesHead == NULL) {
    XFreeKeycodesInit(dpy);
  }

  KeyCode keycode = XFreeKeycodeGet();

  // no free keycodes, cannot map keysym
  if (keycode != 0) {
    KeySym keysym_list[xNumcodes];
    for(int i=0;i<xNumcodes;i++) keysym_list[i] = keysym;
    XChangeKeyboardMapping(dpy, keycode, xNumcodes, keysym_list, 1);
  }

  return keycode;
}

void XKey(KeySym keysym, int down) {
  if (keysym == 0)
    return;

  Display *display = getXDisplay();
  KeyCode keycode = 0;

  if (!down)
    keycode = XKeyEntryGet(keysym);

  // Try to get keysyms from existing keycodes
  if (keycode == 0)
    keycode = XkbKeysymToKeycode(display, keysym);

  // Map non-existing keysyms to new keycodes
  if (keycode == 0)
    keycode = XKeysymMapNew(display, keysym);

  if (down)
    XKeyEntryAdd(keysym, keycode);

  XTestFakeKeyEvent(display, keycode, down, CurrentTime);
  XSync(display, 0);
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

char *XGetScreenshot(int *w, int *h) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);

  XWindowAttributes attr;
  XGetWindowAttributes(display, root, &attr);
  int width = attr.width;
  int height = attr.height;

  XImage *ximage = XGetImage(display, root, 0, 0, width, height, AllPlanes, ZPixmap);

  *w = width;
  *h = height;
  char *pixels = (char *)malloc(width * height * 3);

  for (int row = 0; row < height; row++) {
    for (int col = 0; col < width; col++) {
      int pos = ((row * width) + col) * 3;
      unsigned long pixel = XGetPixel(ximage, col, row);

      pixels[pos]   = (pixel & ximage->red_mask)   >> 16;
      pixels[pos+1] = (pixel & ximage->green_mask) >> 8;
      pixels[pos+2] =  pixel & ximage->blue_mask;
    }
  }

  XDestroyImage(ximage);
  return pixels;
}
