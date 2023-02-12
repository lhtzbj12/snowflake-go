package httputil

import (
	"bytes"
	"net/http"
	"strings"
	"time"
)

func HttpGet(url string, timeoutSeconds int) (*http.Response, error) {
	c := http.Client{
		Timeout: time.Second * time.Duration(timeoutSeconds),
	}
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func HttpPost(url string, body []byte, timeoutSeconds int) (*http.Response, error) {
	c := http.Client{
		Timeout: time.Second * time.Duration(timeoutSeconds),
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
