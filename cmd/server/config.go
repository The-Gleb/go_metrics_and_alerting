package main

import "github.com/num30/config"

type Config struct {
	Address         string `default:":8080" flag:"a" envvar:"ADDRESS"`
	LogLevel        string `default:"info" flag:"l" envvar:"LOGLEVEL"`
	FileStoragePath string `default:"" flag:"f" envvar:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `defaul:"" flag:"d" envvar:"DATABASE_DSN"`
	SignKey         string `default:"secret" flag:"k" envvar:"KEY"`
	StoreInterval   int    `default:"300" flag:"i" envvar:"STORE_INTERVAL"`
	Restore         bool   `default:"true" flag:"r"`
	PrivateKeyPath  string `default:"./private.pem" envvar:"PRIVATE_KEY_PATH"`
}

func MustBuildConfig(cfgFile string) *Config {
	var conf Config
	err := config.NewConfReader(cfgFile).Read(&conf)
	if err != nil {
		panic(err)
	}
	return &conf
}
