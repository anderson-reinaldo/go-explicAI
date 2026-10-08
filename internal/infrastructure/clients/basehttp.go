package clients

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var Mutex sync.Mutex

type BaseHTTP struct {
	Client *resty.Request
}

func NewHttpClient(URL string, timeout int64) *BaseHTTP {
	httpClient := resty.New().SetBaseURL(URL).SetTimeout(time.Duration(timeout) * time.Millisecond).R()

	return &BaseHTTP{Client: httpClient}
}
