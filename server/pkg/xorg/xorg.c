#include "xorg.h"
#include <X11/Xatom.h>
#include <fcntl.h>
#include <sys/file.h>
#include <unistd.h>

static Display *DISPLAY = NULL;
static Window TARGET_WINDOW = None;
static int TARGET_REGION_ENABLED = 0;
static int TARGET_REGION_X = 0;
static int TARGET_REGION_Y = 0;
static int TARGET_REGION_WIDTH = 0;
static int TARGET_REGION_HEIGHT = 0;
static int INPUT_LOCK_FD = -1;
static int INPUT_LEASE_HELD = 0;
static unsigned int INPUT_BUTTONS_DOWN = 0;
static int TARGET_POINTER_VALID = 0;
static int TARGET_POINTER_X = 0;
static int TARGET_POINTER_Y = 0;
static xkeyentry_t *xKeysHead = NULL;
// XTEST virtual keyboard XInput1 device handle — cached so XKey() can dispatch
// via XTestFakeDeviceKeyEvent (XI-aware) instead of XTestFakeKeyEvent (core-only).
// GDK3 selects XI2 for the seat keyboard at startup and ignores core-protocol
// KeyPress events, so core XTest silently drops every key into Firefox.
static XDevice *XTEST_KEYBOARD = NULL;

Display *getXDisplay(void) {
  return DISPLAY;
}

// Discover and open the "Virtual core XTEST keyboard" XInput1 device.
// Must be called after XOpenDisplay. Returns NULL on failure; callers fall back
// to XTestFakeKeyEvent (which doesn't work against modern GDK3, but at least
// doesn't crash).
static XDevice *openXTestKeyboardDevice(Display *display) {
  int ndev = 0;
  XDeviceInfo *devs = XListInputDevices(display, &ndev);
  if (devs == NULL)
    return NULL;

  XDevice *dev = NULL;
  for (int i = 0; i < ndev; i++) {
    if (devs[i].name == NULL)
      continue;
    // Exact match against the XTEST extension's virtual keyboard. X.Org spawns
    // this device at server startup. Use strstr to tolerate naming variations
    // across X servers — e.g. "Virtual core XTEST keyboard" vs "XTEST keyboard".
    if (strstr(devs[i].name, "XTEST") != NULL && strstr(devs[i].name, "keyboard") != NULL) {
      dev = XOpenDevice(display, devs[i].id);
      if (dev != NULL)
        break;
    }
  }

  XFreeDeviceList(devs);
  return dev;
}

int XDisplayOpen(char *name) {
  DISPLAY = XOpenDisplay(name);
  if (DISPLAY == NULL)
    return 1;

  // Best-effort: cache the XTEST keyboard device. If this fails we still return
  // success — XKey falls back to core XTest, which at least preserves the old
  // behaviour (broken, but not a crash).
  XTEST_KEYBOARD = openXTestKeyboardDevice(DISPLAY);
  return 0;
}

void XDisplayClose(void) {
  if (XTEST_KEYBOARD != NULL) {
    XCloseDevice(DISPLAY, XTEST_KEYBOARD);
    XTEST_KEYBOARD = NULL;
  }
  if (INPUT_LOCK_FD >= 0) {
    if (INPUT_LEASE_HELD)
      flock(INPUT_LOCK_FD, LOCK_UN);
    close(INPUT_LOCK_FD);
    INPUT_LOCK_FD = -1;
    INPUT_LEASE_HELD = 0;
  }
  XCloseDisplay(DISPLAY);
}

static int XInputLock(void) {
  if (INPUT_LEASE_HELD)
    return 1;
  if (INPUT_LOCK_FD < 0)
    INPUT_LOCK_FD = open("/tmp/neko-runtime/input.lock", O_CREAT | O_RDWR, 0666);
  if (INPUT_LOCK_FD < 0 || flock(INPUT_LOCK_FD, LOCK_EX) != 0)
    return 0;
  INPUT_LEASE_HELD = 1;
  return 1;
}

static void XInputUnlock(void) {
  if (INPUT_LOCK_FD >= 0 && INPUT_LEASE_HELD) {
    flock(INPUT_LOCK_FD, LOCK_UN);
    INPUT_LEASE_HELD = 0;
  }
}

