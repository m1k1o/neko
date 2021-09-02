const KeyTable = {
  XK_Control_L: 0xffe3, // Left control
  XK_Control_R: 0xffe4, // Right control

  XK_Meta_L: 0xffe7, // Left meta
  XK_Meta_R: 0xffe8, // Right meta
  XK_Alt_L: 0xffe9, // Left alt
  XK_Alt_R: 0xffea, // Right alt
  XK_Super_L: 0xffeb, // Left super
  XK_Super_R: 0xffec, // Right super

  XK_ISO_Level3_Shift: 0xfe03, // AltGr
  XK_Mode_switch: 0xff7e, // Character set switch
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
      case KeyTable.XK_Super_L:
        return KeyTable.XK_Alt_L
      case KeyTable.XK_Super_R:
        return KeyTable.XK_Super_L
      case KeyTable.XK_Alt_L:
        return KeyTable.XK_Mode_switch
      case KeyTable.XK_Alt_R:
        return KeyTable.XK_ISO_Level3_Shift
    }
  }

  return key
}
