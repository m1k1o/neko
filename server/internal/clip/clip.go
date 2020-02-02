// NOTE: I have no fucking clue what I'm doing with this,
// it works, but I am positive I'm doing this very wrong...
// should I be freeing these strings? does go cg them?
// pretty sure this *isn't* thread safe either.... /shrug

package clip

/*
#cgo linux LDFLAGS: -lclipboard

#include "clip.h"
*/
import "C"

func Read() string {
	return C.GoString(C.get_clipboard())
}

func Write(data string) {
	C.set_clipboard(C.CString(data))
}
