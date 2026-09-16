package configuration

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	config = viper.New()
)

func Init() *viper.Viper {
	godotenv.Load()
	config.AutomaticEnv()
	defaultConfigs()
	return config
}

func defaultConfigs() {
	config.SetDefault("server.host", fmt.Sprintf("%s:%s", config.Get("HOST"), config.Get("PORT")))
	config.SetDefault("app.name", "explicaAI")
	config.SetDefault("openai.apikey", config.Get("OPENAI_API_KEY"))
}
