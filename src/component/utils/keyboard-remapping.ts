const KeyTable = {
  XK_Meta_L: 0xffe7, // Left meta
  XK_Meta_R: 0xffe8, // Right meta
  XK_Control_L: 0xffe3, // Left control
  XK_Control_R: 0xffe4, // Right control
}

export const keySymsRemap = function (key: number) {
  const isMac = navigator && navigator.platform.match(/^mac/i)
  const isiOS = navigator && navigator.platform.match(/ipad|iphone|ipod/i)

  // switch command with ctrl on mac and ios
  if (isMac || isiOS) {
    switch (key) {
      case KeyTable.XK_Meta_L:
        return KeyTable.XK_Control_L
      case KeyTable.XK_Meta_R:
        return KeyTable.XK_Control_R
    }
  }

  return key
}
