package utils

import (
	"regexp"
	"strings"
)

var unsafeIDChars = regexp.MustCompile(`[^\p{L}\p{N}._-]+`)

// SafeID maps a user-supplied name to a string that can be used as a
// session or member id in URL paths: letters and digits (any script), '.',
// '_' and '-' are kept, every run of other characters becomes a single '-',
// leading and trailing '-' and '.' are removed, the result is cut to 64
// characters and an empty result becomes "user".
//
// Nothing that has a meaning in a URL path survives ("/", "?", "#", "%",
// spaces), and trimming '.' keeps "." and ".." from ever being an id. Letters
// outside ASCII are kept so that names such as "Müller" or "山田" stay
// recognisable; clients percent-encode ids when they put them in a path.
func SafeID(name string) string {
	id := strings.Trim(unsafeIDChars.ReplaceAllString(name, "-"), "-.")
	if r := []rune(id); len(r) > 64 {
		id = strings.TrimRight(string(r[:64]), "-.")
	}
	if id == "" {
		return "user"
	}
	return id
}
