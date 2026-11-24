package main

import "testing"

func TestPersistentProfileOptIn(t *testing.T) {
	t.Setenv("XHS_BACKEND", "rednote")
	t.Setenv("XHS_BROWSER_PROFILE_DIR", "/app/data/rednote-chromium")
	if got := persistentProfileDir(); got != "/app/data/rednote-chromium" {
		t.Fatalf("REDnote persistent profile not enabled: %q", got)
	}
}

func TestPersistentProfileDisabledWithoutPath(t *testing.T) {
	t.Setenv("XHS_BACKEND", "rednote")
	t.Setenv("XHS_BROWSER_PROFILE_DIR", "")
	if got := persistentProfileDir(); got != "" {
		t.Fatalf("profile should be disabled without path: %q", got)
	}
}

func TestPersistentProfileDisabledForDomestic(t *testing.T) {
	t.Setenv("XHS_BACKEND", "")
	t.Setenv("XHS_BROWSER_PROFILE_DIR", "/app/data/rednote-chromium")
	if got := persistentProfileDir(); got != "" {
		t.Fatalf("profile should not apply to the domestic backend: %q", got)
	}
}
