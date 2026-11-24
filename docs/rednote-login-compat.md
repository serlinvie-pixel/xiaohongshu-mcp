# REDnote login compatibility — experimental

This branch adds an **opt-in** `XHS_BACKEND=rednote` login flow. The original
`xiaohongshu.com` behavior remains unchanged when the variable is omitted.

## Supported by this patch

- QR acquisition opens `https://www.rednote.com/login` instead of the domestic
  `/explore` route, based on the reproducible investigation in upstream #812.
- Login status and scan confirmation also look for a non-guest identity in
  `window.__INITIAL_STATE__.user`, reducing false negatives (upstream #838).
- QR-element waiting is bounded to 45 seconds instead of hanging indefinitely.
- Saved cookies keep their original `.rednote.com` domain. Do not rewrite cookie domains.
- **Experimental persistent browser profile:** only with `XHS_BROWSER_PROFILE_DIR`, the MCP
  holds a *single* Chromium process open across tool calls, saving the site's cookies,
  localStorage, IndexedDB and other browser data in `/app/data/rednote-chromium`.
  Other MCP calls open separate tabs and close the tabs only. The domestic backend
  still uses a fresh temporary browser for each request.
- In persistent mode, `cookies.json` is still written for compatibility, but **is not
  re-imported on startup**; Chromium's private profile directory is the source of truth.
  The original `cookies.json` file remains intact. A fresh QR login may be required.

**Not yet verified after adding the persistent profile**: account sign-in on your deployment, feed search, comments,
notifications, publishing, and other functions. Those endpoints still contain
domestic assumptions and should be tested or adapted individually. Do not claim
full REDnote MCP compatibility from successful QR login alone.

## VPS rollout (noVNC)

These commands run **on the VPS**, not in Termux. Preserve the existing
`/root/xhs-novnc.env` file (contains a private AUTH_TOKEN), with data paths
pointing to the persistent `private-mcp/data` and `private-mcp/images` folders.

1. Keep a **private** backup of `/opt/xiaohongshu-mcp/deploy/private-mcp/data`
   before replacing the runtime; never publish browser profile files, login session
   files, screenshots of logged-in cookies, or your token.
   The Chrome profile may contain sensitive authenticated session data; ensure its
   parent directory and any backups are not readable by untrusted users.
2. Fetch this branch from the fork and check it out.
3. From `docker/novnc` build the local image (this uses Go 1.27.3 inside Docker):
   ```bash
   docker compose --env-file /root/xhs-novnc.env -f compose.yml -f compose.rednote.yml build
   ```
4. Recreate **only** the xiaohongshu-mcp service:
   ```bash
   docker compose --env-file /root/xhs-novnc.env -f compose.yml -f compose.rednote.yml up -d --no-deps --force-recreate xiaohongshu-mcp
   ```
   The REDnote overlay now activates `XHS_BROWSER_PROFILE_DIR=/app/data/rednote-chromium`.
   No additional Docker volume is required because `/app/data` is already mounted.
   **Do not start two MCP containers sharing the same profile directory.**
5. Keep port 6080 bound to 127.0.0.1 and use SSH local forwarding to
   inspect the GUI. Do not expose the unauthenticated noVNC desktop publicly.
6. Existing cookies from the previous ephemeral browser cannot recreate
   missing localStorage/IndexedDB; a **single fresh QR login** may be needed after
   enabling profile mode. Complete any ordinary account verification.
   Check the login-status tool in a new tab and after one container restart
   (without taking a new QR). A login-status result is not enough on its own;
   confirm the genuine account identity before using account-modifying tools.
   Do not print cookie values, storage contents or passwords.
7. An open Chromium process now persists in the background. noVNC can still be
   blank between calls because the *tab* closes. More background RAM use is
   expected. `delete_cookies` deliberately refuses to pretend it can erase the
   full profile while Chromium is live. To log out fully, stop only this service,
   back up then separately remove its private profile and cookies; do not do
   that during troubleshooting.

## Roll back

To restore the previously working official noVNC image without deleting data,
from `docker/novnc` run:

```bash
docker compose --env-file /root/xhs-novnc.env -f compose.yml up -d --no-deps --force-recreate xiaohongshu-mcp
```

This switches back to the official image, **without deleting** the new profile,
previous cookie file, or bind-mounted data. It does not repair an invalidated
REDnote session automatically. If the browser fails with a profile lock error,
stop any *other* process using that profile and diagnose; do not delete profile
files while Chrome is running.

The noVNC image is a separate container and does not modify Ombre Brain or
Xinchao. The user is responsible for restoring a working session if the platform
invalidates any login cookies.

References: https://github.com/xpzouying/xiaohongshu-mcp/issues/812 and
https://github.com/xpzouying/xiaohongshu-mcp/issues/838 .
