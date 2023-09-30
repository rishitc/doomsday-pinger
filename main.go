package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

const cmdLineIndexPingTargetsFileName = 1

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) < 2 {
		logger.Error("invalid number of arguments")
	}
	fileName := os.Args[cmdLineIndexPingTargetsFileName]
	data, err := readFile(fileName)
	if err != nil {
		logger.Error("Unable to open file", fileName, err)
	}

	targets, err := parse(data)
	if err != nil {
		logger.Error("Unable to open file", fileName, err)
	}
	for _, target := range targets {
		if ok := ping(target); !ok {
			logger.Error(fmt.Sprintf("Ping failed for %s", target))
		} else {
			logger.Info(fmt.Sprintf("Ping succeeded for %s", target))
		}
	}
}

func readFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parse(data string) ([]string, error) {
	type Targets struct {
		IPs []string
	}
	var conf Targets
	if _, err := toml.Decode(string(data), &conf); err != nil {
		return []string{}, err
	}
	return conf.IPs, nil
}

func ping(target string) bool {
	return true
}
