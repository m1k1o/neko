package filetransfer

import "os"

func ListFiles(path string) ([]Item, error) {
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	out := make([]Item, len(items))
	for i, item := range items {
		var itemType ItemType
		var size int64 = 0
		if item.IsDir() {
			itemType = ItemTypeDir
		} else {
			itemType = ItemTypeFile
			info, err := item.Info()
			if err == nil {
				size = info.Size()
			}
		}
		out[i] = Item{
			Name: item.Name(),
			Type: itemType,
			Size: size,
		}
	}

	return out, nil
}
