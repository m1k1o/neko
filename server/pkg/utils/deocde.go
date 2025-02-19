package utils

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func Decode(input interface{}, output interface{}) error {
	return mapstructure.Decode(input, output)
}

func Unmarshal(in any, raw []byte, callback func() error) error {
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return callback()
}

func JsonStringAutoDecode(m any) func(rf reflect.Kind, rt reflect.Kind, data any) (any, error) {
	return func(rf reflect.Kind, rt reflect.Kind, data any) (any, error) {
		if rf != reflect.String || rt == reflect.String {
			return data, nil
		}

		raw := data.(string)
		if raw != "" && (raw[0:1] == "{" || raw[0:1] == "[") {
			err := json.Unmarshal([]byte(raw), &m)
			return m, err
		}

		return data, nil
	}
}
