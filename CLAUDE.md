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

- **Mx9**: 4-part `9.24.35.71123`, manifest version = `9.24.35`
- **Mx10**: 4-part `10.18.0.54340`, manifest version = `10.18.0`
- **Mx11.0–11.4**: 4-part URLs on CDN, 3-part manifest version
- **Mx11.5+**: 3-part everywhere (`11.5.0`)

The generator tries 4-part CDN URLs first and falls back to 3-part.

## Installer URLs

CDN base: `https://artifacts.rnd.mendix.com/modelers/`

- **Machine x64**: `Mendix-{VERSION}-Setup.exe`
- **User x64**: `Mendix-{VERSION}-User-x64-Setup.exe`
- **User ARM64**: `Mendix-{VERSION}-User-arm64-Setup.exe`

SHA256 sidecar files at `{url}.sha256` (available from 9.24.34+). Older versions need full download for hash computation.

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
