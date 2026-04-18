# scoop-mendix

Scoop bucket for Mendix Studio Pro versions.

## Project Structure

```
tools/manifest-generator/  Go CLI that generates Scoop manifests
  main.go                  CLI orchestration, SHA256 fetching, manifest completeness checks
  marketplace.go           Mendix Marketplace API client
  templates.go             JSON manifest template (Scoop format)
bucket/                    Generated JSON manifests (Scoop format)
.github/workflows/         CI + daily automation
```

## Key Commands

### Generate manifests
```bash
cd tools/manifest-generator
go run . -bucket-dir ../../bucket -min-major 9
go run . -bucket-dir ../../bucket -skip-sha      # Fast mode (placeholder hashes)
go run . -bucket-dir ../../bucket -dry-run       # Preview only
```

The generator fetches SHA256 hashes automatically:
1. First tries `.sha256` sidecar files (available from 9.24.34+)
2. Falls back to downloading and computing hash (slower, for older versions)

### Import hashes from winget (one-time bootstrap only)
```bash
python3 tools/import-hashes.py
```

This was used for initial setup to avoid downloading 165 multi-GB installers. **Not needed for daily operations** — the Go generator handles new versions automatically.

### Run tests
```bash
cd tools/manifest-generator
go test ./...
```

### Test locally
```powershell
# Add local bucket
scoop bucket add mendix C:\path\to\scoop-mendix

# Install a version
scoop install mendix/mendix-studio-pro-10.0.0
```

## Architecture

The generator runs on any OS (Linux, macOS, Windows). No external tools needed — everything is pure Go.

- **marketplace.go**: Fetches releases from `marketplace.mendix.com/xas/` with CSRF token session init
- **main.go**: Orchestrates generation with resumability — checks if manifest exists with real hashes
- **templates.go**: Single JSON template per version (Scoop format)

## Mendix Version Patterns

**Manifest version** (in JSON filename and `version` field):
- All versions: 3-part semantic version (e.g., `9.24.42`, `10.18.13`, `11.5.0`)

**CDN installer filenames**:
- **Mx9 all**: 4-part `Mendix-9.24.42.98709-User-x64-Setup.exe`
- **Mx10 all**: 4-part `Mendix-10.18.13.89970-User-x64-Setup.exe`
- **Mx11.0–11.4**: 4-part `Mendix-11.4.0.83498-User-x64-Setup.exe`
- **Mx11.5+**: 3-part `Mendix-11.5.0-User-x64-Setup.exe`

The generator tries 4-part CDN URLs first and falls back to 3-part. The installer script uses wildcards (`Get-ChildItem "Mendix-*-User-x64-Setup.exe"`) to handle both patterns.

## Installer URLs

CDN base: `https://artifacts.rnd.mendix.com/modelers/`

- **Machine x64**: `Mendix-{VERSION}-Setup.exe` (requires admin, NOT used in Scoop)
- **User x64**: `Mendix-{VERSION}-User-x64-Setup.exe` (available from 9.23.0+)
- **User ARM64**: `Mendix-{VERSION}-User-arm64-Setup.exe` (available from 9.23.0+)

**Scoop bucket only includes versions with user-scope installers** (9.23.0+) since Scoop is designed for non-admin installs.

SHA256 sidecar files at `{url}.sha256` (available from 9.24.34+). For older versions (9.23.0-9.24.33), hashes are imported from winget-mendix manifests.

## Scoop Bucket Format

- Each version is a separate JSON file: `mendix-studio-pro-{version}.json`
- Scoop prefers user-scope installers (no admin required)
- The manifest includes all three installer variants with architecture-specific URLs
- Multiple versions can coexist side-by-side

## Git Workflow

- Default branch: `main`
- Daily workflow commits new manifests automatically (up to 10/day)

## Safety Rules

- NEVER commit AWS credentials or secrets
- See security guidelines in parent CLAUDE.md files
