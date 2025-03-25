package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	//StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc" env-required:"true"`
	PgDb     DBConfig      `yaml:"postgres" env-required:"true"`
	HTTPConf HTTPConfig    `yaml:"http_server" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}
type DBConfig struct {
	Host     string `yaml:"dbHost" env-required:"true"`
	Port     int    `yaml:"dbPort" env-required:"true"`
	Username string `yaml:"dbUser" env-required:"true"`
	Password string `yaml:"dbPassword" env-required:"true"`
	Database string `yaml:"dbName" env-required:"true"`
	SSLMode  string `yaml:"dbSSLMode" env-required:"true"`
}

type HTTPConfig struct {
	Address     int           `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not exist " + path)
	}

	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("failed to read config " + err.Error())
	}

	return &config
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}
	return res
}