int XSetTargetWindow(unsigned long window, int *width, int *height) {
  Display *display = getXDisplay();
  XWindowAttributes attributes;
  if (window == 0 || !XGetWindowAttributes(display, (Window) window, &attributes))
    return 1;

  TARGET_WINDOW = (Window) window;
  TARGET_REGION_ENABLED = 0;
  *width = attributes.width;
  *height = attributes.height;
  return 0;
}

void XSetTargetRegion(int x, int y, int width, int height) {
  TARGET_WINDOW = None;
  TARGET_REGION_ENABLED = 1;
  TARGET_REGION_X = x;
  TARGET_REGION_Y = y;
  TARGET_REGION_WIDTH = width;
  TARGET_REGION_HEIGHT = height;
}

static int XWindowMatchesTargetRegion(Display *display, Window window) {
  XWindowAttributes attributes;
  if (!XGetWindowAttributes(display, window, &attributes) || attributes.map_state != IsViewable)
    return 0;

  Window child;
  int root_x = 0;
  int root_y = 0;
  if (!XTranslateCoordinates(display, window, DefaultRootWindow(display), 0, 0, &root_x, &root_y, &child))
    return 0;

  return root_x == TARGET_REGION_X && root_y == TARGET_REGION_Y &&
         attributes.width == TARGET_REGION_WIDTH && attributes.height == TARGET_REGION_HEIGHT;
}

static Window XTargetWindow(Display *display) {
  if (!TARGET_REGION_ENABLED) {
    XWindowAttributes attributes;
    if (TARGET_WINDOW == None ||
        !XGetWindowAttributes(display, TARGET_WINDOW, &attributes) ||
        attributes.map_state != IsViewable)
      return None;
    return TARGET_WINDOW;
  }

  Window root = DefaultRootWindow(display);
  Atom client_list = XInternAtom(display, "_NET_CLIENT_LIST_STACKING", True);
  if (client_list == None)
    client_list = XInternAtom(display, "_NET_CLIENT_LIST", True);
  if (client_list == None)
    return None;

  Atom actual_type;
  int actual_format;
  unsigned long item_count = 0;
  unsigned long bytes_after = 0;
  unsigned char *data = NULL;
  int result = XGetWindowProperty(display, root, client_list, 0, ~0L, False, XA_WINDOW,
                                  &actual_type, &actual_format, &item_count, &bytes_after, &data);
  if (result != Success || actual_type != XA_WINDOW || actual_format != 32 || data == NULL) {
    if (data != NULL)
      XFree(data);
    return None;
  }

  Window target = None;
  Window *windows = (Window *)data;
  for (unsigned long i = item_count; i > 0; i--) {
    if (XWindowMatchesTargetRegion(display, windows[i - 1])) {
      target = windows[i - 1];
      break;
    }
  }
  XFree(data);
  return target;
}

static void XFocusTargetWindow(Display *display, Window target) {
  XRaiseWindow(display, target);
  XSetInputFocus(display, target, RevertToParent, CurrentTime);
  XSync(display, 0);
}

static int XTargetWindowOffset(Display *display, Window target, int *x, int *y) {
  if (TARGET_REGION_ENABLED) {
    *x += TARGET_REGION_X;
    *y += TARGET_REGION_Y;
    return 1;
  }
  if (target == None)
    return 0;

  Window child;
  int root_x = 0;
  int root_y = 0;
  if (!XTranslateCoordinates(display, target, DefaultRootWindow(display), 0, 0, &root_x, &root_y, &child))
    return 0;
  *x += root_x;
  *y += root_y;
  return 1;
}

static int XTargetCoordinatesValid(Display *display, Window target, int x, int y) {
  if (TARGET_REGION_ENABLED)
    return x >= 0 && y >= 0 && x < TARGET_REGION_WIDTH && y < TARGET_REGION_HEIGHT;
  if (target == None)
    return 0;

  XWindowAttributes attributes;
  return XGetWindowAttributes(display, target, &attributes) &&
         x >= 0 && y >= 0 && x < attributes.width && y < attributes.height;
}

