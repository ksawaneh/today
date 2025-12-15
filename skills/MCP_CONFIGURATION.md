# MCP Configuration for Go CLI Development

This document describes the Model Context Protocol (MCP) servers that would enable Claude to build, test, and distribute Go CLI applications end-to-end.

---

## Overview

MCPs provide Claude with actual capabilities beyond reading/writing files. For CLI development, these servers enable:

1. **Building & Testing** — Compile Go code, run tests
2. **Version Control** — Create repos, push commits, manage releases
3. **Distribution** — Publish to package managers
4. **Demo Creation** — Record terminal sessions

---

## Recommended MCP Servers

### 1. Go Toolchain MCP (Critical)

Enables Claude to actually compile and test Go code.

```json
{
  "mcpServers": {
    "go-toolchain": {
      "command": "go-mcp-server",
      "args": [
        "--allow-network",
        "--allow-build",
        "--allow-test"
      ],
      "env": {
        "GOPROXY": "https://proxy.golang.org,direct",
        "GOPATH": "/home/user/go",
        "GOBIN": "/home/user/go/bin"
      }
    }
  }
}
```

**Capabilities:**
- `go build` — Compile code
- `go test` — Run tests with output
- `go mod tidy` — Resolve dependencies
- `go vet` — Static analysis
- `go run` — Execute without building
- `go install` — Install binaries

**Example Usage:**
```
Claude: Let me compile and test your code.

[Calls go-toolchain.build with path="./cmd/today"]
Result: Build successful. Binary at ./today (4.2MB)

[Calls go-toolchain.test with path="./..." flags="-race -cover"]
Result: PASS coverage: 78.3% of statements
```

---

### 2. GitHub MCP (Critical)

Enables Claude to manage repositories and releases.

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

**Capabilities:**
- Create repositories
- Push commits
- Create branches and PRs
- Manage releases with assets
- Set up GitHub Actions workflows
- Manage issues

**Example Usage:**
```
Claude: I'll create the repository and push the initial code.

[Calls github.create_repository with name="today" description="Terminal productivity dashboard"]
Result: Created github.com/user/today

[Calls github.push_files with files=[...]]
Result: Pushed 12 files to main branch

[Calls github.create_release with tag="v1.0.0" assets=["today_linux_amd64.tar.gz", ...]]
Result: Release v1.0.0 created with 6 assets
```

---

### 3. SQLite MCP

Enables Claude to prototype and test database schemas.

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite"],
      "env": {
        "SQLITE_DB_PATH": "/tmp/claude-sqlite"
      }
    }
  }
}
```

**Capabilities:**
- Create databases
- Execute queries
- Inspect schema
- Run migrations
- Test data models

**Example Usage:**
```
Claude: Let me prototype the schema before writing the Go code.

[Calls sqlite.execute with query="CREATE TABLE tasks (...)"]
Result: Table created

[Calls sqlite.execute with query="INSERT INTO tasks VALUES (...)"]
Result: 1 row inserted

[Calls sqlite.query with query="SELECT * FROM tasks WHERE done = 0"]
Result: [{id: "1", text: "Test task", done: 0}]
```

---

### 4. GoReleaser MCP

Enables Claude to build and release cross-platform binaries.

```json
{
  "mcpServers": {
    "goreleaser": {
      "command": "goreleaser-mcp-server",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}",
        "HOMEBREW_TAP_TOKEN": "${HOMEBREW_TAP_TOKEN}"
      }
    }
  }
}
```

**Capabilities:**
- Validate `.goreleaser.yaml`
- Build snapshot releases
- Create full releases
- Push to Homebrew taps
- Generate changelogs

**Example Usage:**
```
Claude: I'll create a release for all platforms.

[Calls goreleaser.check]
Result: Configuration valid

[Calls goreleaser.release with snapshot=false]
Result: 
  - Built: today_linux_amd64, today_darwin_arm64, today_windows_amd64.exe
  - Uploaded: 6 archives to GitHub Releases
  - Updated: homebrew-tap/Formula/today.rb
