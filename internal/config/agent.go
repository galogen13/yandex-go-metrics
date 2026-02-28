package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type AgentConfig struct {
	Host           string `json:"address" mapstructure:"address"`                 // адрес сервера, на который будут отправляться метрики
	ReportInterval int    `json:"report_interval" mapstructure:"report_interval"` // количество секунд между отправками метрик на сервер
	PollInterval   int    `json:"poll_interval" mapstructure:"poll_interval"`     // количество секунд между сборами значений метрик
	Key            string `json:"key" mapstructure:"key"`                         // ключ
	RateLimit      int    `json:"rate_limit" mapstructure:"rate_limit"`           // максимальное количество горутин, одновременно отправляющих данные на сервер
	CryptoKeyPath  string `json:"crypto_key" mapstructure:"crypto_key"`           // путь к публичному ключу
}

type FileAgentConfig struct {
	Host           string `json:"address"`         // адрес сервера, на который будут отправляться метрики
	ReportInterval string `json:"report_interval"` // время между отправками метрик на сервер
	PollInterval   string `json:"poll_interval"`   // время между сборами значений метрик
	Key            string `json:"key"`             // ключ
	RateLimit      int    `json:"rate_limit"`      // максимальное количество горутин, одновременно отправляющих данные на сервер
	CryptoKeyPath  string `json:"crypto_key"`      // путь к публичному ключу
}

func GetAgentConfig() (AgentConfig, error) {

	viper.SetDefault("address", "localhost:8080")
	viper.SetDefault("report_interval", 10)
	viper.SetDefault("poll_interval", 2)
	viper.SetDefault("key", "secret_key")
	viper.SetDefault("rate_limit", 1)
	viper.SetDefault("crypto_key", "")
	viper.SetDefault("config", "")

	pflag.StringP("address", "a", viper.GetString("address"), "server address")
	pflag.IntP("report-interval", "r", viper.GetInt("report_interval"), "report interval")
	pflag.IntP("poll-interval", "p", viper.GetInt("poll_interval"), "poll interval")
	pflag.IntP("rate-limit", "l", viper.GetInt("rate_limit"), "rate limit")
	pflag.StringP("key", "k", viper.GetString("key"), "secret key")
	pflag.String("crypto-key", viper.GetString("crypto_key"), "path to crypto key")
	pflag.StringP("config", "c", viper.GetString("config"), "path to configuration file")
	pflag.Parse()

	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		if f := pflag.Lookup("config"); f != nil {
			configPath = f.Value.String()
		}
	}

	if err := parseConfigFile(configPath); err != nil {
		return AgentConfig{}, err
	}

	bindPFlags()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	viper.BindEnv("address", "ADDRESS")
	viper.BindEnv("report_interval", "REPORT_INTERVAL")
	viper.BindEnv("poll_interval", "POLL_INTERVAL")
	viper.BindEnv("key", "KEY")
	viper.BindEnv("rate_limit", "RATE_LIMIT")
	viper.BindEnv("crypto_key", "CRYPTO_KEY")
	viper.BindEnv("config", "CONFIG")

	var cfg AgentConfig

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func parseConfigFile(configPath string) error {
	if configPath == "" {
		return nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var fileConfig FileAgentConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fileConfig)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	if fileConfig.Host != "" {
		viper.Set("address", fileConfig.Host)
	}
	if fileConfig.ReportInterval != "" {
		reportIntervalDuration, err := time.ParseDuration(fileConfig.ReportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse ReportInterval duration: %w", err)
		}
		viper.Set("report_interval", int(reportIntervalDuration.Seconds()))
	}
	if fileConfig.PollInterval != "" {
		pollIntervalDuration, err := time.ParseDuration(fileConfig.PollInterval)
		if err != nil {
			return fmt.Errorf("failed to parse PollInterval duration: %w", err)
		}
		viper.Set("poll_interval", int(pollIntervalDuration.Seconds()))
	}
	if fileConfig.Key != "" {
		viper.Set("key", fileConfig.Key)
	}
	if fileConfig.CryptoKeyPath != "" {
		viper.Set("crypto_key", fileConfig.CryptoKeyPath)
	}
	if fileConfig.RateLimit != 0 {
		viper.Set("rate_limit", fileConfig.RateLimit)
	}

	return nil
}

func bindPFlags() {
	pflag.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			name := strings.ReplaceAll(f.Name, "-", "_")
			viper.BindPFlag(name, f)
		}
	})
}
