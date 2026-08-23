package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
}

func Load() Config {
	p := 8080
	if v, e := strconv.Atoi(os.Getenv("PORT")); e == nil && v > 0 && v < 65536 {
		p = v
	}
	shutdown := 10 * time.Second
	if v, e := strconv.Atoi(os.Getenv("SHUTDOWN_TIMEOUT_MS")); e == nil && v > 0 {
		shutdown = time.Duration(v) * time.Millisecond
	}
	return Config{Port: p, ShutdownTimeout: shutdown}
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
