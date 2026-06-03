# Future Features & Improvements

This document tracks planned features and improvements for future iterations of `cdp`.

---

## 1. Native CLI Version Check & Release Notification

Currently, version checking and updating are handled externally by AI agents via instructions in `SKILL.md` (using `curl` and `gh` shell commands). To make `cdp` more user-friendly for human operators and CLI users, we want to build update-checking logic directly into the Go CLI itself.

### Planned Behavior
* When running `cdp version`, the CLI should print the current local version and also check if a newer release exists on GitHub.
* If a new release is found (e.g., local is `v1.0.0` but latest release is `v1.2.0`), print a friendly message with the download instructions.
* (Optional) To keep general command execution fast, run this check asynchronously or cache the last check time (e.g., check at most once per day) to avoid blocking command startup.

### Implementation Guide

1. **GitHub API Query**
   Using Go's standard library `net/http` package, send a `GET` request to the public repository's release endpoint:
   ```go
   resp, err := http.Get("https://api.github.com/repos/torfahsing/cdp/releases/latest")
   ```

2. **JSON Parsing**
   Define a structure matching the minimal GitHub releases response:
   ```go
   type GitHubRelease struct {
       TagName string `json:"tag_name"`
   }
   ```
   Parse the response body using `encoding/json` (or the `go-json-experiment/json` package already present in `go.mod`):
   ```go
   var release GitHubRelease
   if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
       latestVersion := release.TagName // e.g., "v1.2.0"
   }
   ```

3. **Version Comparison**
   Compare the retrieved `latestVersion` against the locally defined `Version` string.
   * If they don't match, print a message like:
     ```
     A newer version of cdp is available: v1.2.0 (current: v1.0.0)
     To update, run: gh release download -R torfahsing/cdp --pattern 'cdp-*'
     ```
