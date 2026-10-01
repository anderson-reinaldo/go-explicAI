package configuration

import (
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/audiotranscript"
	"github.com/anderson-reinaldo/go-explicAI/internal/gateway/summarize"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/clients/chatgpt"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/clients/whisper"
	"github.com/spf13/viper"
)

type Clients struct {
	AudioTranscript audiotranscript.AudioTranscript
	Summarize       summarize.Summarize
}

func GetClients(config *viper.Viper) *Clients {
	return &Clients{
		AudioTranscript: buildWhisperClient(config.Sub("whisper")),
		Summarize:       buildChatgptClient(config.Sub("chatgpt")),
	}
}

func buildWhisperClient(config *viper.Viper) audiotranscript.AudioTranscript {
	return whisper.NewClient(
		config.GetString("name"),
		config.GetString("host"),
		openAiApiKey,
		config.GetString("model"),
		config.GetInt64("timeout"),
	)
}

func buildChatgptClient(config *viper.Viper) summarize.Summarize {
	return chatgpt.NewClient(
		config.GetString("name"),
		config.GetString("host"),
		openAiApiKey,
		config.GetString("model"),
		config.GetInt64("timeout"),
	)
}
