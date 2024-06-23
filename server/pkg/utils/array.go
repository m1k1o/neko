package utils

func ArrayIn[T comparable](val T, array []T) (exists bool, index int) {
	exists, index = false, -1

	for i, a := range array {
		if a == val {
			exists, index = true, i
			return
		}
	}

	return
}