static Window XPrepareInputTarget(Display *display) {
  if (!TARGET_REGION_ENABLED && TARGET_WINDOW == None)
    return DefaultRootWindow(display);

  Window target = XTargetWindow(display);
  if (target != None)
    XFocusTargetWindow(display, target);
  return target;
}

void XMove(int x, int y) {
  Display *display = getXDisplay();
  int temporary_lease = !INPUT_LEASE_HELD;
  if (!XInputLock())
    return;
  Window target = XPrepareInputTarget(display);
  if (target == None || !XTargetCoordinatesValid(display, target, x, y) ||
      !XTargetWindowOffset(display, target, &x, &y)) {
    TARGET_POINTER_VALID = 0;
    if (temporary_lease)
      XInputUnlock();
    return;
  }
  TARGET_POINTER_X = x;
  TARGET_POINTER_Y = y;
  TARGET_POINTER_VALID = 1;
  XWarpPointer(display, None, DefaultRootWindow(display), 0, 0, 0, 0, x, y);
  XSync(display, 0);
  if (temporary_lease)
    XInputUnlock();
}

void XCursorPosition(int *x, int *y) {
  Display *display = getXDisplay();
  if ((TARGET_REGION_ENABLED || TARGET_WINDOW != None) && TARGET_POINTER_VALID) {
    *x = TARGET_POINTER_X;
    *y = TARGET_POINTER_Y;
    if (TARGET_REGION_ENABLED) {
      *x -= TARGET_REGION_X;
      *y -= TARGET_REGION_Y;
    } else {
      Window target = XTargetWindow(display);
      int offset_x = 0;
      int offset_y = 0;
      if (target != None && XTargetWindowOffset(display, target, &offset_x, &offset_y)) {
        *x -= offset_x;
        *y -= offset_y;
      }
    }
    return;
  }
  Window root = DefaultRootWindow(display);
  Window window;
  int i;
  unsigned mask;
  XQueryPointer(display, root, &root, &window, x, y, &i, &i, &mask);
  if (TARGET_REGION_ENABLED) {
    *x -= TARGET_REGION_X;
    *y -= TARGET_REGION_Y;
  } else if (TARGET_WINDOW != None) {
    Window target = XTargetWindow(display);
    int offset_x = 0;
    int offset_y = 0;
    if (target != None && XTargetWindowOffset(display, target, &offset_x, &offset_y)) {
      *x -= offset_x;
      *y -= offset_y;
    }
  }
}

void XScroll(int deltaX, int deltaY, int control) {
  Display *display = getXDisplay();
  int temporary_lease = !INPUT_LEASE_HELD;
  if (!XInputLock())
    return;
  Window target = XPrepareInputTarget(display);
  if (target == None || ((TARGET_REGION_ENABLED || TARGET_WINDOW != None) && !TARGET_POINTER_VALID)) {
    if (temporary_lease)
      XInputUnlock();
    return;
  }
  if (TARGET_POINTER_VALID)
    XWarpPointer(display, None, DefaultRootWindow(display), 0, 0, 0, 0, TARGET_POINTER_X, TARGET_POINTER_Y);
  if (control)
    XkbLockModifiers(display, XkbUseCoreKbd, ControlMask, ControlMask);

  int ydir;
  if (deltaY > 0) {
    ydir = 4; // button 4 is up
  } else {
    ydir = 5; // button 5 is down
  }

  int xdir;
  if (deltaX > 0) {
    xdir = 6; // button 6 is right
  } else {
    xdir = 7; // button 7 is left
  }

  for (int i = 0; i < abs(deltaY); i++) {
    XTestFakeButtonEvent(display, ydir, 1, CurrentTime);
    XTestFakeButtonEvent(display, ydir, 0, CurrentTime);
  }

  for (int i = 0; i < abs(deltaX); i++) {
    XTestFakeButtonEvent(display, xdir, 1, CurrentTime);
    XTestFakeButtonEvent(display, xdir, 0, CurrentTime);
  }

  if (control)
    XkbLockModifiers(display, XkbUseCoreKbd, ControlMask, 0);

  XSync(display, 0);
  if (temporary_lease)
    XInputUnlock();
}

