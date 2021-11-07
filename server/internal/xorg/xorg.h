#pragma once

#include <X11/XKBlib.h>
#include <X11/extensions/Xrandr.h>
#include <X11/extensions/XTest.h>
#include <libclipboard.h>
#include <stdlib.h> /* For free() */
#include <stdio.h> /* For fputs() */
#include <string.h> /* For strdup() */

extern void goCreateScreenSize(int index, int width, int height, int mwidth, int mheight);
extern void goSetScreenRates(int index, int rate_index, short rate);

typedef struct xkeys_t {
  KeySym keysym;
  KeyCode keycode;
  struct xkeys_t *next;
} xkeys_t;

/* Returns the main display, closed either on exit or when closeMainDisplay()
* is invoked. This removes a bit of the overhead of calling XOpenDisplay() &
* XCloseDisplay() everytime the main display needs to be used.
*
* Note that this is almost certainly not thread safe. */
Display *getXDisplay(void);
clipboard_c *getClipboard(void);

void XMove(int x, int y);
void XScroll(int x, int y);
void XButton(unsigned int button, int down);
void XKey(unsigned long key, int down);

void XClipboardSet(char *src);
char *XClipboardGet();

void XGetScreenConfigurations();
void XSetScreenConfiguration(int index, short rate);
int XGetScreenSize();
short XGetScreenRate();

void XDisplayClose(void);
void XDisplaySet(char *input);

void SetKeyboardModifiers(int num_lock, int caps_lock, int scroll_lock);
