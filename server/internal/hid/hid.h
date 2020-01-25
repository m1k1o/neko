#pragma once

#ifndef XDISPLAY_H
  #define XDISPLAY_H

  #include <X11/Xlib.h>
  #include <X11/extensions/XTest.h>
  #include <stdint.h>
  #include <stdlib.h>
  #include <stdio.h> /* For fputs() */
  #include <string.h> /* For strdup() */

  /* Returns the main display, closed either on exit or when closeMainDisplay()
  * is invoked. This removes a bit of the overhead of calling XOpenDisplay() &
  * XCloseDisplay() everytime the main display needs to be used.
  *
  * Note that this is almost certainly not thread safe. */
  Display *getXDisplay(void);

  void XMove(int x, int y);
  void XScroll(int x, int y);
  void XButton(unsigned int button, int down);
  void XKey(unsigned long key, int down);

  void closeXDisplay(void);
  #ifdef __cplusplus
    extern "C"
    {
  #endif
    void setXDisplay(char *input);
  #ifdef __cplusplus
    }
  #endif
#endif

