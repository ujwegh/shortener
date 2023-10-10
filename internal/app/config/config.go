package config

import (
	"flag"
	"strconv"
	"strings"
)

type AppConfig struct {
	Host             string
	Port             int
	ServerAddr       string
	ShortenedURLAddr string
}

func ParseFlags() AppConfig {
	config := AppConfig{}
	var FlagServerAddr string
	flag.StringVar(&FlagServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.ShortenedURLAddr, "b", "localhost:8080", "address and port for shortened url")
	flag.Parse()

	config.ServerAddr = FlagServerAddr
	config.Host = strings.Split(FlagServerAddr, ":")[0]
	port, err := strconv.Atoi(strings.Split(FlagServerAddr, ":")[1])
	if err != nil {
		panic(err)
	}
	config.Port = port

	return config
}
