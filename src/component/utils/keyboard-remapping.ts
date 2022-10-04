export const KeyTable = {
  XK_v: 0x0076, // U+0076 LATIN SMALL LETTER V

  XK_Control_L: 0xffe3, // Left control
  XK_Control_R: 0xffe4, // Right control

  XK_Meta_L: 0xffe7, // Left meta
  XK_Meta_R: 0xffe8, // Right meta
  XK_Alt_L: 0xffe9, // Left alt
  XK_Alt_R: 0xffea, // Right alt
  XK_Super_L: 0xffeb, // Left super
  XK_Super_R: 0xffec, // Right super

  XK_ISO_Level3_Shift: 0xfe03, // AltGr
}

export const keySymsRemap = function (key: number) {
  const isMac = navigator && navigator.platform.match(/^mac/i)
  const isiOS = navigator && navigator.platform.match(/ipad|iphone|ipod/i)

  // switch command with ctrl and option with altgr on mac and ios
  if (isMac || isiOS) {
    switch (key) {
      case KeyTable.XK_Meta_L: // meta is used by guacamole for CMD key
      case KeyTable.XK_Super_L: // super is used by novnc for CMD key
        return KeyTable.XK_Control_L
      case KeyTable.XK_Meta_R:
      case KeyTable.XK_Super_R:
        return KeyTable.XK_Control_R
      case KeyTable.XK_Alt_L: // alt (option key on mac) behaves like altgr
      case KeyTable.XK_Alt_R:
        return KeyTable.XK_ISO_Level3_Shift
    }
  }

  return key
}
