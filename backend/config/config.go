package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const DefaultShutdownTimeout = 0 * time.Second

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
}

func Load() Config {
	p := 8080
	if v, e := strconv.Atoi(os.Getenv("PORT")); e == nil && v > 0 && v < 65536 {
		p = v
	}
	return Config{Port: p}
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
