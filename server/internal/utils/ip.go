package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// dig @resolver1.opendns.com ANY myip.opendns.com +short -4

func GetIP() (string, error) {
	rsp, err := http.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
