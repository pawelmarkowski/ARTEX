# SysAIT Isolated ARTEX Hardening (`PLAT-INT-2` / `PLAT-26`)

Hardened fork configuration for running ARTEX as an air-gapped security validation worker in Kubernetes (`artex-system` namespace on `cf-lcj11-a`, PodSecurity `restricted`, `readOnlyRootFilesystem: true`).

## 1. Container & PodSecurity `restricted` Runtime (`Dockerfile`, `start.sh`, `cmd/artex/main.go`)
- Runs strictly non-root as `artex` (`UID 1000`, `GID 1000`) with `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`, `seccompProfile: RuntimeDefault`, and `find / -xdev -perm /6000 -exec chmod a-s {} +`.
- `readOnlyRootFilesystem: true` compatible:
  - Writable mounts expected at `/var/lib/artex` (`emptyDir`), `/app/keys` (`emptyDir`), and `/tmp` (`emptyDir`).
  - `cmd/artex/main.go` and `start.sh` initialize `-data`, `$HOME`, and `$HOME/.cache` (`0700`) before starting child tools (run with `-exec-uid=-1`).
  - Playwright browsers live under `/opt/ms-playwright` (`ENV PLAYWRIGHT_BROWSERS_PATH=/opt/ms-playwright`, world-readable `0755`) rather than `/root/.cache`.
  - Preinstalled BBOT module dependencies live in `/opt/bbot-home` (`0755`) and are symlinked into `$HOME/.bbot/tools` on startup by `start.sh`.
- `/app/hardening/` (`intercept-rules.sql`, `apply-intercept-rules.sh`) is baked into the image for database seeding.

## 2. Stripped Offensive / L2 / Brute-Force Tooling
- **Removed from `tools` stage:** `subfinder` (external DNS enumeration).
- **Removed from Debian runtime:** `arp-scan`, `arping`, `fping`, `netdiscover`, `nfs-common`, `python3-impacket`, `python3-paramiko`, `python3-scapy`, `rpcbind`, `hashcat`, `hydra`, `john`, `ncrack`, `openssh-client`, `ssh-audit`, `sshpass`, `inetutils-telnet`, `whois`, `nmap`, `vim`, `iputils-ping`.
- **Retained web/API verification toolkit:** `nuclei` (+ baked templates at `/opt/nuclei-templates`, `DISABLE_NUCLEI_TEMPLATES_PUBLIC_DOWNLOAD=true`), `httpx`, `katana`, `ffuf`, `dalfox`, `sqlmap`, `bbot`, `@playwright/mcp` / `playwright-cli` / Chromium, `curl`, `wget`, `jq`, `ripgrep`, `git`, `python3`, `node`.
- Updated `skills/bbot-recon/SKILL.md` to remove `subfinder` and `nmap` instructions.

## 3. Exfiltration, Webhook & Telemetry Lockdown
- **Update check & prompt boundary defaults:** `ENV ARTEX_UPDATE_CHECK=0 ARTEX_UNTRUSTED_BOUNDARY=1`.
- **External notification & chat-bridge kill switch (`server/external_channels_gate.go`):**
  - `ENV ARTEX_DISABLE_NOTIFY=1` disables `Notifier.Run`, `Notifier.enabled()`, and `chatBridgeRunner.Run`.
  - Mutating/outbound API handlers (`/api/notify/channels*`, `/api/notify/deliveries/*/retry`, `/api/chat-bridges/*`, and `PUT /api/settings` with `notify_enabled: true`) fail closed with HTTP `403 Forbidden`.
  - `GET /api/notify/meta` reports `"enabled": false`, `"disabled_by_policy": true`, and `"kinds": []`.
- **Upstream GitHub bot removed:** Deleted `.github/workflows/bot.yml`, `.github/scripts/bot.sh`, and `scripts/test-bot-context.sh`.

## 4. CI & GHCR Container Publishing (`.github/workflows/ci.yml`)
- Runs Go checks (`gofmt`, `go build`, `go vet`, `go test -p 1 ./...`, `go test -race`), `govulncheck`, `shellcheck`, frontend checks/build, and Python benchmark tests.
- On push to `main` or `feat/**` (and `workflow_dispatch`), builds `linux/amd64` container images and pushes to `ghcr.io/pawelmarkowski/artex`.
