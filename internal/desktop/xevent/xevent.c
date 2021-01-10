#include "xevent.h"

static int XEventError(Display *display, XErrorEvent *event) {
  char message[100];

  int error;
  error = XGetErrorText(display, event->error_code, message, sizeof(message));
  if (error) {
    goXEventError(event, "Could not get error message.");
  } else {
    goXEventError(event, message);
  }

  return 1;
}

void XEventLoop(char *name) {
  Display *display = XOpenDisplay(name);
  Window root = RootWindow(display, 0);

  int xfixes_event_base, xfixes_error_base;
  if (!XFixesQueryExtension(display, &xfixes_event_base, &xfixes_error_base)) {
    return;
  }

  XFixesSelectCursorInput(display, root, XFixesDisplayCursorNotifyMask);
  XSync(display, 0);
  XSetErrorHandler(XEventError);

  while (goXEventActive()) {
    XEvent event;
    XNextEvent(display, &event);

    // XFixesDisplayCursorNotify
    if (event.type == xfixes_event_base + 1) {
      XFixesCursorNotifyEvent notifyEvent = *((XFixesCursorNotifyEvent *) &event);
      if (notifyEvent.subtype == XFixesDisplayCursorNotify) {
        goXEventCursorChanged(notifyEvent);
        continue;
      }
    }
  }

  XCloseDisplay(display);
}
