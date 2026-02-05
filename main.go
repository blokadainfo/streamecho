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
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

type Config struct {
	MumbleAddress  string
	MumbleUsername string
	MumblePassword string
	MumbleChannel  string
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

	mc, mcP := os.LookupEnv("MUMBLE_CHANNEL")
	if !mcP {
		slog.Warn("MUMBLE_CHANNEL is empty, using default root channel")
		mc = ""
	}

	su, suP := os.LookupEnv("STREAM_URL")
	if !suP {
		log.Fatal("STREAM_URL env var must be set")
	}

	return Config{
		MumbleAddress:  ma,
		MumbleUsername: mu,
		MumblePassword: mp,
		MumbleChannel:  mc,
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
	gc.Attach(&gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			if c.MumbleChannel == "" {
				slog.Info("Joined default channel")
				return
			}

			ch := e.Client.Channels.Find(c.MumbleChannel)
			if ch == nil {
				slog.Warn("Channel not found", "name", c.MumbleChannel)
				return
			}
			e.Client.Self.Move(ch)
			slog.Info("Joined channel", "name", ch.Name)
		},
	})
	gc.Attach(&gumbleutil.Listener{
		Disconnect: func(e *gumble.DisconnectEvent) {
			slog.Error("Disconnected from the server, exiting...")
			os.Exit(1)
		},
	})
	Connect(gc, c.MumbleAddress, c.StreamUrl)
}
