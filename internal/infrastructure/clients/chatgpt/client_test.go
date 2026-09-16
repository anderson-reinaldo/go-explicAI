package chatgpt

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

	//go:embed embed/chatgpt_resume_response.json
	chatgptResumeResponse string

	//go:embed embed/chatgpt_resume_response_fail_marsall.json
	chatgptResumeResponseFailMarsall string

	//go:embed embed/chatgpt_resume_response_no_choices.json
	chatgptResumeResponseNoChoices string

	//go:embed embed/chatgpt_resume_response_fail_marshall_arguments.json
	chatgptResumeResponseFailMarshallArguments string

	//go:embed embed/chatgpt_fulltext_response.json
	chatgptFullTextResponse string

	//go:embed embed/chatgpt_fulltext_response_empty.json
	chatgptFullTextResponseEmpty string
)

type (
	ChatgptClientTestSuite struct {
		suite.Suite
		ctx context.Context

		chatgptClient Client
	}
)

func TestChatgptClient(t *testing.T) {
	suite.Run(t, new(ChatgptClientTestSuite))
}

func (s *ChatgptClientTestSuite) SetupTest() {
	s.ctx = context.Background()
	config.AddConfigPath("embed")
	config.SetConfigName("client_config")

	if err := config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed conf file read: %w", err))
	}

	s.chatgptClient = getClientConfig(*config.Sub("chatgpt"))
}

func (s *ChatgptClientTestSuite) TearDownTest() {
	mock.AssertExpectationsForObjects(s.T())
}

func (s *ChatgptClientTestSuite) TestChatgptResume() {
	s.Run("successful request/response", func() {
		var response ChatResumeCompletionsResponse
		json.Unmarshal([]byte(chatgptResumeResponse), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		result, err := s.chatgptClient.Resume(s.ctx, "teste 1")
		s.NoError(err)
		s.Equal("Title test", result.Title)
		s.Equal("Description test", result.Description)
		s.Equal("Brief teste.", result.BriefResume)
		s.Equal("Medium Test", result.MediumResume)

	})

	s.Run("fail response with status code error", func() {
		var response ChatResumeCompletionsResponse
		json.Unmarshal([]byte(chatgptResumeResponse), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusInternalServerError,
			ResponseObject: nil,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.Resume(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt resume request: response= | status=500 Internal Server Error")
	})

	s.Run("fail response with unmarshall error", func() {
		var response []map[string]string
		json.Unmarshal([]byte(chatgptResumeResponseFailMarsall), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.Resume(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt resume request: error=json: cannot unmarshal array into Go value of type chatgpt.ChatResumeCompletionsResponse")

	})

	s.Run("fail response no choices", func() {
		var response ChatResumeCompletionsResponse
		json.Unmarshal([]byte(chatgptResumeResponseNoChoices), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.Resume(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt resume request: no choices in response")

	})

	s.Run("fail response resume marshall arguments", func() {
		var response ChatResumeCompletionsResponse
		json.Unmarshal([]byte(chatgptResumeResponseFailMarshallArguments), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.Resume(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt resume request: error=invalid character 'e' in literal true (expecting 'r')")

	})

}

func (s *ChatgptClientTestSuite) TestChatgptFullText() {
	s.Run("successful request/response", func() {
		var response ChatFullTextCompletionResponse
		json.Unmarshal([]byte(chatgptFullTextResponse), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		result, err := s.chatgptClient.FullTextOrganize(s.ctx, "teste 1")
		s.NoError(err)
		s.Equal("text", *result)
	})

	s.Run("fail response with unmarshall error", func() {
		var response []map[string]string
		json.Unmarshal([]byte(chatgptResumeResponseFailMarsall), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.FullTextOrganize(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt full text organize request: error=json: cannot unmarshal array into Go value of type chatgpt.ChatFullTextCompletionResponse")
	})

	s.Run("fail response with empty response", func() {
		var response ChatFullTextCompletionResponse
		json.Unmarshal([]byte(chatgptFullTextResponseEmpty), &response)

		httpServerMockParams := clients.HttpServerMockParams{
			ExpectedPath:   basePath,
			ExpectedMethod: http.MethodPost,
			ResponseStatus: http.StatusOK,
			ResponseObject: response,
		}

		server := clients.StartMockServer(httpServerMockParams,
			config.Sub("chatgpt").GetString("host"),
		)

		defer server.Close()

		_, err := s.chatgptClient.FullTextOrganize(s.ctx, "teste 1")
		s.Error(err)
		s.EqualError(err, "error on chatgpt full text organize request: empty response")
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
