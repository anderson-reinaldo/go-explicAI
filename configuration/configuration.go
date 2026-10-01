package configuration

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/log"
	"github.com/spf13/viper"
)

var (
	openAiApiKey = os.Getenv("OPENAI_API_KEY")
	config       = viper.New()
)

func Init() *viper.Viper {
	defaultConfigs()

	if openAiApiKey == "" {
		log.LogError(context.Background(), "OPENAI_API_KEY is not set",
			errors.New("openai api key is required"))
	}
	return config
}

func defaultConfigs() {
	config.SetDefault("server.host", fmt.Sprintf("%s:%s", config.Get("HOST"), config.Get("PORT")))
	config.SetDefault("app.name", "explicaAI")
	config.SetDefault("whisper.name", "whisper")
	config.SetDefault("whisper.url", "api.openai.com")
	config.SetDefault("whisper.host", "https://api.openai.com")
	config.SetDefault("whisper.timeout", 3000)
	config.SetDefault("whisper.model", "whisper-1")
	config.SetDefault("chatgpt.url", "api.openai.com")
	config.SetDefault("chatgpt.host", "https://api.openai.com")
	config.SetDefault("chatgpt.timeout", 3000)
	config.SetDefault("chatgpt.model", "gpt-4o")
	config.SetDefault("database.url", "postgre://admin:admin@localhost:5432/explicai")
}
