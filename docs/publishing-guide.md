# Publishing CDP to github.com/torfahsing

Guide for setting up the repo, publishing releases, and making the tool available to others.

## Prerequisites

- Authenticated with GitHub: `gh auth status`
- Branch renamed to `main` (done)
- LICENSE and THIRD-PARTY-NOTICES in place (done, MIT license)

## 1. Prepare .gitignore

Add build artifacts before first push:

```bash
# .gitignore should contain:
/cdp
~/
dist/
```

`dist/` contains ~44MB of binaries — these go in GitHub Releases, not in git history.

## 2. Review what gets committed

**Include:**
- All Go source (`cmd/`, `internal/`)
- `go.mod`, `go.sum`
- `Makefile`
- `LICENSE`, `THIRD-PARTY-NOTICES`
- `README.md`
- `skills/cdp/SKILL.md`
- `.claude-plugin/plugin.json`

**Exclude (via .gitignore):**
- `dist/` — build artifacts
- `/cdp` — local binary
- `~/` — volta artifact from old zshrc bug

**Review before committing:**
- `docs/` — decide what's personal notes vs project docs
- `docs/superpowers/` — check contents, may be internal

## 3. Create the repo

```bash
gh repo create torfahsing/cdp \
  --source . \
  --private \
  --description "Chrome DevTools Protocol CLI for Claude Code" \
  --push
```

This creates the repo, adds remote, and pushes `main` in one command.

Change `--private` to `--public` if it should be visible outside the org.

## 4. Create a GitHub Release

Build binaries, tag, and release:

```bash
# Build all platforms
make build-all

# Tag the version
git tag v0.1.0
git push origin v0.1.0

# Create release with binaries
gh release create v0.1.0 dist/* \
  --title "v0.1.0" \
  --notes "Initial release — Chrome DevTools Protocol CLI for Claude Code"
```

Binaries will be downloadable at:
`github.com/torfahsing/cdp/releases/tag/v0.1.0`

## 5. Future releases

For subsequent versions:

```bash
make build-all
git tag v0.2.0
git push origin v0.2.0
gh release create v0.2.0 dist/* \
  --title "v0.2.0" \
  --notes "Description of changes"
```

## 6. Optional: automate releases with GitHub Actions

Can add later — a workflow that triggers on tag push, builds all platforms, and creates a release automatically. Not needed to start.

## 7. Installation instructions for users

Add to README once published:

```bash
# Download latest release (example for macOS ARM)
gh release download -R torfahsing/cdp --pattern 'cdp-darwin-arm64'
chmod +x cdp-darwin-arm64
mv cdp-darwin-arm64 /usr/local/bin/cdp
```

Or with curl:
```bash
curl -sL https://github.com/torfahsing/cdp/releases/latest/download/cdp-darwin-arm64 -o /usr/local/bin/cdp
chmod +x /usr/local/bin/cdp
```

## Dependencies and licensing

- Project: MIT License
- All dependencies are permissive (MIT, Apache-2.0, BSD-3-Clause)
- Apache-2.0 deps (cobra, mousetrap) credited in THIRD-PARTY-NOTICES
- No copyleft obligations
