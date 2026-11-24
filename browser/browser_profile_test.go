package browser

import "testing"

func TestWithProfileDirOption(t *testing.T) {
	var cfg browserConfig
	WithProfileDir("/app/data/rednote-chromium")(&cfg)
	if cfg.profileDir != "/app/data/rednote-chromium" {
		t.Fatalf("profileDir = %q", cfg.profileDir)
	}
}
