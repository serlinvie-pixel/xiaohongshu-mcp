# Private VPS deployment (Cloudflare Tunnel)

This deployment profile keeps the MCP service off the public network:

- The host port is bound to `127.0.0.1`, not `0.0.0.0`.
- A non-empty `AUTH_TOKEN` is mandatory.
- Login cookies and runtime state persist in `data/`.
- Publishing media is mounted read-only from `images/`.
- The image is pinned to upstream release `v2.7.4`; update it deliberately after reviewing release notes.

It does not change the upstream `docker/docker-compose.yml`.

## 1. Check the server and Docker

The upstream Docker build currently targets Linux x86-64. On the VPS, check:

```bash
uname -m
docker --version
docker compose version
```

For this pinned image, `uname -m` should report `x86_64`. If it reports `aarch64` or another architecture, stop here and verify an image for that architecture before deploying.

## 2. Prepare the environment

Clone your fork, then enter this folder:

```bash
git clone --branch deploy/private-mcp https://github.com/serlinvie-pixel/xiaohongshu-mcp.git
cd xiaohongshu-mcp/deploy/private-mcp
cp .env.example .env
openssl rand -hex 32
```

Copy the generated random value into the `AUTH_TOKEN` line in `.env`. Keep the token private. Then lock down the file and create the persistent directories:

```bash
chmod 600 .env
mkdir -p data images
```

Do not put cookies, the real `.env`, or other runtime data into Git.

## 3. Validate and start

Before starting on a small VPS, check the resources already used by other services:

```bash
free -h
df -h
docker stats --no-stream
```

This service starts Chromium. If the VPS is already close to its memory limit, do not start it yet; add memory/swap headroom or move the service to a separate host first.

Run these commands from `deploy/private-mcp/`:

```bash
docker compose config --quiet
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100 xiaohongshu-mcp
```

The configuration check fails if `AUTH_TOKEN` is missing or empty. The service is published only on host loopback at `127.0.0.1:18060`.

Verify the bind address:

```bash
ss -lnt | grep ':18060'
```

The listener should show `127.0.0.1:18060`, never `0.0.0.0:18060` or `[::]:18060`. Do not open TCP port 18060 in the VPS firewall or security group.

## 4. Connect Cloudflare Tunnel

Choose the origin address based on where `cloudflared` runs.

### cloudflared installed on the VPS host

Set the tunnel's service/origin to:

```text
http://127.0.0.1:18060
```

### cloudflared runs in Docker

A container's `127.0.0.1` is its own loopback, not the VPS host or this MCP container. Connect the cloudflared container to the Docker network created above:

```bash
docker network connect xiaohongshu-mcp-net <cloudflared-container-name>
```

Then set the tunnel origin to:

```text
http://xiaohongshu-mcp:18060
```

If you changed `XHS_DOCKER_NETWORK` in `.env`, use that network name in the `docker network connect` command. Keep the MCP container and cloudflared container on the same Docker network.

The public hostname should forward to the MCP origin without rewriting the `/mcp` path. Keep application authentication enabled; a Cloudflare Tunnel is not a substitute for the MCP bearer token. A Cloudflare Access policy can be added as an additional access layer if it fits the client.

## 5. Configure an MCP client

Use the public tunnel URL with the MCP path:

```text
https://<your-tunnel-hostname>/mcp
```

The client must support a custom HTTP header and send:

```text
Authorization: Bearer <the-value-of-AUTH_TOKEN>
```

Do not paste the real token into GitHub, a public issue, or an untrusted client. If your chosen client cannot send custom headers, do not remove authentication just to make it connect; use a trusted local proxy that injects the bearer token, or choose a client that supports request headers.

After connecting, use the MCP login/QR-code tool exposed by the server and scan it with the Xiaohongshu mobile app. The login state is stored under `data/` and survives container replacement.

## 6. Back up and update safely

The `data/` directory contains the login cookie and runtime state. Treat it as a credential store. Back it up securely and never upload it to GitHub.

To update, change `XHS_IMAGE` in `.env` to a reviewed release tag, then run:

```bash
docker compose pull
docker compose up -d
```

The bind-mounted `data/` and `images/` directories are retained. Do not use `docker compose down -v` as part of routine updates.

To inspect logs:

```bash
docker compose logs -f --tail=100 xiaohongshu-mcp
```
