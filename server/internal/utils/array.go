package utils

func ArrayIn[T comparable](val T, array any) (exists bool, index int) {
	index = -1
	slice, _ := array.([]T)

	for i, v := range slice {
		if v == val {
			index = i
			exists = true
			return
		}
	}

	return
}
