package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"demodesk/neko/internal/types"
)

func GetCursorImage(cursor *types.CursorImage) ([]byte, error) {
	width := int(cursor.Width)
	height := int(cursor.Height)
	pixels := cursor.Pixels

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			pos := ((row * height) + col) * 8

			img.SetRGBA(col, row, color.RGBA{
				A: pixels[pos+3],
				R: pixels[pos+2],
				G: pixels[pos+1],
				B: pixels[pos+0],
			})
		}
	}

	out := new(bytes.Buffer)
	err := png.Encode(out, img)
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
