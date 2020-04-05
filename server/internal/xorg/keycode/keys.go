package keycode

import "n.eko.moe/neko/internal/types"

var BACKSPACE = types.Key{
	Name:   "BACKSPACE",
	Value:  "BackSpace",
	Code:   8,
	Keysym: int(0xff08),
}

var TAB = types.Key{
	Name:   "TAB",
	Value:  "Tab",
	Code:   9,
	Keysym: int(0xFF09),
}

var CLEAR = types.Key{
	Name:   "CLEAR",
	Value:  "Clear",
	Code:   12,
	Keysym: int(0xFF0B),
}

var ENTER = types.Key{
	Name:   "ENTER",
	Value:  "Enter",
	Code:   13,
	Keysym: int(0xFF0D),
}

var SHIFT = types.Key{
	Name:   "SHIFT",
	Value:  "Shift",
	Code:   16,
	Keysym: int(0xFFE1),
}

var CTRL = types.Key{
	Name:   "CTRL",
	Value:  "Ctrl",
	Code:   17,
	Keysym: int(0xFFE3),
}

var ALT = types.Key{
	Name:   "ALT",
	Value:  "Alt",
	Code:   18,
	Keysym: int(0xFFE9),
}

var PAUSE = types.Key{
	Name:   "PAUSE",
	Value:  "Pause",
	Code:   19,
	Keysym: int(0xFF13),
}

var CAPS_LOCK = types.Key{
	Name:   "CAPS_LOCK",
	Value:  "Caps Lock",
	Code:   20,
	Keysym: int(0xFFE5),
}

var ESCAPE = types.Key{
	Name:   "ESCAPE",
	Value:  "Escape",
	Code:   27,
	Keysym: int(0xFF1B),
}

var SPACE = types.Key{
	Name:   "SPACE",
	Value:  " ",
	Code:   32,
	Keysym: int(0x0020),
}

var PAGE_UP = types.Key{
	Name:   "PAGE_UP",
	Value:  "Page Up",
	Code:   33,
	Keysym: int(0xFF55),
}

var PAGE_DOWN = types.Key{
	Name:   "PAGE_DOWN",
	Value:  "Page Down",
	Code:   34,
	Keysym: int(0xFF56),
}

var END = types.Key{
	Name:   "END",
	Value:  "End",
	Code:   35,
	Keysym: int(0xFF57),
}

var HOME = types.Key{
	Name:   "HOME",
	Value:  "Home",
	Code:   36,
	Keysym: int(0xFF50),
}

var LEFT_ARROW = types.Key{
	Name:   "LEFT_ARROW",
	Value:  "Left Arrow",
	Code:   37,
	Keysym: int(0xFF51),
}

var UP_ARROW = types.Key{
	Name:   "UP_ARROW",
	Value:  "Up Arrow",
	Code:   38,
	Keysym: int(0xFF52),
}

var RIGHT_ARROW = types.Key{
	Name:   "RIGHT_ARROW",
	Value:  "Right Arrow",
	Code:   39,
	Keysym: int(0xFF53),
}

var DOWN_ARROW = types.Key{
	Name:   "DOWN_ARROW",
	Value:  "Down Arrow",
	Code:   40,
	Keysym: int(0xFF54),
}

var INSERT = types.Key{
	Name:   "INSERT",
	Value:  "Insert",
	Code:   45,
	Keysym: int(0xFF63),
}

var DELETE = types.Key{
	Name:   "DELETE",
	Value:  "Delete",
	Code:   46,
	Keysym: int(0xFFFF),
}

var KEY_0 = types.Key{
	Name:   "KEY_0",
	Value:  "0",
	Code:   48,
	Keysym: int(0x0030),
}

var KEY_1 = types.Key{
	Name:   "KEY_1",
	Value:  "1",
	Code:   49,
	Keysym: int(0x0031),
}

var KEY_2 = types.Key{
	Name:   "KEY_2",
	Value:  "2",
	Code:   50,
	Keysym: int(0x0032),
}

