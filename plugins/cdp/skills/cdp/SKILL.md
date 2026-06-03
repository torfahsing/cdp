---
name: cdp
description: Use when interacting with a browser — inspecting pages, clicking elements, filling forms, taking screenshots, evaluating JavaScript, reviewing Figma designs, or any Chrome DevTools task. Replaces chrome-devtools-mcp with compact Bash output that saves context tokens.
---

# cdp — Chrome DevTools Protocol CLI

CLI for controlling Chrome from the terminal. Use `cdp` commands via Bash instead of MCP tool calls — produces compact output with minimal context overhead.

**Windows:** Use `cdp.exe` instead of `cdp` in all commands below.

## Setup

Chrome must have remote debugging enabled. Either:
- Launch with `--remote-debugging-port=9222`
- Enable via `chrome://inspect/#remote-debugging` (Chrome 144+)

Verify: `cdp pages` (or `cdp.exe pages` on Windows) should list open tabs.

## Core Workflow

**snapshot → interact → verify**

```bash
cdp snapshot                    # get element UIDs
cdp click a3                    # interact by UID
cdp snapshot                    # verify result
```

UIDs (like `a1`, `a2`) come from `cdp snapshot` and are used by all interaction commands. They're valid until the next `snapshot` call.

## Quick Reference

| Task | Command |
|------|---------|
| List tabs | `cdp pages` |
| Select tab | `cdp select <id>` |
| Navigate | `cdp navigate <url>` |
| Go back/forward | `cdp navigate --back` / `--forward` |
| Reload | `cdp navigate --reload` |
| Open new tab | `cdp open <url>` |
| A11y snapshot | `cdp snapshot` |
| Screenshot to file | `cdp screenshot --file shot.png` |
| Full page screenshot | `cdp screenshot --file full.png --full-page` |
| Run JavaScript | `cdp eval "document.title"` |
| Click element | `cdp click <uid>` |
| Double click | `cdp click <uid> --dbl` |
| Fill input | `cdp fill <uid> "text"` |
| Fill multiple | `cdp fill-form --field a1=val --field a2=val` |
| Type text | `cdp type "hello" --submit Enter` |
| Press key | `cdp key "Control+A"` |
| Hover | `cdp hover <uid>` |
| Wait for text | `cdp wait "Loading complete"` |
| Handle dialog | `cdp dialog accept` |
| Upload file | `cdp upload <uid> /path/to/file` |
| Resize viewport | `cdp resize 1920 1080` |
| Emulate mobile | `cdp emulate --viewport 375x812x3` |
| Dark mode | `cdp emulate --color-scheme dark` |
| Network requests | `cdp network` |
| Console messages | `cdp console` |
| Perf trace | `cdp perf start` / `cdp perf stop --file trace.json` |
| Heap snapshot | `cdp memory heap.heapsnapshot` |
| Print version | `cdp version` |

All commands accept `--json` for structured output, `--port` to override Chrome port, `--timeout` for custom timeout.

## Figma Workflow

```bash
cdp navigate "https://www.figma.com/design/..."
cdp snapshot                    # find layer names
cdp click a15                   # select a layer
cdp key "Shift+1"               # zoom to fit selection
cdp key "Shift+2"               # zoom tight to selection
cdp screenshot --file design.png
```

**Note:** Strip `&m=dev` from Figma URLs before navigating. Dev Mode triggers an access-request dialog that blocks the canvas.

## Form Filling Pattern

```bash
cdp snapshot                    # find form fields
# [textbox uid=a3] "Email"
# [textbox uid=a4] "Password"
# [button uid=a5] "Submit"

cdp fill-form --field a3=user@example.com --field a4=secret
cdp click a5
cdp wait "Dashboard"            # wait for navigation
cdp screenshot --file result.png
```

## Testing a Page

```bash
cdp navigate http://localhost:3000
cdp wait "Ready"
cdp snapshot                    # verify elements rendered
cdp eval "document.querySelectorAll('.error').length"  # check for errors
cdp screenshot --file test.png  # visual verification
```

## State

A background daemon auto-starts on first command, holds the Chrome WebSocket open, and proxies all subsequent commands. UIDs and page selection live in daemon memory. `cdp disconnect` stops the daemon. Idle auto-shutdown after 30 minutes.

## Updating cdp

When the user asks to update cdp or check for a new version, run this workflow:

```bash
# 1. Check current version
CURRENT=$(cdp version)

# 2. Get latest release tag from GitHub
LATEST=$(curl -s \
  -H "Authorization: token $(gh auth token)" \
  "https://api.github.com/repos/torfahsing/cdp/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)

echo "Current: $CURRENT  Latest: $LATEST"

# 3. If up to date, stop here
if [ "$CURRENT" = "$LATEST" ]; then
  echo "Already up to date."
  exit 0
fi

# 4. Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
BINARY="cdp-${OS}-${ARCH}"
INSTALL_PATH=$(which cdp)

# 5. Stop the daemon so the binary is not in use
cdp disconnect 2>/dev/null || true

# 6. Get asset API URL and download (browser_download_url redirects to HTML on GHE without Accept header)
TOKEN=$(gh auth token)
ASSET_URL=$(curl -s \
  -H "Authorization: token $TOKEN" \
  "https://api.github.com/repos/torfahsing/cdp/releases/latest" \
  | grep -A1 "\"name\": \"${BINARY}\"" | grep '"url"' | head -1 | cut -d'"' -f4)

curl -fsSL \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/octet-stream" \
  "$ASSET_URL" \
  -o /tmp/cdp-update

chmod +x /tmp/cdp-update
cp /tmp/cdp-update "$INSTALL_PATH"
rm /tmp/cdp-update

# 7. Verify
cdp version
```

Notes:
- Uses `gh auth token` for GitHub Enterprise auth — no separate token management needed
- Stops the daemon before replacing the binary to avoid file-in-use errors
- `$INSTALL_PATH` resolves to wherever `cdp` is installed (e.g. `~/.local/bin/cdp`)
- Uses the API asset URL with `Accept: application/octet-stream` — `browser_download_url` returns HTML on GHE without this header

## Common Issues

| Problem | Fix |
|---------|-----|
| "cannot connect to Chrome" | Enable remote debugging — see Setup |
| "unknown UID" | Run `cdp snapshot` first to generate UIDs |
| Stale page selection | Run `cdp pages` and `cdp select <id>` |
| Screenshot is blank | Page may not be loaded — add `cdp wait "text"` first |
| "daemon did not start" | Check Chrome is running with remote debugging |
| Connection lost after Chrome restart | Run `cdp disconnect` then retry |
