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

SHA256 sidecar files at `{url}.sha256` (available from 9.24.34+). For older versions (9.23.0-9.24.33), the generator downloads and computes hashes.

## Scoop Bucket Format

- Each version is a separate JSON file: `mendix-studio-pro-{version}.json`
- Scoop prefers user-scope installers (no admin required)
- The manifest includes architecture-specific URLs (x64, ARM64)
- Multiple versions can coexist side-by-side

### Alias Manifests

Aliases allow users to install "latest in branch" semantics:
- `mendix-studio-pro.json` → latest overall
- `mendix-studio-pro-10.json` → latest Mx10
- `mendix-studio-pro-10.24.json` → latest 10.24.x

These are generated automatically by the Go generator after creating versioned manifests.

## Git Workflow

- Default branch: `main`
- Daily workflow commits new manifests automatically (up to 10/day)
- Workflow regenerates both versioned manifests AND alias manifests

## Key Learnings

### Mendix Installation Behavior

1. **Executable name**: `studiopro.exe` (not `Mendix.exe`)
2. **Install location**: `%LOCALAPPDATA%\Programs\Mendix\<version-with-build>\`
   - Example: `10.18.13.89970` or `11.9.1` (Mx11.5+ uses 3-part)
3. **User vs Machine scope**: Only user-scope installers are supported (9.23.0+)
4. **Installer behavior**: Ignores `/DIR` parameter, always installs to default location

### Scoop Integration Approach

1. **Don't fight the installer**: Let Mendix install where it wants
2. **No shimming needed**: Mendix registers itself in Start Menu
3. **Scoop tracks**: Installation/uninstallation via registry
4. **User access**: Via Start Menu or direct path

### Version Management

1. **Versioned manifests**: 165 manifests for specific versions (9.23.0-11.9.1)
2. **Alias manifests**: 41 aliases for semantic version selection
3. **Daily updates**: Automatically checks for new releases and updates both types

## Safety Rules

- NEVER commit AWS credentials or secrets
- See security guidelines in parent CLAUDE.md files
