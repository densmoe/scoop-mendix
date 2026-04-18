# Using the Mendix Scoop Bucket

## Add the bucket

```powershell
scoop bucket add mendix https://github.com/densmoe/scoop-mendix
```

## Install a version

### Latest in a branch

```powershell
# Latest overall
scoop install mendix/mendix-studio-pro

# Latest Mx11
scoop install mendix/mendix-studio-pro-11

# Latest Mx10
scoop install mendix/mendix-studio-pro-10

# Latest 10.24.x
scoop install mendix/mendix-studio-pro-10.24
```

### Specific version

```powershell
# Exact version (user-scope, no admin required)
scoop install mendix/mendix-studio-pro-11.9.1
scoop install mendix/mendix-studio-pro-10.18.13
scoop install mendix/mendix-studio-pro-9.24.42
```

## Machine-scope Installation

For system-wide installation (requires admin rights):

```powershell
# User-scope (default, no admin)
scoop install mendix/mendix-studio-pro-11.9.1

# Machine-scope (requires admin, x64 only)
scoop install mendix/mendix-studio-pro-11.9.1-machine
```

**Differences:**
- **User-scope**: Installs to `%LOCALAPPDATA%\Programs\Mendix\`, no admin, supports ARM64
- **Machine-scope**: Installs to `%ProgramFiles%\Mendix\`, requires admin, x64 only

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
