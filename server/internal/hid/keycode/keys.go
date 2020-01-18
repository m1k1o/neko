package keycode

type Key struct {
	Name   string
	Value  string
	Code   int
	Keysym int
}

var BACKSPACE = Key{
	Name:   "BACKSPACE",
	Value:  "BackSpace",
	Code:   8,
	Keysym: int(0xff08),
}

var TAB = Key{
	Name:   "TAB",
	Value:  "Tab",
	Code:   9,
	Keysym: int(0xFF09),
}

var CLEAR = Key{
	Name:   "CLEAR",
	Value:  "Clear",
	Code:   12,
	Keysym: int(0xFF0B),
}

var ENTER = Key{
	Name:   "ENTER",
	Value:  "Enter",
	Code:   13,
	Keysym: int(0xFF0D),
}

var SHIFT = Key{
	Name:   "SHIFT",
	Value:  "Shift",
	Code:   16,
	Keysym: int(0xFFE1),
}

var CTRL = Key{
	Name:   "CTRL",
	Value:  "Ctrl",
	Code:   17,
	Keysym: int(0xFFE3),
}

var ALT = Key{
	Name:   "ALT",
	Value:  "Alt",
	Code:   18,
	Keysym: int(0xFFE9),
}

var PAUSE = Key{
	Name:   "PAUSE",
	Value:  "Pause",
	Code:   19,
	Keysym: int(0xFF13),
}

var CAPS_LOCK = Key{
	Name:   "CAPS_LOCK",
	Value:  "Caps Lock",
	Code:   20,
	Keysym: int(0xFFE5),
}

var ESCAPE = Key{
	Name:   "ESCAPE",
	Value:  "Escape",
	Code:   27,
	Keysym: int(0xFF1B),
}

var SPACE = Key{
	Name:   "SPACE",
	Value:  " ",
	Code:   32,
	Keysym: int(0x0020),
}

var PAGE_UP = Key{
	Name:   "PAGE_UP",
	Value:  "Page Up",
	Code:   33,
	Keysym: int(0xFF55),
}

var PAGE_DOWN = Key{
	Name:   "PAGE_DOWN",
	Value:  "Page Down",
	Code:   34,
	Keysym: int(0xFF56),
}

var END = Key{
	Name:   "END",
	Value:  "End",
	Code:   35,
	Keysym: int(0xFF57),
}

var HOME = Key{
	Name:   "HOME",
	Value:  "Home",
	Code:   36,
	Keysym: int(0xFF50),
}

var LEFT_ARROW = Key{
	Name:   "LEFT_ARROW",
	Value:  "Left Arrow",
	Code:   37,
	Keysym: int(0xFF51),
}

var UP_ARROW = Key{
	Name:   "UP_ARROW",
	Value:  "Up Arrow",
	Code:   38,
	Keysym: int(0xFF52),
}

var RIGHT_ARROW = Key{
	Name:   "RIGHT_ARROW",
	Value:  "Right Arrow",
	Code:   39,
	Keysym: int(0xFF53),
}

var DOWN_ARROW = Key{
	Name:   "DOWN_ARROW",
	Value:  "Down Arrow",
	Code:   40,
	Keysym: int(0xFF54),
}

var INSERT = Key{
	Name:   "INSERT",
	Value:  "Insert",
	Code:   45,
	Keysym: int(0xFF63),
}

var DELETE = Key{
	Name:   "DELETE",
	Value:  "Delete",
	Code:   46,
	Keysym: int(0xFFFF),
}

var KEY_0 = Key{
	Name:   "KEY_0",
	Value:  "0",
	Code:   48,
	Keysym: int(0x0030),
}

var KEY_1 = Key{
	Name:   "KEY_1",
	Value:  "1",
	Code:   49,
	Keysym: int(0x0031),
}

var KEY_2 = Key{
	Name:   "KEY_2",
	Value:  "2",
	Code:   50,
	Keysym: int(0x0032),
}

var KEY_3 = Key{
	Name:   "KEY_3",
	Value:  "3",
	Code:   51,
	Keysym: int(0x0033),
}

var KEY_4 = Key{
	Name:   "KEY_4",
	Value:  "4",
	Code:   52,
	Keysym: int(0x0034),
}

var KEY_5 = Key{
	Name:   "KEY_5",
	Value:  "5",
	Code:   53,
	Keysym: int(0x0035),
}

var KEY_6 = Key{
	Name:   "KEY_6",
	Value:  "6",
	Code:   54,
	Keysym: int(0x0036),
}

var KEY_7 = Key{
	Name:   "KEY_7",
	Value:  "7",
	Code:   55,
	Keysym: int(0x0037),
}

var KEY_8 = Key{
	Name:   "KEY_8",
	Value:  "8",
	Code:   56,
	Keysym: int(0x0038),
}

var KEY_9 = Key{
	Name:   "KEY_9",
	Value:  "9",
	Code:   57,
	Keysym: int(0x0039),
}

var KEY_A = Key{
	Name:   "KEY_A",
	Value:  "a",
	Code:   65,
	Keysym: int(0x0061),
}

