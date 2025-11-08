package main

import (
	"crypto/tls"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	_ "layeh.com/gumble/opus"
)

type Config struct {
	MumbleAddress  string
	MumbleUsername string
	MumblePassword string
	StreamUrl      string
}

func LoadConfig() Config {
	ma, maP := os.LookupEnv("MUMBLE_ADDRESS")
	if !maP {
		log.Fatal("MUMBLE_ADDRESS env var must be set")
	}

	mu, muP := os.LookupEnv("MUMBLE_USERNAME")
	if !muP {
		slog.Warn("MUMBLE_USERNAME is empty, using default StreamEcho username")
		mu = "StreamEcho"
	}

	mp, mpP := os.LookupEnv("MUMBLE_PASSWORD")
	if !mpP {
		slog.Warn("MUMBLE_PASSWORD is empty, using default empty password")
		mp = ""
	}

	su, suP := os.LookupEnv("STREAM_URL")
	if !suP {
		log.Fatal("STREAM_URL env var must be set")
	}

	return Config{
		MumbleAddress:  ma,
		MumbleUsername: mu,
		MumblePassword: mp,
		StreamUrl:      su,
	}
}

func Connect(c *gumble.Config, addr string, streamUrl string) {
	client, err := GumbleDialInsecure(addr, c)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Connected to the Mumble server", "address", addr, "username", c.Username)

	for {
		source := gumbleffmpeg.SourceFile(streamUrl)
		stream := gumbleffmpeg.New(client, source)

		if err := stream.Play(); err != nil {
			slog.Error("Failed playing the stream", "error", err)
		}

		stream.Wait()
		slog.Warn("Stream stopped, reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

// Same as gumble.Dial() but with InsecureSkipVerify set to true
func GumbleDialInsecure(addr string, config *gumble.Config) (*gumble.Client, error) {
	return gumble.DialWithDialer(new(net.Dialer), addr, config, &tls.Config{
		InsecureSkipVerify: true,
	})
}

func main() {
	c := LoadConfig()
	gc := gumble.NewConfig()
	gc.Username = c.MumbleUsername
	gc.Password = c.MumblePassword
	Connect(gc, c.MumbleAddress, c.StreamUrl)
}
