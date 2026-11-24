package config

import (
	"errors"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvDefaults(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{"MARKETPLACE_API_KEY": "k"}))
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.Addr != ":8080" || cfg.DataPath != "data/marketplace.json" || cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("FromEnv = %+v", cfg)
	}
	if len(cfg.Coupons) != 0 {
		t.Errorf("Coupons = %v, want empty", cfg.Coupons)
	}
}

func TestFromEnvOverrides(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"MARKETPLACE_API_KEY": "k",
		"PORT":                "9090",
		"DATA_PATH":           "/srv/catalog.json",
		"SHUTDOWN_TIMEOUT":    "3s",
		"COUPONS":             "spring10=10, VIP=25",
	}))
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.Addr != ":9090" || cfg.DataPath != "/srv/catalog.json" || cfg.ShutdownTimeout != 3*time.Second {
		t.Errorf("FromEnv = %+v", cfg)
	}
	if cfg.Coupons["SPRING10"] != 10 || cfg.Coupons["VIP"] != 25 {
		t.Errorf("Coupons = %v", cfg.Coupons)
	}
}

func TestFromEnvMissingKey(t *testing.T) {
	if _, err := FromEnv(env(nil)); !errors.Is(err, ErrMissingAPIKey) {
		t.Errorf("FromEnv error = %v, want ErrMissingAPIKey", err)
	}
}

func TestFromEnvBadValues(t *testing.T) {
	for name, m := range map[string]map[string]string{
		"timeout":       {"SHUTDOWN_TIMEOUT": "soon"},
		"coupon no =":   {"COUPONS": "SPRING10"},
		"coupon 0":      {"COUPONS": "FREE=0"},
		"coupon 150":    {"COUPONS": "WOW=150"},
		"coupon no pct": {"COUPONS": "A=x"},
	} {
		m["MARKETPLACE_API_KEY"] = "k"
		if _, err := FromEnv(env(m)); err == nil {
			t.Errorf("%s: FromEnv returned nil error", name)
		}
	}
}
