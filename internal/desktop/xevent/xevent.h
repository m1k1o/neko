#pragma once

#include <X11/Xlib.h>
#include <X11/extensions/Xfixes.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

extern void goXEventCursorChanged(XFixesCursorNotifyEvent event);
extern void goXEventError(XErrorEvent *event, char *message);
extern int goXEventActive();

static int XEventError(Display *display, XErrorEvent *event);
void XEventLoop(char *display);

