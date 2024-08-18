package utils

func ArrayIn[T comparable](val T, array []T) (exists bool, index int) {
	index = -1

	for i, v := range array {
		if v == val {
			index = i
			exists = true
			return
		}
	}

	return
}
