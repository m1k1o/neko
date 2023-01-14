#pragma once

#include <X11/Xlib.h>
#include <X11/XKBlib.h>
#include <X11/Xutil.h>
#include <X11/extensions/Xrandr.h>
#include <X11/extensions/XTest.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>

extern void goCreateScreenSize(int index, int width, int height, int mwidth, int mheight);
extern void goSetScreenRates(int index, int rate_index, short rate);

Display *getXDisplay(void);
int XDisplayOpen(char *input);
void XDisplayClose(void);

void XMove(int x, int y);
void XCursorPosition(int *x, int *y);
void XScroll(int x, int y);
void XButton(unsigned int button, int down);

typedef struct xkeyentry_t {
  KeySym keysym;
  KeyCode keycode;
  struct xkeyentry_t *next;
} xkeyentry_t;

typedef struct xkeycode_t {
  KeyCode keycode;
  struct xkeycode_t *next;
} xkeycode_t;

static void XKeyEntryAdd(KeySym keysym, KeyCode keycode);
static KeyCode XKeyEntryGet(KeySym keysym);
static KeyCode XkbKeysymToKeycode(Display *dpy, KeySym keysym);
void XKey(KeySym keysym, int down);

void XGetScreenConfigurations();
void XSetScreenConfiguration(int index, short rate);
int XGetScreenSize();
short XGetScreenRate();

void XSetKeyboardModifier(int mod, int on);
char XGetKeyboardModifiers();
XFixesCursorImage *XGetCursorImage(void);

char *XGetScreenshot(int *w, int *h);
