package utils

import (
	"bytes"
	"encoding/base64"
	"image/png"

	"demodesk/neko/internal/types"
)

func GetCursorImage(cursor *types.CursorImage) ([]byte, error) {
	out := new(bytes.Buffer)
	err := png.Encode(out, cursor.Image)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func GetCursorImageURI(cursor *types.CursorImage) (string, error) {
	img, err := GetCursorImage(cursor)
	if err != nil {
		return "", err
	}

	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(img)
	return uri, nil
}
