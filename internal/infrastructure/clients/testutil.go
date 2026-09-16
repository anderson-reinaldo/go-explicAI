package clients

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
)

type (
	HttpServerMockParams struct {
		ResponseObject interface{}
		ExpectedPath   string
		ExpectedMethod string
		ResponseStatus int
	}
)

func StartMockServer(params HttpServerMockParams, host string) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(mockHandler(params)))
	listener, err := net.Listen("tcp", host)
	if err != nil {
		panic(fmt.Errorf("failed listen http: %w", err))
	}

	defer server.Listener.Close()
	server.Listener = listener
	server.Start()

	return server
}

func mockHandler(params HttpServerMockParams) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != params.ExpectedPath || r.Method != params.ExpectedMethod {
			http.Error(w, "unexpected request", http.StatusNotImplemented)
			return
		}

		writeJSONResponse(w, params.ResponseStatus, params.ResponseObject)
	}
}

func writeJSONResponse(w http.ResponseWriter, status int, responseObject interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if responseObject != nil {
		_ = json.NewEncoder(w).Encode(responseObject)
	}
}
