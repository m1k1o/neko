package utils

import (
	"os"

	"m1k1o/neko/internal/types"
)

func ListFiles(path string) ([]types.FileListItem, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]types.FileListItem, len(items))
	for i, item := range items {
		var itemType string = ""
		var size int64 = 0
		if item.IsDir() {
			itemType = "dir"
		} else {
			itemType = "file"
			info, err := item.Info()
			if err == nil {
				size = info.Size()
			}
		}
		out[i] = types.FileListItem{
			Filename: item.Name(),
			Type:     itemType,
			Size:     size,
		}
	}

	return out, nil
}