var KEY_B = Key{
	Name:   "KEY_B",
	Value:  "b",
	Code:   66,
	Keysym: int(0x0062),
}

var KEY_C = Key{
	Name:   "KEY_C",
	Value:  "c",
	Code:   67,
	Keysym: int(0x0063),
}

var KEY_D = Key{
	Name:   "KEY_D",
	Value:  "d",
	Code:   68,
	Keysym: int(0x0064),
}

var KEY_E = Key{
	Name:   "KEY_E",
	Value:  "e",
	Code:   69,
	Keysym: int(0x0065),
}

var KEY_F = Key{
	Name:   "KEY_F",
	Value:  "f",
	Code:   70,
	Keysym: int(0x0066),
}

var KEY_G = Key{
	Name:   "KEY_G",
	Value:  "g",
	Code:   71,
	Keysym: int(0x0067),
}

var KEY_H = Key{
	Name:   "KEY_H",
	Value:  "h",
	Code:   72,
	Keysym: int(0x0068),
}

var KEY_I = Key{
	Name:   "KEY_I",
	Value:  "i",
	Code:   73,
	Keysym: int(0x0069),
}

var KEY_J = Key{
	Name:   "KEY_J",
	Value:  "j",
	Code:   74,
	Keysym: int(0x006a),
}

var KEY_K = Key{
	Name:   "KEY_K",
	Value:  "k",
	Code:   75,
	Keysym: int(0x006b),
}

var KEY_L = Key{
	Name:   "KEY_L",
	Value:  "l",
	Code:   76,
	Keysym: int(0x006c),
}

var KEY_M = Key{
	Name:   "KEY_M",
	Value:  "m",
	Code:   77,
	Keysym: int(0x006d),
}

var KEY_N = Key{
	Name:   "KEY_N",
	Value:  "n",
	Code:   78,
	Keysym: int(0x006e),
}

var KEY_O = Key{
	Name:   "KEY_O",
	Value:  "o",
	Code:   79,
	Keysym: int(0x006f),
}

var KEY_P = Key{
	Name:   "KEY_P",
	Value:  "p",
	Code:   80,
	Keysym: int(0x0070),
}

var KEY_Q = Key{
	Name:   "KEY_Q",
	Value:  "q",
	Code:   81,
	Keysym: int(0x0071),
}

var KEY_R = Key{
	Name:   "KEY_R",
	Value:  "r",
	Code:   82,
	Keysym: int(0x0072),
}

var KEY_S = Key{
	Name:   "KEY_S",
	Value:  "s",
	Code:   83,
	Keysym: int(0x0073),
}

var KEY_T = Key{
	Name:   "KEY_T",
	Value:  "t",
	Code:   84,
	Keysym: int(0x0074),
}

var KEY_U = Key{
	Name:   "KEY_U",
	Value:  "u",
	Code:   85,
	Keysym: int(0x0075),
}

var KEY_V = Key{
	Name:   "KEY_V",
	Value:  "v",
	Code:   86,
	Keysym: int(0x0076),
}

var KEY_W = Key{
	Name:   "KEY_W",
	Value:  "w",
	Code:   87,
	Keysym: int(0x0077),
}

var KEY_X = Key{
	Name:   "KEY_X",
	Value:  "x",
	Code:   88,
	Keysym: int(0x0078),
}

var KEY_Y = Key{
	Name:   "KEY_Y",
	Value:  "y",
	Code:   89,
	Keysym: int(0x0079),
}

var KEY_Z = Key{
	Name:   "KEY_Z",
	Value:  "z",
	Code:   90,
	Keysym: int(0x007a),
}

var WIN_LEFT = Key{
	Name:   "WIN_LEFT",
	Value:  "Win Left",
	Code:   91,
	Keysym: int(0xFFEB),
}

var WIN_RIGHT = Key{
	Name:   "WIN_RIGHT",
	Value:  "Win Right",
	Code:   92,
	Keysym: int(0xFF67),
}

var PAD_0 = Key{
	Name:   "PAD_0",
	Value:  "Num Pad 0",
	Code:   96,
	Keysym: int(0xFFB0),
}

var PAD_1 = Key{
	Name:   "PAD_1",
	Value:  "Num Pad 1",
	Code:   97,
	Keysym: int(0xFFB1),
}

var PAD_2 = Key{
	Name:   "PAD_2",
	Value:  "Num Pad 2",
	Code:   98,
	Keysym: int(0xFFB2),
}

var PAD_3 = Key{
	Name:   "PAD_3",
	Value:  "Num Pad 3",
	Code:   99,
	Keysym: int(0xFFB3),
}

var PAD_4 = Key{
	Name:   "PAD_4",
	Value:  "Num Pad 4",
	Code:   100,
	Keysym: int(0xFFB4),
}

var PAD_5 = Key{
	Name:   "PAD_5",
	Value:  "Num Pad 5",
	Code:   101,
	Keysym: int(0xFFB5),
}

var PAD_6 = Key{
	Name:   "PAD_6",
	Value:  "Num Pad 6",
	Code:   102,
	Keysym: int(0xFFB6),
}

var PAD_7 = Key{
	Name:   "PAD_7",
	Value:  "Num Pad 7",
	Code:   103,
	Keysym: int(0xFFB7),
}

