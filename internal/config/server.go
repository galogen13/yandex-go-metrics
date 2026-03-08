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

// ServerConfig - структура параметров сервера
type ServerConfig struct {
	Host                 string `json:"address" mapstructure:"address"`
	LogLevel             string `json:"log_level" mapstructure:"log_level"`
	StoreInterval        *int   `json:"store_interval" mapstructure:"store_interval"` // указатель, т.к. в переменной может быть 0, что важно для нас
	FileStoragePath      string `json:"file_storage_path" mapstructure:"file_storage_path"`
	RestoreStorage       *bool  `json:"restore" mapstructure:"restore"`
	DatabaseDSN          string `json:"database_dsn" mapstructure:"database_dsn"`
	Key                  string `json:"key" mapstructure:"key"`
	AuditFile            string `json:"audit_file" mapstructure:"audit_file"`
	AuditURL             string `json:"audit_url" mapstructure:"audit_url"`
	CryptoKeyPath        string `json:"crypto_key" mapstructure:"crypto_key"` // путь к приватному ключу
	UseDatabaseAsStorage bool
	StoreOnUpdate        bool
	StorePeriodically    bool
}

type FileServerConfig struct {
	Host            string `json:"address"`
	LogLevel        string `json:"log_level"`
	StoreInterval   string `json:"store_interval"` // указатель, т.к. в переменной может быть 0, что важно для нас
	FileStoragePath string `json:"file_storage_path"`
	RestoreStorage  *bool  `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	Key             string `json:"key"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	CryptoKeyPath   string `json:"crypto_key"` // путь к приватному ключу
}

func GetServerConfig() (*ServerConfig, error) {

	viper.SetDefault("address", "localhost:8080")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("store_interval", 300)
	viper.SetDefault("file_storage_path", "./metricsstorage")
	viper.SetDefault("restore", false)
	viper.SetDefault("database_dsn", "")
	viper.SetDefault("key", "")
	viper.SetDefault("audit_file", "")
	viper.SetDefault("audit_url", "")
	viper.SetDefault("crypto_key", "")
	viper.SetDefault("config", "")

	pflag.StringP("address", "a", viper.GetString("address"), "server address")
	pflag.StringP("log-level", "l", viper.GetString("log_level"), "log level")
	pflag.IntP("store-interval", "i", viper.GetInt("store_interval"), "data store interval")
	pflag.StringP("file-storage", "f", viper.GetString("file_storage_path"), "file storage path")
	pflag.BoolP("restore", "r", viper.GetBool("restore"), "restore storage from file")
	pflag.StringP("database-dsn", "d", viper.GetString("database_dsn"), "database DSN")
	pflag.String("k", viper.GetString("key"), "secret key")
	pflag.String("audit-file", viper.GetString("audit_file"), "audit file")
	pflag.String("audit-url", viper.GetString("audit_url"), "audit URL")
	pflag.String("crypto-key", viper.GetString("crypto_key"), "crypto key path")
	pflag.StringP("config", "c", viper.GetString("config"), "path to configuration file")
	pflag.Parse()

	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		if f := pflag.Lookup("config"); f != nil {
			configPath = f.Value.String()
		}
	}

	if err := parseServerConfigFile(configPath); err != nil {
		return nil, err
	}

	bindPFlags()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	viper.BindEnv("address", "ADDRESS")
	viper.BindEnv("log_level", "LOG_LEVEL")
	viper.BindEnv("store_interval", "STORE_INTERVAL")
	viper.BindEnv("file_storage_path", "FILE_STORAGE_PATH")
	viper.BindEnv("restore", "RESTORE")
	viper.BindEnv("database_dsn", "DATABASE_DSN")
	viper.BindEnv("key", "KEY")
	viper.BindEnv("audit_file", "AUDIT_FILE")
	viper.BindEnv("audit_url", "AUDIT_URL")
	viper.BindEnv("crypto_key", "CRYPTO_KEY")
	viper.BindEnv("config", "CONFIG")

	var cfg = &ServerConfig{}

	if err := viper.Unmarshal(cfg); err != nil {
		return cfg, err
	}

	cfg.UseDatabaseAsStorage = (cfg.DatabaseDSN != "")

	cfg.StoreOnUpdate = (*cfg.StoreInterval == 0) && !cfg.UseDatabaseAsStorage

	cfg.StorePeriodically = (*cfg.StoreInterval != 0) && !cfg.UseDatabaseAsStorage

	*cfg.RestoreStorage = *cfg.RestoreStorage && !cfg.UseDatabaseAsStorage

	return cfg, nil

}

func parseServerConfigFile(configPath string) error {
	if configPath == "" {
		return nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var fileConfig FileServerConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fileConfig)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	if fileConfig.Host != "" {
		viper.Set("address", fileConfig.Host)
	}
	if fileConfig.LogLevel != "" {
		viper.Set("log_level", fileConfig.LogLevel)
	}
	if fileConfig.StoreInterval != "" {
		storeIntervalDuration, err := time.ParseDuration(fileConfig.StoreInterval)
		if err != nil {
			return fmt.Errorf("failed to parse storeInterval duration: %w", err)
		}
		viper.Set("store_interval", int(storeIntervalDuration.Seconds()))
	}

	if fileConfig.FileStoragePath != "" {
		viper.Set("file_storage_path", fileConfig.FileStoragePath)
	}

	if fileConfig.RestoreStorage != nil {
		viper.Set("restore", *fileConfig.RestoreStorage)
	}

	if fileConfig.DatabaseDSN != "" {
		viper.Set("database_dsn", fileConfig.DatabaseDSN)
	}

	if fileConfig.AuditFile != "" {
		viper.Set("audit_file", fileConfig.AuditFile)
	}

	if fileConfig.AuditURL != "" {
		viper.Set("audit_url", fileConfig.AuditURL)
	}

	if fileConfig.Key != "" {
		viper.Set("key", fileConfig.Key)
	}

	if fileConfig.CryptoKeyPath != "" {
		viper.Set("crypto_key", fileConfig.CryptoKeyPath)
	}

	return nil
}
