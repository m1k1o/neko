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
  if (button == 0) return;

  Display *display = getXDisplay();
  XTestFakeButtonEvent(display, button, down, CurrentTime);
  XSync(display, 0);
}

void XKey(unsigned long key, int down) {
  if (key == 0) return;

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

void XPutWindowBelow(Window window) {
  Display *display = getXDisplay();
  Window root = RootWindow(display, 0);

  // The WM_TRANSIENT_FOR property is defined by the [ICCCM] for managed windows.
  // This specification extends the use of the property to override-redirect windows.
  // If an override-redirect is a pop-up on behalf of another window, then the Client
  // SHOULD set WM_TRANSIENT_FOR on the override-redirect to this other window.
  //
  // As an example, a Client should set WM_TRANSIENT_FOR on dropdown menus to the
  // toplevel application window that contains the menubar.

  // Remove WM_TRANSIENT_FOR
  Atom WM_TRANSIENT_FOR = XInternAtom(display, "WM_TRANSIENT_FOR", 0);
  XDeleteProperty(display, window, WM_TRANSIENT_FOR);

  // Add _NET_WM_STATE_BELOW
  XClientMessageEvent clientMessageEvent;
  memset(&clientMessageEvent, 0, sizeof(clientMessageEvent));

  // window  = the respective client window
  // message_type = _NET_WM_STATE
  // format = 32
  // data.l[0] = the action, as listed below
  //   _NET_WM_STATE_REMOVE  0 // remove/unset property
  //   _NET_WM_STATE_ADD     1 // add/set property
  //   _NET_WM_STATE_TOGGLE  2 // toggle property 
  // data.l[1] = first property to alter
  // data.l[2] = second property to alter
  // data.l[3] = source indication
  // other data.l[] elements = 0

  clientMessageEvent.type         = ClientMessage;
  clientMessageEvent.window       = window;
  clientMessageEvent.message_type = XInternAtom(display, "_NET_WM_STATE", 0);
  clientMessageEvent.format       = 32;
  clientMessageEvent.data.l[0]    = 1;
  clientMessageEvent.data.l[1]    = XInternAtom(display, "_NET_WM_STATE_BELOW", 0);
  clientMessageEvent.data.l[3]    = 1;

  XSendEvent(display, root, 0, SubstructureRedirectMask | SubstructureNotifyMask, (XEvent *)&clientMessageEvent);
}
