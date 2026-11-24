// Package config reads the marketplace server settings from the environment.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrMissingAPIKey is returned when MARKETPLACE_API_KEY is empty.
var ErrMissingAPIKey = errors.New("config: MARKETPLACE_API_KEY is not set")

// Config holds the server settings.
type Config struct {
	Addr            string
	DataPath        string
	APIKey          string
	ShutdownTimeout time.Duration
	Coupons         map[string]int
}

// FromEnv builds a Config from getenv, usually os.Getenv.
//
//	PORT                 listen port, default 8080
//	DATA_PATH            catalog file, default data/marketplace.json
//	MARKETPLACE_API_KEY  bearer token for write routes, required
//	SHUTDOWN_TIMEOUT     graceful shutdown limit, default 10s
//	COUPONS              CODE=PERCENT pairs, for example SPRING10=10,VIP=25
func FromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:            ":8080",
		DataPath:        "data/marketplace.json",
		ShutdownTimeout: 10 * time.Second,
	}

	if v := getenv("PORT"); v != "" {
		port, _ := strconv.Atoi(v)
		cfg.Addr = ":" + strconv.Itoa(port)
	}
	if v := getenv("DATA_PATH"); v != "" {
		cfg.DataPath = v
	}

	cfg.APIKey = getenv("MARKETPLACE_API_KEY")
	if cfg.APIKey == "" {
		return Config{}, ErrMissingAPIKey
	}

	if v := getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("config: SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = d
	}

	coupons, err := parseCoupons(getenv("COUPONS"))
	if err != nil {
		return Config{}, err
	}
	cfg.Coupons = coupons
	return cfg, nil
}

// parseCoupons reads comma-separated CODE=PERCENT pairs. Codes are stored
// upper case. Percents must be 1 to 100.
func parseCoupons(s string) (map[string]int, error) {
	out := map[string]int{}
	if strings.TrimSpace(s) == "" {
		return out, nil
	}
	for _, pair := range strings.Split(s, ",") {
		code, pct, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok || code == "" {
			return nil, fmt.Errorf("config: COUPONS: bad entry %q", pair)
		}
		n, err := strconv.Atoi(pct)
		if err != nil || n < 1 || n > 100 {
			return nil, fmt.Errorf("config: COUPONS: %s percent must be 1 to 100, got %q", code, pct)
		}
		out[strings.ToUpper(code)] = n
	}
	return out, nil
}
