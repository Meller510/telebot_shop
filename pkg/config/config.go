package config

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.AddConfigPath("configs")
	viper.AddConfigPath("assets/label")
	viper.AddConfigPath(".")
}

type Config struct {
	PathImage     string
	Support       string
	TelegramToken string
	Email         Mail
	DataBase      DB
}

type DB struct {
	Username string
	Password string
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DBname   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Mail struct {
	Address  string
	Password string
	Initial  string `mapstructure:"initial"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

func InitConfig() (*Config, error) {
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, errReadConfig
	}

	var cfg Config

	if err := viper.UnmarshalKey("pathImage", &cfg.PathImage); err != nil || len(cfg.PathImage) == 0 {
		return nil, errImagePath
	}

	if err := viper.UnmarshalKey("support", &cfg.Support); err != nil {
		return nil, errSupport
	}

	if err := viper.UnmarshalKey("mail", &cfg.Email); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("db", &cfg.DataBase); err != nil {
		return nil, err
	}

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func parseEnv(cfg *Config) error {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		if !errors.Is(err, &viper.ConfigFileNotFoundError{}) {
			logrus.Error(errFoundEnvFile)
		}
	}

	if err := viper.BindEnv("TOKEN"); err != nil {
		return err
	}

	if err := viper.BindEnv("DB_USERNAME"); err != nil {
		return err
	}

	if err := viper.BindEnv("DB_PASSWORD"); err != nil {
		return err
	}

	if err := viper.BindEnv("MAIL_ADDRESS"); err != nil {
		return err
	}

	if err := viper.BindEnv("MAIL_PASSWORD"); err != nil {
		return err
	}

	cfg.TelegramToken = viper.GetString("TOKEN")
	cfg.DataBase.Username = viper.GetString("DB_USERNAME")
	cfg.DataBase.Password = viper.GetString("DB_PASSWORD")
	cfg.Email.Address = viper.GetString("MAIL_ADDRESS")
	cfg.Email.Password = viper.GetString("MAIL_PASSWORD")

	return checkENV(cfg)
}