void XButton(unsigned int button, int down) {
  if (button == 0)
    return;

  Display *display = getXDisplay();
  int temporary_lease = !INPUT_LEASE_HELD;
  if (!XInputLock())
    return;
  Window target = XPrepareInputTarget(display);
  if (target == None || ((TARGET_REGION_ENABLED || TARGET_WINDOW != None) && !TARGET_POINTER_VALID)) {
    if (!down && button <= sizeof(INPUT_BUTTONS_DOWN) * 8) {
      unsigned int mask = 1U << (button - 1);
      if (INPUT_BUTTONS_DOWN & mask) {
        XSetInputFocus(display, DefaultRootWindow(display), RevertToParent, CurrentTime);
        XTestFakeButtonEvent(display, button, 0, CurrentTime);
        XSync(display, 0);
      }
      INPUT_BUTTONS_DOWN &= ~mask;
    }
    if (INPUT_BUTTONS_DOWN == 0 && xKeysHead == NULL)
      XInputUnlock();
    if (temporary_lease)
      XInputUnlock();
    return;
  }
  if (TARGET_POINTER_VALID)
    XWarpPointer(display, None, DefaultRootWindow(display), 0, 0, 0, 0, TARGET_POINTER_X, TARGET_POINTER_Y);
  XTestFakeButtonEvent(display, button, down, CurrentTime);
  XSync(display, 0);
  if (button <= sizeof(INPUT_BUTTONS_DOWN) * 8) {
    unsigned int mask = 1U << (button - 1);
    if (down)
      INPUT_BUTTONS_DOWN |= mask;
    else
      INPUT_BUTTONS_DOWN &= ~mask;
  }
  if (INPUT_BUTTONS_DOWN == 0 && xKeysHead == NULL)
    XInputUnlock();
}

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

// From https://github.com/TigerVNC/tigervnc/blob/a434ef3377943e89165ac13c537cd0f28be97f84/unix/x0vncserver/XDesktop.cxx#L401-L453
KeyCode XkbAddKeyKeysym(Display* dpy, KeySym keysym) {
  int types[1];
  unsigned int key;
  XkbDescPtr xkb;
  XkbMapChangesRec changes;
  KeySym *syms;
  KeySym upper, lower;

  xkb = XkbGetMap(dpy, XkbAllComponentsMask, XkbUseCoreKbd);

  if (!xkb)
    return 0;

  for (key = xkb->max_key_code; key >= xkb->min_key_code; key--) {
    if (XkbKeyNumGroups(xkb, key) == 0)
      break;
  }

  // no free keycodes
  if (key < xkb->min_key_code)
    return 0;

  // assign empty structure
  changes = *(XkbMapChangesRec *) malloc(sizeof(XkbMapChangesRec));
  for (int i = 0; i < sizeof(changes); i++) ((char *) &changes)[i] = 0;

  XConvertCase(keysym, &lower, &upper);

  if (upper == lower)
    types[XkbGroup1Index] = XkbOneLevelIndex;
  else
    types[XkbGroup1Index] = XkbAlphabeticIndex;

  XkbChangeTypesOfKey(xkb, key, 1, XkbGroup1Mask, types, &changes);

  syms = XkbKeySymsPtr(xkb,key);
  if (upper == lower)
    syms[0] = keysym;
  else {
    syms[0] = lower;
    syms[1] = upper;
  }

  changes.changed |= XkbKeySymsMask;
  changes.first_key_sym = key;
  changes.num_key_syms = 1;

  if (XkbChangeMap(dpy, xkb, &changes)) {
    return key;
  }

  return 0;
}

