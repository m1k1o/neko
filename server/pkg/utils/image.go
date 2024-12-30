package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
)

func CreatePNGImage(img *image.RGBA) ([]byte, error) {
	out := new(bytes.Buffer)
	err := png.Encode(out, img)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func CreateJPGImage(img *image.RGBA, quality int) ([]byte, error) {
	out := new(bytes.Buffer)
	err := jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func CreatePNGImageURI(img *image.RGBA) (string, error) {
	data, err := CreatePNGImage(img)
	if err != nil {
		return "", err
	}

	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	return uri, nil
}
