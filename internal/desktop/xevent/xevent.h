#pragma once

#include <X11/Xatom.h>
#include <X11/Xlib.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>

extern void goXEventCursorChanged(XFixesCursorNotifyEvent event);
extern void goXEventClipboardUpdated();
extern void goXEventError(XErrorEvent *event, char *message);
extern int goXEventActive();

static int XEventError(Display *display, XErrorEvent *event);
void XEventLoop(char *display);