void XKey(KeySym keysym, int down) {
  if (keysym == 0)
    return;

  Display *display = getXDisplay();
  int temporary_lease = !INPUT_LEASE_HELD;
  if (!XInputLock())
    return;
  Window target = XPrepareInputTarget(display);
  if (target == None) {
    if (!down) {
      KeyCode keycode = XKeyEntryGet(keysym);
      if (keycode != 0) {
        XSetInputFocus(display, DefaultRootWindow(display), RevertToParent, CurrentTime);
        if (XTEST_KEYBOARD != NULL)
          XTestFakeDeviceKeyEvent(display, XTEST_KEYBOARD, keycode, 0, NULL, 0, CurrentTime);
        else
          XTestFakeKeyEvent(display, keycode, 0, CurrentTime);
        XSync(display, 0);
      }
    }
    if (INPUT_BUTTONS_DOWN == 0 && xKeysHead == NULL)
      XInputUnlock();
    if (temporary_lease)
      XInputUnlock();
    return;
  }
  KeyCode keycode = 0;

  if (!down)
    keycode = XKeyEntryGet(keysym);

  // Try to get keysyms from existing keycodes
  if (keycode == 0)
    keycode = XkbKeysymToKeycode(display, keysym);

  // Map non-existing keysyms to new keycodes
  if (keycode == 0)
    keycode = XkbAddKeyKeysym(display, keysym);

  if (keycode == 0) {
    if (temporary_lease)
      XInputUnlock();
    return;
  }

  if (down)
    XKeyEntryAdd(keysym, keycode);

  // Prefer XTestFakeDeviceKeyEvent: it dispatches via the XTEST XInput device,
  // so XI2 listeners (GDK3 inside Firefox) actually receive the event. The core
  // XTestFakeKeyEvent path produces core-protocol events that GDK3's XI2 keyboard
  // path silently drops.
  if (XTEST_KEYBOARD != NULL) {
    XTestFakeDeviceKeyEvent(display, XTEST_KEYBOARD, keycode, down, NULL, 0, CurrentTime);
  } else {
    XTestFakeKeyEvent(display, keycode, down, CurrentTime);
  }
  XSync(display, 0);
  if (INPUT_BUTTONS_DOWN == 0 && xKeysHead == NULL)
    XInputUnlock();
}

Status XSetScreenConfiguration(int width, int height, short rate) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);
  XRRScreenConfiguration *conf = XRRGetScreenInfo(display, root);

  XRRScreenSize *xrrs;
  int num_sizes;
  xrrs = XRRConfigSizes(conf, &num_sizes);

  int size_index = -1;
  for (int i = 0; i < num_sizes; i++) {
    if (xrrs[i].width == width && xrrs[i].height == height) {
      size_index = i;
      break;
    }
  }

  // if we cannot find the size
  if (size_index == -1) {
    return RRSetConfigFailed;
  }

  Status status;
  status = XRRSetScreenConfigAndRate(display, conf, root, size_index, RR_Rotate_0, rate, CurrentTime);

  XRRFreeScreenConfigInfo(conf);
  return status;
}

void XGetScreenConfiguration(int *width, int *height, short *rate) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);
  XRRScreenConfiguration *conf = XRRGetScreenInfo(display, root);

  Rotation current_rotation;
  SizeID current_size_id = XRRConfigCurrentConfiguration(conf, &current_rotation);

  XRRScreenSize *xrrs;
  int num_sizes;
  xrrs = XRRConfigSizes(conf, &num_sizes);

  // if we cannot find the size
  if (current_size_id >= num_sizes) {
    return;
  }

  *width = xrrs[current_size_id].width;
  *height = xrrs[current_size_id].height;
  *rate = XRRConfigCurrentRate(conf);

  XRRFreeScreenConfigInfo(conf);
}

void XGetScreenConfigurations() {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);
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

// Inspired by https://github.com/raboof/xrandr/blob/master/xrandr.c
void XCreateScreenMode(int width, int height, short rate) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);

  // create new mode info
  XRRModeInfo *mode_info = XCreateScreenModeInfo(width, height, rate);

  // create new mode
  RRMode mode = XRRCreateMode(display, root, mode_info);
  XSync(display, 0);

  // add new mode to all outputs
	XRRScreenResources *resources = XRRGetScreenResources(display, root);
  for (int i = 0; i < resources->noutput; ++i) {
    XRRAddOutputMode(display, resources->outputs[i], mode);
  }

  XRRFreeScreenResources(resources);
  XRRFreeModeInfo(mode_info);
}

