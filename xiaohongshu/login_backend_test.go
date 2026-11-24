package xiaohongshu

import "testing"

func TestLoginTargetsDefault(t *testing.T) {
	t.Setenv("XHS_BACKEND", "")
	if got := loginExploreURL(); got != "https://www.xiaohongshu.com/explore" {
		t.Fatalf("domestic explore URL: %q", got)
	}
	if got := loginQRCodeURL(); got != loginExploreURL() {
		t.Fatalf("domestic QR URL: %q", got)
	}
}

func TestLoginTargetsRednote(t *testing.T) {
	t.Setenv("XHS_BACKEND", "rednote")
	if got := loginExploreURL(); got != "https://www.rednote.com/explore" {
		t.Fatalf("REDnote explore URL: %q", got)
	}
	if got := loginQRCodeURL(); got != "https://www.rednote.com/login" {
		t.Fatalf("REDnote QR URL: %q", got)
	}
}

func TestUnknownBackendKeepsDomesticDefault(t *testing.T) {
	t.Setenv("XHS_BACKEND", "unknown")
	if useRednote() {
		t.Fatal("unknown backend must not enable REDnote")
	}
}