```

---

### 5. VHS MCP

Enables Claude to create demo GIFs programmatically.

```json
{
  "mcpServers": {
    "vhs": {
      "command": "vhs-mcp-server",
      "args": [],
      "env": {
        "VHS_FONT": "JetBrains Mono"
      }
    }
  }
}
```

**Capabilities:**
- Execute tape files
- Generate GIFs/MP4s
- Optimize output
- Capture specific terminal sizes

**Example Usage:**
```
Claude: I'll record a demo showing the core workflow.

[Calls vhs.record with tape="demo.tape"]
Result: Generated demo.gif (2.4MB, 12 seconds)

[Calls vhs.optimize with input="demo.gif" colors=64]
Result: Optimized to demo.gif (890KB)
```

---

### 6. Homebrew MCP

Enables Claude to manage Homebrew taps and formulae.

```json
{
  "mcpServers": {
    "homebrew": {
      "command": "homebrew-mcp-server",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

**Capabilities:**
- Create tap repositories
- Generate formulae from releases
- Test installations locally
- Push formula updates

**Example Usage:**
```
Claude: I'll set up Homebrew distribution.

[Calls homebrew.create_tap with name="homebrew-tap"]
Result: Created github.com/user/homebrew-tap

[Calls homebrew.create_formula with binary_url="..." sha256="..."]
Result: Created Formula/today.rb

[Calls homebrew.test_install with formula="today"]
Result: Installation successful, binary works
```

---

## Full Configuration

Complete `~/.config/claude/mcp.json`:

```json
{
  "mcpServers": {
    "go-toolchain": {
      "command": "go-mcp-server",
      "args": ["--allow-network", "--allow-build", "--allow-test"],
      "env": {
        "GOPROXY": "https://proxy.golang.org,direct"
      }
    },
    
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite"]
    },
    
    "goreleaser": {
      "command": "goreleaser-mcp-server",
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}",
        "HOMEBREW_TAP_TOKEN": "${HOMEBREW_TAP_TOKEN}"
      }
    },
    
    "vhs": {
      "command": "vhs-mcp-server"
    },
    
    "homebrew": {
      "command": "homebrew-mcp-server",
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"]
    }
  }
}
```

---

## Environment Variables Required

```bash
# GitHub Personal Access Token (repo, workflow scopes)
export GITHUB_TOKEN="ghp_..."

# Separate token for Homebrew tap (if using separate account)
export HOMEBREW_TAP_TOKEN="ghp_..."

# AUR SSH key (for Arch Linux packages)
export AUR_SSH_KEY="$(cat ~/.ssh/aur_ed25519)"
```

---

## Workflow With All MCPs

What becomes possible:

```
User: Build me a terminal habit tracker

Claude:
1. Creates project structure                    [filesystem]
2. Writes Go code                               [filesystem]
3. Compiles and runs tests                      [go-toolchain]
4. Fixes any compilation errors                 [go-toolchain]
5. Prototypes database schema                   [sqlite]
6. Creates GitHub repository                    [github]
7. Pushes code                                  [github]
8. Sets up GitHub Actions                       [github]
9. Records demo GIF                             [vhs]
10. Validates GoReleaser config                 [goreleaser]
11. Creates v1.0.0 release                      [goreleaser]
12. Sets up Homebrew tap                        [homebrew]
13. User runs: brew install user/tap/tracker
```

**Total time: One conversation. Published, installable software.**

---

## Priority Order

If limited MCP slots available:

| Priority | MCP | Impact |
|----------|-----|--------|
| 1 | `go-toolchain` | Actually verify code works |
| 2 | `github` | Publish and distribute |
| 3 | `sqlite` | Prototype data layer |
| 4 | `goreleaser` | Cross-platform releases |
| 5 | `vhs` | Create demos |
| 6 | `homebrew` | Easy installation |

---

## MCP Development Status

| MCP | Status | Notes |
|-----|--------|-------|
| `github` | ✅ Available | Official Anthropic MCP |
| `sqlite` | ✅ Available | Official Anthropic MCP |
| `filesystem` | ✅ Available | Official Anthropic MCP |
| `go-toolchain` | ❌ Needed | Would need to be built |
| `goreleaser` | ❌ Needed | Would need to be built |
| `vhs` | ❌ Needed | Would need to be built |
| `homebrew` | ❌ Needed | Would need to be built |

The `go-toolchain` MCP would have the highest impact and should be the first to be developed.
