package clients

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type BaseHTTP struct {
	Client *resty.Request
}

func NewHttpClient(URL string, timeout int64) *BaseHTTP {
	httpClient := resty.New().SetBaseURL(URL).SetTimeout(time.Duration(timeout) * time.Millisecond).R()

	return &BaseHTTP{Client: httpClient}
}
