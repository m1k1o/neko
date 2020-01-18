package utils

import "encoding/json"

func Unmarshal(in interface{}, raw []byte, callback func() error) error {
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return callback()
}