// Inspired by https://fossies.org/linux/xwayland/hw/xwayland/xwayland-cvt.c
XRRModeInfo *XCreateScreenModeInfo(int hdisplay, int vdisplay, short vrefresh) {
  char name[128];
  snprintf(name, sizeof name, "%dx%d_%d", hdisplay, vdisplay, vrefresh);
  XRRModeInfo *modeinfo = XRRAllocModeInfo(name, strlen(name));

#ifdef _LIBCVT_H_
  struct libxcvt_mode_info *mode_info;

  // get screen mode from libxcvt, if available
  mode_info = libxcvt_gen_mode_info(hdisplay, vdisplay, vrefresh, false, false);

  modeinfo->width      = mode_info->hdisplay;
  modeinfo->height     = mode_info->vdisplay;
  modeinfo->dotClock   = mode_info->dot_clock * 1000;
  modeinfo->hSyncStart = mode_info->hsync_start;
  modeinfo->hSyncEnd   = mode_info->hsync_end;
  modeinfo->hTotal     = mode_info->htotal;
  modeinfo->vSyncStart = mode_info->vsync_start;
  modeinfo->vSyncEnd   = mode_info->vsync_end;
  modeinfo->vTotal     = mode_info->vtotal;
  modeinfo->modeFlags  = mode_info->mode_flags;

  free(mode_info);
#else
  // fallback to a simple mode without refresh rate
  modeinfo->width = hdisplay;
  modeinfo->height = vdisplay;
#endif

  return modeinfo;
}

void XSetKeyboardModifier(unsigned char mod, int on) {
  Display *display = getXDisplay();
  int temporary_lease = !INPUT_LEASE_HELD;
  if (!XInputLock())
    return;
  Window target = XPrepareInputTarget(display);
  if (target == None) {
    if (temporary_lease)
      XInputUnlock();
    return;
  }
  XkbLockModifiers(display, XkbUseCoreKbd, mod, on ? mod : 0);
  XFlush(display);
  if (temporary_lease)
    XInputUnlock();
}

unsigned char XGetKeyboardModifiers() {
  XkbStateRec xkbState;
  Display *display = getXDisplay();
  XkbGetState(display, XkbUseCoreKbd, &xkbState);
  // XkbStateFieldFromRec() doesn't work properly because
  // state.lookup_mods isn't properly updated, so we do this manually
  return XkbBuildCoreState(XkbStateMods(&xkbState), xkbState.group);
}

XFixesCursorImage *XGetCursorImage(void) {
  Display *display = getXDisplay();
  return XFixesGetCursorImage(display);
}

char *XGetScreenshot(int *w, int *h) {
  Display *display = getXDisplay();
  Window root = DefaultRootWindow(display);

  int x = 0;
  int y = 0;
  int width = 0;
  int height = 0;
  XWindowAttributes root_attr;
  if (!XGetWindowAttributes(display, root, &root_attr)) {
    *w = 0;
    *h = 0;
    return NULL;
  }

  if (TARGET_REGION_ENABLED) {
    x = TARGET_REGION_X;
    y = TARGET_REGION_Y;
    width = TARGET_REGION_WIDTH;
    height = TARGET_REGION_HEIGHT;
  } else if (TARGET_WINDOW != None) {
    Window target = XTargetWindow(display);
    XWindowAttributes target_attr;
    Window child;
    if (target == None || !XGetWindowAttributes(display, target, &target_attr) ||
        !XTranslateCoordinates(display, target, root, 0, 0, &x, &y, &child)) {
      *w = 0;
      *h = 0;
      return NULL;
    }
    width = target_attr.width;
    height = target_attr.height;
  } else {
    width = root_attr.width;
    height = root_attr.height;
  }

  if (x < 0 || y < 0 || width <= 0 || height <= 0 ||
      x + width > root_attr.width || y + height > root_attr.height) {
    *w = 0;
    *h = 0;
    return NULL;
  }

  XImage *ximage = XGetImage(display, root, x, y, width, height, AllPlanes, ZPixmap);
  if (ximage == NULL) {
    *w = 0;
    *h = 0;
    return NULL;
  }

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
