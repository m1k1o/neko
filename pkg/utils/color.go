package utils

import (
	"fmt"
	"regexp"
)

const (
	char = "&"
)

// Colors: http://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html
var re = regexp.MustCompile(char + `(?m)([0-9]{1,2};[0-9]{1,2}|[0-9]{1,2})`)

func Color(str string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + "\033[" + groups[1] + "m"
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func Colorf(format string, a ...interface{}) string {
	return fmt.Sprintf(Color(format), a...)
}
