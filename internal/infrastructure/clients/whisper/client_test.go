package whisper

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/clients"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	config = viper.New()

	//go:embed embed/whisper_response.json
	whisperResponse string

	//go:embed embed/whisper_response_error_marshall.json
	whisperResponseError string

	//go:embed embed/whisper_response_empty.json
	whisperResponseEmpty string
)

type (
	WhisperClientTestSuite struct {
		suite.Suite
		ctx           context.Context
		whisperClient Client
	}
)

func TestWhisperClient(t *testing.T) {
	suite.Run(t, new(WhisperClientTestSuite))
}

func (s *WhisperClientTestSuite) SetupTest() {
	s.ctx = context.Background()
	config.AddConfigPath("embed")
	config.SetConfigName("client_config")

	if err := config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed conf file read: %w", err))
	}

	s.whisperClient = getClientConfig(*config.Sub("whisper"))
}

func (s *WhisperClientTestSuite) TearDownTest() {
	mock.AssertExpectationsForObjects(s.T())
}

func (s *WhisperClientTestSuite) TestWhisper() {
	s.Run("successful request/response", func() {
		var response Response
		json.Unmarshal([]byte(whisperResponse), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("whisper").GetString("host"),
		)

		defer server.Close()

		text, err := s.whisperClient.Transcribe(s.ctx, []byte{})

		s.Require().NoError(err)
		s.Equal("teste ok", *text)
	})

	s.Run("fail response with status code error", func() {
		var response Response
		json.Unmarshal([]byte(whisperResponse), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusInternalServerError,
			ResponseObject: nil,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("whisper").GetString("host"),
		)

		defer server.Close()

		_, err := s.whisperClient.Transcribe(s.ctx, []byte{})

		s.Require().Error(err)
		s.EqualError(err, "error on whisper request: response= | status=500 Internal Server Error")
	})

	s.Run("fail response with unmarshall error", func() {
		var response []map[string]string
		json.Unmarshal([]byte(whisperResponseError), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("whisper").GetString("host"),
		)

		defer server.Close()

		_, err := s.whisperClient.Transcribe(s.ctx, []byte{})

		s.Require().Error(err)
		s.EqualError(err, "error on whisper request: error=json: cannot unmarshal array into Go value of type whisper.Response")
	})

	s.Run("fail response with empty response", func() {
		var response Response
		json.Unmarshal([]byte(whisperResponseEmpty), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("whisper").GetString("host"),
		)

		defer server.Close()

		_, err := s.whisperClient.Transcribe(s.ctx, []byte{})

		s.Require().Error(err)
		s.EqualError(err, "error on whisper request: error=empty response")
	})

}

func getClientConfig(viper viper.Viper) Client {
	client := NewClient(
		viper.GetString("name"),
		viper.GetString("url"),
		viper.GetString("apiKey"),
		viper.GetString("model"),
		viper.GetInt64("timeout"),
	)

	return *client
}