var KEY_3 = types.Key{
	Name:   "KEY_3",
	Value:  "3",
	Code:   51,
	Keysym: int(0x0033),
}

var KEY_4 = types.Key{
	Name:   "KEY_4",
	Value:  "4",
	Code:   52,
	Keysym: int(0x0034),
}

var KEY_5 = types.Key{
	Name:   "KEY_5",
	Value:  "5",
	Code:   53,
	Keysym: int(0x0035),
}

var KEY_6 = types.Key{
	Name:   "KEY_6",
	Value:  "6",
	Code:   54,
	Keysym: int(0x0036),
}

var KEY_7 = types.Key{
	Name:   "KEY_7",
	Value:  "7",
	Code:   55,
	Keysym: int(0x0037),
}

var KEY_8 = types.Key{
	Name:   "KEY_8",
	Value:  "8",
	Code:   56,
	Keysym: int(0x0038),
}

var KEY_9 = types.Key{
	Name:   "KEY_9",
	Value:  "9",
	Code:   57,
	Keysym: int(0x0039),
}

var KEY_A = types.Key{
	Name:   "KEY_A",
	Value:  "a",
	Code:   65,
	Keysym: int(0x0061),
}

var KEY_B = types.Key{
	Name:   "KEY_B",
	Value:  "b",
	Code:   66,
	Keysym: int(0x0062),
}

var KEY_C = types.Key{
	Name:   "KEY_C",
	Value:  "c",
	Code:   67,
	Keysym: int(0x0063),
}

var KEY_D = types.Key{
	Name:   "KEY_D",
	Value:  "d",
	Code:   68,
	Keysym: int(0x0064),
}

var KEY_E = types.Key{
	Name:   "KEY_E",
	Value:  "e",
	Code:   69,
	Keysym: int(0x0065),
}

var KEY_F = types.Key{
	Name:   "KEY_F",
	Value:  "f",
	Code:   70,
	Keysym: int(0x0066),
}

var KEY_G = types.Key{
	Name:   "KEY_G",
	Value:  "g",
	Code:   71,
	Keysym: int(0x0067),
}

var KEY_H = types.Key{
	Name:   "KEY_H",
	Value:  "h",
	Code:   72,
	Keysym: int(0x0068),
}

var KEY_I = types.Key{
	Name:   "KEY_I",
	Value:  "i",
	Code:   73,
	Keysym: int(0x0069),
}

var KEY_J = types.Key{
	Name:   "KEY_J",
	Value:  "j",
	Code:   74,
	Keysym: int(0x006a),
}

var KEY_K = types.Key{
	Name:   "KEY_K",
	Value:  "k",
	Code:   75,
	Keysym: int(0x006b),
}

var KEY_L = types.Key{
	Name:   "KEY_L",
	Value:  "l",
	Code:   76,
	Keysym: int(0x006c),
}

var KEY_M = types.Key{
	Name:   "KEY_M",
	Value:  "m",
	Code:   77,
	Keysym: int(0x006d),
}

var KEY_N = types.Key{
	Name:   "KEY_N",
	Value:  "n",
	Code:   78,
	Keysym: int(0x006e),
}

var KEY_O = types.Key{
	Name:   "KEY_O",
	Value:  "o",
	Code:   79,
	Keysym: int(0x006f),
}

var KEY_P = types.Key{
	Name:   "KEY_P",
	Value:  "p",
	Code:   80,
	Keysym: int(0x0070),
}

var KEY_Q = types.Key{
	Name:   "KEY_Q",
	Value:  "q",
	Code:   81,
	Keysym: int(0x0071),
}

var KEY_R = types.Key{
	Name:   "KEY_R",
	Value:  "r",
	Code:   82,
	Keysym: int(0x0072),
}

var KEY_S = types.Key{
	Name:   "KEY_S",
	Value:  "s",
	Code:   83,
	Keysym: int(0x0073),
}

