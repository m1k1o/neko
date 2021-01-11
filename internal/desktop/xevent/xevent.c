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

  Atom XA_CLIPBOARD;
  XA_CLIPBOARD = XInternAtom(display, "CLIPBOARD", 0);
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

    // XFixesSelectionNotifyEvent
    if (event.type == CreateNotify) {
      char *name;
      XFetchName(display, event.xcreatewindow.window, &name);
 
      char *role;
      XTextProperty text_data;
      Atom atom = XInternAtom(display, "WM_WINDOW_ROLE", True);
      int status = XGetTextProperty(display, event.xcreatewindow.window, &text_data, atom);
      role = (char *)text_data.value;
    
      goXEventWindowCreated(event.xcreatewindow, name, role);
      XFree(name);
      continue;
    }
  }

  XCloseDisplay(display);
}
