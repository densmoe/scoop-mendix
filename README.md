# scoop-mendix

Scoop bucket for Mendix Studio Pro - side-by-side version management on Windows.

## Installation

```powershell
# Add the bucket
scoop bucket add mendix https://github.com/densmoe/scoop-mendix

# Install latest version
scoop install mendix/mendix-studio-pro

# Install latest Mx10 or Mx11
scoop install mendix/mendix-studio-pro-10
scoop install mendix/mendix-studio-pro-11

# Install latest patch in a minor version
scoop install mendix/mendix-studio-pro-10.24

# Install exact version
scoop install mendix/mendix-studio-pro-10.18.13

# List available versions
scoop search mendix
```

See [USAGE.md](USAGE.md) for more examples.

## Features

- **Semantic version selection**: Install by major (`-10`), minor (`-10.24`), or exact version
- **Side-by-side versions**: Multiple Studio Pro versions can coexist
- **User-scope install**: No admin rights required (x64 + ARM64, Mendix 9.23.0+)
- **Machine-scope install**: System-wide installation available (x64 only, requires admin)
- **Daily updates**: Automatic manifest generation for new releases

## Developer Setup

### Prerequisites
- Go 1.21+

### Generate New Manifests

The daily workflow automatically generates manifests for new releases. To manually generate:

```bash
cd tools/manifest-generator
go run . -bucket-dir ../../bucket -min-major 9
```

### Generator Flags
- `-bucket-dir`: Output directory for Scoop manifests (default: `../../bucket`)
- `-skip-sha`: Skip SHA256 computation (uses placeholders)
- `-dry-run`: Preview without writing files
- `-version-types`: Filter by type (default: `LTS,MTS,Stable`)
- `-min-major`: Minimum major version (default: `10`)
- `-max-versions`: Limit batch size (default: unlimited)
- `-workers`: Parallel workers (default: `5`)

## How It Works

1. **Queries Mendix Marketplace API** for all Studio Pro releases
2. **Generates versioned manifests** for each version (user + machine scope)
3. **Generates alias manifests** for semantic version selection (user-scope only)
4. **Fetches SHA256 hashes** from CDN `.sha256` sidecar files (9.24.34+) or computes them
5. **Runs daily via GitHub Actions** to catch new releases (up to 10 per day)

### Manifest Types

**Versioned manifests** (330 files):
- User-scope: `mendix-studio-pro-10.18.13.json` (x64 + ARM64, no admin)
- Machine-scope: `mendix-studio-pro-10.18.13-machine.json` (x64 only, requires admin)
- Each points to specific installer on CDN
- Immutable after creation

**Alias manifests** (41 files, user-scope only):
- Major: `mendix-studio-pro-10.json` → latest 10.x
- Minor: `mendix-studio-pro-10.24.json` → latest 10.24.x
- Latest: `mendix-studio-pro.json` → newest overall
- Regenerated when new versions are published

### Installation Details

Mendix Studio Pro installs to `%LOCALAPPDATA%\Programs\Mendix\<version>\` (user-scope) or `%ProgramFiles%\Mendix\<version>\` (machine-scope) and registers itself in the Start Menu. Scoop tracks the installation via the Windows registry and can uninstall it, but does not create command-line shims. Users launch Mendix via the Start Menu or directly from the install location.

## Related Projects

- [winget-mendix](https://github.com/densmoe/winget-mendix) - Windows Package Manager manifests
- [homebrew-mendix](https://github.com/densmoe/homebrew-mendix) - macOS Homebrew tap