var KEY_T = types.Key{
	Name:   "KEY_T",
	Value:  "t",
	Code:   84,
	Keysym: int(0x0074),
}

var KEY_U = types.Key{
	Name:   "KEY_U",
	Value:  "u",
	Code:   85,
	Keysym: int(0x0075),
}

var KEY_V = types.Key{
	Name:   "KEY_V",
	Value:  "v",
	Code:   86,
	Keysym: int(0x0076),
}

var KEY_W = types.Key{
	Name:   "KEY_W",
	Value:  "w",
	Code:   87,
	Keysym: int(0x0077),
}

var KEY_X = types.Key{
	Name:   "KEY_X",
	Value:  "x",
	Code:   88,
	Keysym: int(0x0078),
}

var KEY_Y = types.Key{
	Name:   "KEY_Y",
	Value:  "y",
	Code:   89,
	Keysym: int(0x0079),
}

var KEY_Z = types.Key{
	Name:   "KEY_Z",
	Value:  "z",
	Code:   90,
	Keysym: int(0x007a),
}

var WIN_LEFT = types.Key{
	Name:   "WIN_LEFT",
	Value:  "Win Left",
	Code:   91,
	Keysym: int(0xFFEB),
}

var WIN_RIGHT = types.Key{
	Name:   "WIN_RIGHT",
	Value:  "Win Right",
	Code:   92,
	Keysym: int(0xFF67),
}

var PAD_0 = types.Key{
	Name:   "PAD_0",
	Value:  "Num Pad 0",
	Code:   96,
	Keysym: int(0xFFB0),
}

var PAD_1 = types.Key{
	Name:   "PAD_1",
	Value:  "Num Pad 1",
	Code:   97,
	Keysym: int(0xFFB1),
}

var PAD_2 = types.Key{
	Name:   "PAD_2",
	Value:  "Num Pad 2",
	Code:   98,
	Keysym: int(0xFFB2),
}

var PAD_3 = types.Key{
	Name:   "PAD_3",
	Value:  "Num Pad 3",
	Code:   99,
	Keysym: int(0xFFB3),
}

var PAD_4 = types.Key{
	Name:   "PAD_4",
	Value:  "Num Pad 4",
	Code:   100,
	Keysym: int(0xFFB4),
}

var PAD_5 = types.Key{
	Name:   "PAD_5",
	Value:  "Num Pad 5",
	Code:   101,
	Keysym: int(0xFFB5),
}

var PAD_6 = types.Key{
	Name:   "PAD_6",
	Value:  "Num Pad 6",
	Code:   102,
	Keysym: int(0xFFB6),
}

var PAD_7 = types.Key{
	Name:   "PAD_7",
	Value:  "Num Pad 7",
	Code:   103,
	Keysym: int(0xFFB7),
}

var PAD_8 = types.Key{
	Name:   "PAD_8",
	Value:  "Num Pad 8",
	Code:   104,
	Keysym: int(0xFFB8),
}

var PAD_9 = types.Key{
	Name:   "PAD_9",
	Value:  "Num Pad 9",
	Code:   105,
	Keysym: int(0xFFB9),
}

var MULTIPLY = types.Key{
	Name:   "MULTIPLY",
	Value:  "*",
	Code:   106,
	Keysym: int(0xFFAA),
}

var ADD = types.Key{
	Name:   "ADD",
	Value:  "+",
	Code:   107,
	Keysym: int(0xFFAB),
}

var SUBTRACT = types.Key{
	Name:   "SUBTRACT",
	Value:  "-",
	Code:   109,
	Keysym: int(0xFFAD),
}

var DECIMAL = types.Key{
	Name:   "DECIMAL",
	Value:  ".",
	Code:   110,
	Keysym: int(0xFFAE),
}

var DIVIDE = types.Key{
	Name:   "DIVIDE",
	Value:  "/",
	Code:   111,
	Keysym: int(0xFFAF),
}

var KEY_F1 = types.Key{
	Name:   "KEY_F1",
	Value:  "f1",
	Code:   112,
	Keysym: int(0xFFBE),
}

