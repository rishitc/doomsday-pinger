package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/BurntSushi/toml"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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
		if ok := ping(logger, target); !ok {
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

var pingPayload = []byte("HELLO-R-U-THERE")

const location = "location"

func ping(logger *slog.Logger, target string) bool {
	const logPrefix = "ping"
	c, err := icmp.ListenPacket("udp4", "")
	if err != nil {
		logger.Error(err.Error(), slog.String(location, logPrefix))
		return false
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: pingPayload,
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		logger.Error(err.Error(), slog.String(location, logPrefix))
	}
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP(target), Port: 80}); err != nil {
		logger.Error(err.Error(), slog.String(location, logPrefix))
	}

	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		logger.Error(err.Error(), slog.String(location, logPrefix))
	}
	rm, err := icmp.ParseMessage(58, rb[:n])
	if err != nil {
		logger.Error(err.Error(), slog.String(location, logPrefix))
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		logger.Info(fmt.Sprintf("got reflection from %v", peer), slog.String(location, logPrefix))
	default:
		logger.Info(fmt.Sprintf("got %+v; want echo reply", rm), slog.String(location, logPrefix))
		b, _ := rm.Body.Marshal(4)
		logger.Info(string(b[4:]))

	}
	return true
}
