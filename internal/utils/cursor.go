package utils

import (
	"bytes"
    "image"
    "image/color"
	"image/png"
	"encoding/base64"

	"demodesk/neko/internal/types"
)

func GetCursorImageURI(cursor *types.CursorImage) (string, error) {
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
		return "", err
	}

	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes())
	return uri, nil
}
