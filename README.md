# cdp

A CLI for Chrome DevTools Protocol. Controls a running Chrome instance from the terminal — take snapshots, click elements, evaluate JavaScript, capture screenshots, and more.

Built as a lightweight alternative to chrome-devtools-mcp. Where MCP tool calls consume significant context window tokens, `cdp` produces compact output via standard Bash calls.

## Setup

### 1. Install

**Option A — Browser (no tools required):**

1. Go to [github.com/torfahsing/cdp/releases/latest](https://github.com/torfahsing/cdp/releases/latest)
2. Download the binary for your platform:
   - `cdp-darwin-arm64` — macOS Apple Silicon
   - `cdp-darwin-amd64` — macOS Intel
   - `cdp-linux-amd64` — Linux
   - `cdp-windows-amd64.exe` — Windows
3. Move to your PATH and make executable:
   - **macOS/Linux:**
     ```bash
     chmod +x cdp-darwin-arm64 && mv cdp-darwin-arm64 /usr/local/bin/cdp
     ```
   - **Windows:** install to Git Bash PATH so Claude Code can find it:
     ```bash
     # In Git Bash
     cp cdp-windows-amd64.exe /usr/local/bin/cdp.exe
     ```
     For CMD/PowerShell, also copy to System32:
     ```powershell
     Copy-Item cdp-windows-amd64.exe C:\Windows\System32\cdp.exe
     ```

**Option B — `gh` CLI:**

```bash
# macOS ARM (Apple Silicon)
gh release download -R torfahsing/cdp --pattern 'cdp-darwin-arm64'
chmod +x cdp-darwin-arm64 && mv cdp-darwin-arm64 /usr/local/bin/cdp

# macOS Intel
gh release download -R torfahsing/cdp --pattern 'cdp-darwin-amd64'
chmod +x cdp-darwin-amd64 && mv cdp-darwin-amd64 /usr/local/bin/cdp

# Linux AMD64
gh release download -R torfahsing/cdp --pattern 'cdp-linux-amd64'
chmod +x cdp-linux-amd64 && mv cdp-linux-amd64 /usr/local/bin/cdp

# Windows (Git Bash)
gh release download -R torfahsing/cdp --pattern 'cdp-windows-amd64.exe'
cp cdp-windows-amd64.exe /usr/local/bin/cdp.exe
```

**Build from source (requires Go):**

```bash
make install   # builds and copies to ~/.local/bin/cdp
```

### 2. Enable Chrome remote debugging

**Option A — Launch Chrome with the flag:**

```bash
# macOS
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222

# Linux
google-chrome --remote-debugging-port=9222
```

```powershell
# Windows (PowerShell)
& "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222
```

**Option B — Enable from within Chrome (no restart needed):**

1. Navigate to `chrome://inspect/#remote-debugging`
2. Enable remote debugging and accept the permission dialog

### 3. Verify connection

```bash
cdp pages
```

You should see a list of open tabs.

### 4. Install the Claude Code skill

In Claude Code, run these slash commands:

```
# With SSH:
/plugin marketplace add git@github.com:torfahsing/cdp.git

# With gh CLI (HTTPS):
gh auth setup-git
/plugin marketplace add https://github.com/torfahsing/cdp.git
```

Then install and reload:

```
/plugin install cdp
/reload-plugins
```

## Quick Start

```bash
# See what's open
cdp pages

# Navigate somewhere
cdp navigate https://example.com

# Take a snapshot (accessibility tree with UIDs)
cdp snapshot

# Click an element using its UID from the snapshot
cdp click a3

# Take a screenshot
cdp screenshot --file page.png
```

## Core Workflow

The typical workflow is **snapshot → interact → verify**:

```bash
# 1. See the page structure
cdp snapshot
# Output:
# [link uid=a1] "Home" href="/"
# [button uid=a2] "Sign In"
# [textbox uid=a3] "Email" placeholder="you@example.com"
# [textbox uid=a4] "Password"

# 2. Fill the form
cdp fill a3 "user@example.com"
cdp fill a4 "hunter2"

# 3. Submit
cdp click a2

# 4. Verify the result
cdp snapshot
cdp screenshot --file after-login.png
```

## Commands Reference

### Navigation

```bash
cdp navigate https://example.com       # go to URL
cdp navigate --back                     # browser back
cdp navigate --forward                  # browser forward
cdp navigate --reload                   # reload page
cdp navigate --reload --ignore-cache    # hard reload
```

### Page Management

```bash
cdp pages                    # list all open tabs
cdp select <page-id>         # select a tab for commands
cdp select <page-id> --focus # select and bring to front
cdp open https://example.com # open new tab
cdp open https://example.com --background  # open without focusing
cdp close <page-id>          # close a tab
```

The selected page persists across invocations (stored in `~/.cdp/state.json`). All commands operate on the selected page. If none is selected, the first available page is used.

### Snapshots (Accessibility Tree)

```bash
cdp snapshot                  # compact a11y tree with UIDs
cdp snapshot --verbose        # include generic/none-role elements
cdp snapshot --file tree.txt  # save to file
cdp snapshot --json           # structured JSON output
```

The snapshot assigns short UIDs (`a1`, `a2`, ...) to each element. These UIDs are used by interaction commands (`click`, `fill`, `hover`, etc.). UIDs are valid until the next `snapshot` call.

### Screenshots

```bash
cdp screenshot --file page.png                # viewport screenshot
cdp screenshot --file full.png --full-page    # entire scrollable page
cdp screenshot --file photo.jpg --format jpeg --quality 90
cdp screenshot --file fast.webp --format webp --quality 60
cdp screenshot                                # raw PNG to stdout (pipe to file)
cdp screenshot --json                         # base64-encoded in JSON
```

### JavaScript Evaluation

```bash
cdp eval "document.title"
# "Example Domain"

cdp eval "document.querySelectorAll('a').length"
# 3

cdp eval "window.location.href"
# "https://example.com/"

# Multi-line / complex expressions
cdp eval "(() => { const els = document.querySelectorAll('img'); return els.length; })()"

# Get structured data
cdp eval 'JSON.stringify(performance.timing)' --json
```

### Clicking & Hovering

```bash
cdp click a2                  # click element
cdp click a2 --dbl            # double-click
cdp click a2 --snapshot       # click, then print updated snapshot
cdp hover a5                  # hover (trigger tooltips, menus)
cdp drag a1 a2                # drag element a1 onto a2
```

### Filling Forms

```bash
# Single field
cdp fill a3 "hello world"

# Multiple fields at once
cdp fill-form --field a3=user@example.com --field a4=password123

# Select dropdown
cdp fill a5 "Option B"
```

### Typing & Keys

```bash
# Type into focused element (simulates keystrokes)
cdp type "hello world"
cdp type "search query" --submit Enter    # type then press Enter

# Press keys or combinations
cdp key Enter
cdp key Tab
cdp key "Control+A"
cdp key "Control+C"
cdp key "Shift+1"       # useful for Figma zoom-to-selection
cdp key Escape
```

### Waiting

```bash
cdp wait "Loading complete"              # wait for text to appear
cdp wait "Success" "Done" "Complete"     # wait for any of these
cdp wait "Ready" --timeout 60000         # custom timeout (ms)
```

### Dialogs

```bash
cdp dialog accept              # accept alert/confirm/prompt
cdp dialog dismiss             # dismiss/cancel
cdp dialog accept --text "yes" # respond to prompt dialog
```

### File Upload

```bash
cdp upload a7 /path/to/file.pdf    # upload through file input element
```

### Viewport & Emulation

```bash
cdp resize 1920 1080                           # set viewport size

cdp emulate --viewport 375x812x3               # iPhone-like viewport (WxHxDPR)
cdp emulate --color-scheme dark                 # dark mode
cdp emulate --color-scheme light                # light mode
cdp emulate --user-agent "Custom Agent/1.0"     # override UA
cdp emulate --cpu-throttle 4                    # 4x CPU slowdown
cdp emulate --geo "37.7749x-122.4194"           # San Francisco
cdp emulate --network "Slow 3G"                 # throttle network
```

### Console & Network

```bash
cdp console            # list console messages
cdp console --json     # structured output

cdp network            # list network requests (via Performance API)
cdp network --json     # structured output with timing data
```

### Performance Tracing

```bash
cdp perf start                         # start recording trace
cdp perf start --reload                # start and reload page
cdp perf stop --file trace.json        # stop and save trace
cdp perf stop --file trace.json.gz     # compressed trace
cdp perf insight <set-id> <name>       # analyze specific insight
```

### Heap Snapshots

```bash
cdp memory heap.heapsnapshot    # capture heap snapshot for analysis
```

## Using with Claude Code

### Allow cdp commands

Replace MCP tool calls with compact Bash calls. Instead of the MCP server consuming context tokens with tool schemas and structured responses, `cdp` outputs minimal text.

**Claude Code hook example** — add to `.claude/settings.json` to have Claude use `cdp` automatically:

```json
{
  "permissions": {
    "allow": [
      "Bash(cdp *)"
    ]
  }
}
```

**Example Claude Code session:**

```
> Navigate to the login page and fill in credentials

$ cdp navigate https://app.example.com/login
Navigated to https://app.example.com/login "Login"

$ cdp snapshot
[heading uid=a1] "Sign In"
[textbox uid=a2] "Email"
[textbox uid=a3] "Password"
[button uid=a4] "Log In"

$ cdp fill-form --field a2=user@example.com --field a3=secret
Filled a2 with "user@example.com"
Filled a3 with "secret"

$ cdp click a4
Clicked a4
```

### Figma Workflows

Works the same as the MCP server for Figma design review:

```bash
# Open Figma file
cdp navigate "https://www.figma.com/design/..."

# Take snapshot to find layers
cdp snapshot

# Click a layer to select it
cdp click a15

# Zoom to fit selection
cdp key "Shift+1"

# Screenshot the design
cdp screenshot --file design.png
```

## Global Flags

Every command accepts these flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `9222` | Chrome remote debugging port |
| `--json` | `false` | Output as JSON instead of plain text |
| `--timeout` | `30000` | Timeout in milliseconds |

The port can also be set via the `CDP_PORT` environment variable.

## State

`cdp` is stateless between invocations — each command connects, executes, and disconnects. Two pieces of state persist in `~/.cdp/state.json`:

1. **Selected page** — which tab commands target (set via `cdp select`)
2. **UID mappings** — element UIDs from the last `cdp snapshot` (used by `click`, `fill`, etc.)

If Chrome restarts or tabs change, stale state is detected and reset automatically.
