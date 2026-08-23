package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DefaultShutdownTimeout 是未配置环境变量时使用的关停超时。
// 取非零值，避免关停被单个慢请求无限期挂住、拖累重启。
const DefaultShutdownTimeout = 10 * time.Second

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
}

func Load() Config {
	p := 8080
	if v, e := strconv.Atoi(os.Getenv("PORT")); e == nil && v > 0 && v < 65536 {
		p = v
	}
	shutdownTimeout := DefaultShutdownTimeout
	if v, e := strconv.Atoi(os.Getenv("SHUTDOWN_TIMEOUT_MS")); e == nil && v >= 0 {
		shutdownTimeout = time.Duration(v) * time.Millisecond
	}
	return Config{Port: p, ShutdownTimeout: shutdownTimeout}
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
