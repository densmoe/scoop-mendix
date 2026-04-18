# scoop-mendix

Scoop bucket for Mendix Studio Pro - side-by-side version management on Windows.

## Installation

```powershell
# Add the bucket
scoop bucket add mendix https://github.com/densmoe/scoop-mendix

# Install a specific version
scoop install mendix/mendix-studio-pro-10.0.0

# List available versions
scoop search mendix
```

## Architecture Support

Each version includes up to three installer variants:
- **Machine x64** — traditional admin installer
- **User x64** — no admin required (default)
- **User ARM64** — no admin required, for ARM devices

Scoop prefers user-scope installers, so the manifest defaults to the user x64 variant. Use `--arch arm64` for ARM devices.

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

1. Queries Mendix Marketplace API for all Studio Pro releases
2. Filters by version type (LTS/MTS/Stable) and minimum version
3. Skips versions that already have complete manifests
4. Validates installer availability on CDN (HEAD request + Content-Length > 100MB to detect error pages)
5. Fetches SHA256 hashes from CDN `.sha256` sidecar files, or streams the full download for older versions
6. Generates one JSON manifest per version in Scoop format

### Daily Automation

The GitHub Actions workflow runs daily and processes up to 10 new versions per run, catching up incrementally as new Mendix releases are published.

## Scoop Manifest Format

Each version gets a single JSON file (e.g., `mendix-studio-pro-10.0.0.json`) with:
- Version number
- Download URLs for all three installer variants (x64 machine/user, ARM64 user)
- SHA256 hashes
- Architecture-specific installer switches
- Side-by-side installation support via `persist` directory

## Related Projects

- [winget-mendix](https://github.com/densmoe/winget-mendix) - Windows Package Manager manifests
- [homebrew-mendix](https://github.com/densmoe/homebrew-mendix) - macOS Homebrew tap
