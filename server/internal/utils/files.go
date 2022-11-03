package utils

import (
	"os"

	"m1k1o/neko/internal/types"
)

func ListFiles(path string) (*[]types.FileListItem, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]types.FileListItem, len(items))
	for i, item := range items {
		var itemType string = ""
		if item.IsDir() {
			itemType = "dir"
		} else {
			itemType = "file"
		}
		out[i] = types.FileListItem{
			Filename: item.Name(),
			Type:     itemType,
		}
	}

	return &out, nil
}
