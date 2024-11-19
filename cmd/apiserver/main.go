package main

import (
	"flag"
	"log"

	"github.com/Olzheke2003/GoRestApi/internal/app/apiserver"

	"github.com/BurntSushi/toml"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "configs/apiserver.toml", "path to config file")
}

func main() {
	flag.Parse()

	// Создаем конфигурацию
	config := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем новый API сервер
	server := apiserver.New(config)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
