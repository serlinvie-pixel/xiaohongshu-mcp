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

**Not yet verified**: account sign-in on your deployment, feed search, comments,
notifications, publishing, and other functions. Those endpoints still contain
domestic assumptions and should be tested or adapted individually. Do not claim
full REDnote MCP compatibility from successful QR login alone.

## VPS rollout (noVNC)

These commands run **on the VPS**, not in Termux. Preserve the existing
`/root/xhs-novnc.env` file (contains a private AUTH_TOKEN), with data paths
pointing to the persistent `private-mcp/data` and `private-mcp/images` folders.

1. Keep a private backup before replacing the runtime; never publish session
   files, screenshots of logged-in cookies, or your token.
2. Fetch this branch from the fork and check it out.
3. From `docker/novnc` build the local image (this uses Go 1.27.3 inside Docker):
   ```bash
   docker compose --env-file /root/xhs-novnc.env -f compose.yml -f compose.rednote.yml build
   ```
4. Recreate only the xiaohongshu-mcp service:
   ```bash
   docker compose --env-file /root/xhs-novnc.env -f compose.yml -f compose.rednote.yml up -d
   ```
5. Keep port 6080 bound to 127.0.0.1 and use SSH local forwarding to
   inspect the GUI. Do not expose the unauthenticated noVNC desktop publicly.
6. Test a **single** QR login and inspect logs for
   `扫码登录成功，cookies 已保存`. Check only the file size/timestamp of the
   persistent `cookies.json`; never print its contents.

## Roll back

To restore the previously working official noVNC image without deleting data,
from `docker/novnc` run:

```bash
docker compose --env-file /root/xhs-novnc.env -f compose.yml up -d --force-recreate
```

The noVNC image is a separate container and does not modify Ombre Brain or
Xinchao. The user is responsible for restoring a working session if the platform
invalidates any login cookies.

References: https://github.com/xpzouying/xiaohongshu-mcp/issues/812 and
https://github.com/xpzouying/xiaohongshu-mcp/issues/838 .
