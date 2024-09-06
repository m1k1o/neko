#pragma once

#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xrandr.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>
#include <string.h>

extern void goXEventCursorChanged(XFixesCursorNotifyEvent event);
extern void goXEventClipboardUpdated();
extern void goXEventConfigureNotify(Display *display, Window window, char *name, char *role);
extern void goXEventUnmapNotify(Window window);
extern void goXEventWMChangeState(Display *display, Window window, ulong state);
extern void goXEventError(XErrorEvent *event, char *message);
extern int goXEventActive();

static int XEventError(Display *display, XErrorEvent *event);
void XSetupErrorHandler();
void XEventLoop(char *display);

static void XWindowManagerStateEvent(Display *display, Window window, ulong action, ulong first, ulong second);
void XFileChooserHide(Display *display, Window window);
