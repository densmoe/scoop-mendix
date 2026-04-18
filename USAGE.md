# Using the Mendix Scoop Bucket

## Add the bucket

```powershell
scoop bucket add mendix https://github.com/densmoe/scoop-mendix
```

## Install a specific version

```powershell
# Install latest version
scoop install mendix/mendix-studio-pro-11.9.1

# Install a specific older version
scoop install mendix/mendix-studio-pro-10.18.13

# Install Mx9
scoop install mendix/mendix-studio-pro-9.24.42
```

## List available versions

```powershell
scoop search mendix-studio-pro
```

Or check the bucket directory on GitHub:
https://github.com/densmoe/scoop-mendix/tree/main/bucket

## Side-by-side installations

Multiple versions can be installed at the same time:

```powershell
scoop install mendix/mendix-studio-pro-11.9.1
scoop install mendix/mendix-studio-pro-10.18.13
scoop install mendix/mendix-studio-pro-9.24.42
```

Each version installs independently and won't conflict.

## Update/Uninstall

```powershell
# Update a version (when new patch releases come out)
scoop update mendix-studio-pro-11.9.1

# Uninstall a version
scoop uninstall mendix-studio-pro-11.9.1

# List installed versions
scoop list mendix
```

## Architecture Support

The manifests support:
- **x64** - Standard 64-bit Windows
- **ARM64** - Windows on ARM devices

Scoop will automatically pick the right installer for your system.
