package xiaohongshu

import (
	"os"
	"strings"

	"github.com/go-rod/rod"
)

// rednote is opt-in: the original xiaohongshu.com behavior stays the default.
// Cookies are origin-scoped; do not change their domain to work around redirects.
func useRednote() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("XHS_BACKEND")), "rednote")
}

func loginExploreURL() string {
	if useRednote() {
		return "https://www.rednote.com/explore"
	}
	return "https://www.xiaohongshu.com/explore"
}

func loginQRCodeURL() string {
	if useRednote() {
		// REDnote does not automatically open the QR modal on /explore.
		return "https://www.rednote.com/login"
	}
	return loginExploreURL()
}

// A sidebar element alone is not reliable on REDnote. Require a non-guest
// identity in the page's state. For the domestic site, preserve the existing
// selector as a fast path.
func checkPageAuthenticated(page *rod.Page) (bool, error) {
	exists, _, err := page.Has(`.main-container .user .link-wrapper .channel`)
	if err != nil {
		return false, err
	}
	if exists && !useRednote() {
		return true, nil
	}

	res, err := page.Eval(`() => {
		const unwrap = value => {
			if (!value || typeof value !== "object") return value;
			return value.value !== undefined ? value.value :
				(value._value !== undefined ? value._value : value);
		};
		const root = window.__INITIAL_STATE__;
		const user = root && root.user;
		const candidates = user ? [user.userInfo, user.currentUser, user.me] : [];
		const info = candidates.map(unwrap).find(value =>
			value && typeof value === "object" && value.guest !== true &&
			Boolean(value.userId || value.user_id || value.nickname));
		return Boolean(info);
	}`)
	if err != nil {
		return false, err
	}
	return res.Value.Bool(), nil
}


// rednoteSessionCookiePresent is a scan-only fallback for REDnote's login
// redirect, which sometimes fails to populate the old sidebar or page-state
// identity before the login popup closes. No cookie values are ever logged.
// This must not be used as a general login-status check: cookies can be stale.
func rednoteSessionCookiePresent(page *rod.Page) (bool, error) {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return false, err
	}
	hasSession, hasIdentity := false, false
	for _, c := range cks {
		if c == nil || c.Value == "" {
			continue
		}
		domain := strings.TrimPrefix(strings.ToLower(c.Domain), ".")
		if domain != "rednote.com" && !strings.HasSuffix(domain, ".rednote.com") {
			continue
		}
		switch c.Name {
		case "web_session":
			hasSession = true
		case "id_token":
			hasIdentity = true
		}
	}
	return hasSession && hasIdentity, nil
}


// Safe login diagnostic: only boolean page state and cookie presence.
// Never log cookie values, URL query strings, nicknames, or user IDs.
func rednoteLoginDiagnostics(page *rod.Page) string {
	res, err := page.Eval(`() => {
		const unwrap = value => {
			if (!value || typeof value !== "object") return value;
			return value.value !== undefined ? value.value :
				(value._value !== undefined ? value._value : value);
		};
		const user = window.__INITIAL_STATE__ && window.__INITIAL_STATE__.user;
		const candidates = user ? [user.userInfo, user.currentUser, user.me] : [];
		const infos = candidates.map(unwrap).filter(value =>
			value && typeof value === "object");
		const hasIdentity = infos.some(value => value.guest !== true &&
			Boolean(value.userId || value.user_id || value.nickname));
		const guest = infos.some(value => value.guest === true);
		return JSON.stringify({
			rednote_origin: location.hostname === "rednote.com" ||
				location.hostname.endsWith(".rednote.com"),
			login_path: location.pathname === "/login",
			explore_path: location.pathname === "/explore",
			sidebar_present: Boolean(document.querySelector(".main-container .user .link-wrapper .channel")),
			login_button_present: Boolean(document.querySelector("#login-btn")),
			user_state_present: Boolean(user),
			identity_present: hasIdentity,
			guest_present: guest
		});
	}`)
	if err != nil {
		return "page_state_unavailable"
	}
	return res.Value.String()
}

// Only the presence of named cookies is returned, never their values.
func rednoteCookiePresence(page *rod.Page) (bool, bool, error) {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return false, false, err
	}
	hasSession, hasID := false, false
	for _, c := range cks {
		if c == nil || c.Value == "" {
			continue
		}
		domain := strings.TrimPrefix(strings.ToLower(c.Domain), ".")
		if domain != "rednote.com" && !strings.HasSuffix(domain, ".rednote.com") {
			continue
		}
		if c.Name == "web_session" {
			hasSession = true
		}
		if c.Name == "id_token" {
			hasID = true
		}
	}
	return hasSession, hasID, nil
}