var PAD_8 = Key{
	Name:   "PAD_8",
	Value:  "Num Pad 8",
	Code:   104,
	Keysym: int(0xFFB8),
}

var PAD_9 = Key{
	Name:   "PAD_9",
	Value:  "Num Pad 9",
	Code:   105,
	Keysym: int(0xFFB9),
}

var MULTIPLY = Key{
	Name:   "MULTIPLY",
	Value:  "*",
	Code:   106,
	Keysym: int(0xFFAA),
}

var ADD = Key{
	Name:   "ADD",
	Value:  "+",
	Code:   107,
	Keysym: int(0xFFAB),
}

var SUBTRACT = Key{
	Name:   "SUBTRACT",
	Value:  "-",
	Code:   109,
	Keysym: int(0xFFAD),
}

var DECIMAL = Key{
	Name:   "DECIMAL",
	Value:  ".",
	Code:   110,
	Keysym: int(0xFFAE),
}

var DIVIDE = Key{
	Name:   "DIVIDE",
	Value:  "/",
	Code:   111,
	Keysym: int(0xFFAF),
}

var KEY_F1 = Key{
	Name:   "KEY_F1",
	Value:  "f1",
	Code:   112,
	Keysym: int(0xFFBE),
}

var KEY_F2 = Key{
	Name:   "KEY_F2",
	Value:  "f2",
	Code:   113,
	Keysym: int(0xFFBF),
}

var KEY_F3 = Key{
	Name:   "KEY_F3",
	Value:  "f3",
	Code:   114,
	Keysym: int(0xFFC0),
}

var KEY_F4 = Key{
	Name:   "KEY_F4",
	Value:  "f4",
	Code:   115,
	Keysym: int(0xFFC1),
}

var KEY_F5 = Key{
	Name:   "KEY_F5",
	Value:  "f5",
	Code:   116,
	Keysym: int(0xFFC2),
}

var KEY_F6 = Key{
	Name:   "KEY_F6",
	Value:  "f6",
	Code:   117,
	Keysym: int(0xFFC3),
}

var KEY_F7 = Key{
	Name:   "KEY_F7",
	Value:  "f7",
	Code:   118,
	Keysym: int(0xFFC4),
}

var KEY_F8 = Key{
	Name:   "KEY_F8",
	Value:  "f8",
	Code:   119,
	Keysym: int(0xFFC5),
}

var KEY_F9 = Key{
	Name:   "KEY_F9",
	Value:  "f9",
	Code:   120,
	Keysym: int(0xFFC6),
}

var KEY_F10 = Key{
	Name:   "KEY_F10",
	Value:  "f10",
	Code:   121,
	Keysym: int(0xFFC7),
}

var KEY_F11 = Key{
	Name:   "KEY_F11",
	Value:  "f11",
	Code:   122,
	Keysym: int(0xFFC8),
}

var KEY_F12 = Key{
	Name:   "KEY_F12",
	Value:  "f12",
	Code:   123,
	Keysym: int(0xFFC9),
}

var NUM_LOCK = Key{
	Name:   "NUM_LOCK",
	Value:  "Num Lock",
	Code:   144,
	Keysym: int(0xFF7F),
}

var SCROLL_LOCK = Key{
	Name:   "SCROLL_LOCK",
	Value:  "Scroll Lock",
	Code:   145,
	Keysym: int(0xFF14),
}

var SEMI_COLON = Key{
	Name:   "SEMI_COLON",
	Value:  ";",
	Code:   186,
	Keysym: int(0x003b),
}

var EQUAL = Key{
	Name:   "EQUAL",
	Value:  "=",
	Code:   187,
	Keysym: int(0x003d),
}

var COMMA = Key{
	Name:   "COMMA",
	Value:  ",",
	Code:   188,
	Keysym: int(0x002c),
}

var DASH = Key{
	Name:   "DASH",
	Value:  "-",
	Code:   189,
	Keysym: int(0x002d),
}

var PERIOD = Key{
	Name:   "PERIOD",
	Value:  ".",
	Code:   190,
	Keysym: int(0x002e),
}

var FORWARD_SLASH = Key{
	Name:   "FORWARD_SLASH",
	Value:  "/",
	Code:   191,
	Keysym: int(0x002f),
}

var GRAVE = Key{
	Name:   "GRAVE",
	Value:  "`",
	Code:   192,
	Keysym: int(0x0060),
}

var OPEN_BRACKET = Key{
	Name:   "OPEN_BRACKET",
	Value:  "[",
	Code:   219,
	Keysym: int(0x005b),
}

var BACK_SLASH = Key{
	Name:   "BACK_SLASH",
	Value:  "\\",
	Code:   220,
	Keysym: int(0x005c),
}

var CLOSE_BRAKET = Key{
	Name:   "CLOSE_BRAKET",
	Value:  "]",
	Code:   221,
	Keysym: int(0x005d),
}

var SINGLE_QUOTE = Key{
	Name:   "SINGLE_QUOTE",
	Value:  "'",
	Code:   222,
	Keysym: int(0x0022),
}
