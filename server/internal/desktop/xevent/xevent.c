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

  Atom WM_WINDOW_ROLE = XInternAtom(display, "WM_WINDOW_ROLE", 1);
  Atom XA_CLIPBOARD = XInternAtom(display, "CLIPBOARD", 0);
  XFixesSelectSelectionInput(display, root, XA_CLIPBOARD, XFixesSetSelectionOwnerNotifyMask);
  XFixesSelectCursorInput(display, root, XFixesDisplayCursorNotifyMask);
  XSelectInput(display, root, SubstructureNotifyMask);

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

    // XFixesSelectionNotifyEvent
    if (event.type == xfixes_event_base + XFixesSelectionNotify) {
      XFixesSelectionNotifyEvent notifyEvent = *((XFixesSelectionNotifyEvent *) &event);
      if (notifyEvent.subtype == XFixesSetSelectionOwnerNotify && notifyEvent.selection == XA_CLIPBOARD) {
        goXEventClipboardUpdated();
        continue;
      }
    }

    // ConfigureNotify
    if (event.type == ConfigureNotify) {
      Window window = event.xconfigure.window;

      char *name;
      XFetchName(display, window, &name);

      XTextProperty role;
      XGetTextProperty(display, window, &role, WM_WINDOW_ROLE);

      goXEventConfigureNotify(display, window, name, role.value);
      XFree(name);
      continue;
    }

    // UnmapNotify
    if (event.type == UnmapNotify) {
      Window window = event.xunmap.window;
      goXEventUnmapNotify(window);
      continue;
    }
  }

  XCloseDisplay(display);
}
