package utils

import "fmt"

func ErrorsJoin(errs []error) error {
	var resErr error

	for _, err := range errs {
		if err == nil {
			continue
		}

		if resErr == nil {
			resErr = err
			continue
		}

		resErr = fmt.Errorf("%w; %s", resErr, err.Error())
	}

	return resErr
}