var KEY_F2 = types.Key{
	Name:   "KEY_F2",
	Value:  "f2",
	Code:   113,
	Keysym: int(0xFFBF),
}

var KEY_F3 = types.Key{
	Name:   "KEY_F3",
	Value:  "f3",
	Code:   114,
	Keysym: int(0xFFC0),
}

var KEY_F4 = types.Key{
	Name:   "KEY_F4",
	Value:  "f4",
	Code:   115,
	Keysym: int(0xFFC1),
}

var KEY_F5 = types.Key{
	Name:   "KEY_F5",
	Value:  "f5",
	Code:   116,
	Keysym: int(0xFFC2),
}

var KEY_F6 = types.Key{
	Name:   "KEY_F6",
	Value:  "f6",
	Code:   117,
	Keysym: int(0xFFC3),
}

var KEY_F7 = types.Key{
	Name:   "KEY_F7",
	Value:  "f7",
	Code:   118,
	Keysym: int(0xFFC4),
}

var KEY_F8 = types.Key{
	Name:   "KEY_F8",
	Value:  "f8",
	Code:   119,
	Keysym: int(0xFFC5),
}

var KEY_F9 = types.Key{
	Name:   "KEY_F9",
	Value:  "f9",
	Code:   120,
	Keysym: int(0xFFC6),
}

var KEY_F10 = types.Key{
	Name:   "KEY_F10",
	Value:  "f10",
	Code:   121,
	Keysym: int(0xFFC7),
}

var KEY_F11 = types.Key{
	Name:   "KEY_F11",
	Value:  "f11",
	Code:   122,
	Keysym: int(0xFFC8),
}

var KEY_F12 = types.Key{
	Name:   "KEY_F12",
	Value:  "f12",
	Code:   123,
	Keysym: int(0xFFC9),
}

var NUM_LOCK = types.Key{
	Name:   "NUM_LOCK",
	Value:  "Num Lock",
	Code:   144,
	Keysym: int(0xFF7F),
}

var SCROLL_LOCK = types.Key{
	Name:   "SCROLL_LOCK",
	Value:  "Scroll Lock",
	Code:   145,
	Keysym: int(0xFF14),
}

var SEMI_COLON = types.Key{
	Name:   "SEMI_COLON",
	Value:  ";",
	Code:   186,
	Keysym: int(0x003b),
}

var EQUAL = types.Key{
	Name:   "EQUAL",
	Value:  "=",
	Code:   187,
	Keysym: int(0x003d),
}

var COMMA = types.Key{
	Name:   "COMMA",
	Value:  ",",
	Code:   188,
	Keysym: int(0x002c),
}

var DASH = types.Key{
	Name:   "DASH",
	Value:  "-",
	Code:   189,
	Keysym: int(0x002d),
}

var PERIOD = types.Key{
	Name:   "PERIOD",
	Value:  ".",
	Code:   190,
	Keysym: int(0x002e),
}

var FORWARD_SLASH = types.Key{
	Name:   "FORWARD_SLASH",
	Value:  "/",
	Code:   191,
	Keysym: int(0x002f),
}

var GRAVE = types.Key{
	Name:   "GRAVE",
	Value:  "`",
	Code:   192,
	Keysym: int(0x0060),
}

var OPEN_BRACKET = types.Key{
	Name:   "OPEN_BRACKET",
	Value:  "[",
	Code:   219,
	Keysym: int(0x005b),
}

var BACK_SLASH = types.Key{
	Name:   "BACK_SLASH",
	Value:  "\\",
	Code:   220,
	Keysym: int(0x005c),
}

var CLOSE_BRAKET = types.Key{
	Name:   "CLOSE_BRAKET",
	Value:  "]",
	Code:   221,
	Keysym: int(0x005d),
}

var SINGLE_QUOTE = types.Key{
	Name:   "SINGLE_QUOTE",
	Value:  "'",
	Code:   222,
	Keysym: int(0x0022),
}
