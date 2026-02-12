package main

import (
	"crypto/tls"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

type Config struct {
	MumbleAddress     string
	MumbleUsername    string
	MumblePassword    string
	MumbleChannel     string
	MumbleChannelLock bool
	StreamUrl         string
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

	mcl, mclP := os.LookupEnv("MUMBLE_CHANNEL_LOCK")
	if !mclP {
		slog.Warn("MUMBLE_CHANNEL_LOCK is not set, defaulting to false")
		mcl = "false"
	}
	mclV, err := strconv.ParseBool(mcl)
	if err != nil {
		log.Fatal("MUMBLE_CHANNEL_LOCK must be a boolean")
	}

	su, suP := os.LookupEnv("STREAM_URL")
	if !suP {
		log.Fatal("STREAM_URL env var must be set")
	}

	return Config{
		MumbleAddress:     ma,
		MumbleUsername:    mu,
		MumblePassword:    mp,
		MumbleChannel:     mc,
		MumbleChannelLock: mclV,
		StreamUrl:         su,
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
				if c.MumbleChannelLock {
					slog.Warn("Channel lock (whisper feature) will not work")
				}
			} else if ch := e.Client.Channels.Find(c.MumbleChannel); ch != nil {
				e.Client.Self.Move(ch)
				slog.Info("Joined channel", "name", ch.Name)

				if c.MumbleChannelLock {
					vt := &gumble.VoiceTarget{ID: 1}
					vt.AddChannel(ch, false, false, "")
					e.Client.Send(vt)
					e.Client.VoiceTarget = vt
					slog.Info("Enabled channel lock (whisper feature)", "name", ch.Name)
				}
			} else {
				slog.Warn("Channel not found", "name", c.MumbleChannel)
				if c.MumbleChannelLock {
					slog.Warn("Channel lock (whisper feature) will not work")
				}
			}
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
