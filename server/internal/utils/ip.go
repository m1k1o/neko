package utils

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

// dig @resolver1.opendns.com ANY myip.opendns.com +short -4

func GetIP(serverUrl string) (string, error) {
	tr := &http.Transport{
		Proxy: nil, // ignore proxy
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          30,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{Transport: tr}
	rsp, err := client.Get(serverUrl)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
