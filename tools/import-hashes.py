#!/usr/bin/env python3
"""Import SHA256 hashes from winget manifests into Scoop manifests

This script extracts SHA256 hashes from winget-mendix manifests and updates
the corresponding Scoop manifests. This is necessary because:

1. SHA256 sidecar files (.sha256) only exist on CDN from 9.24.34+
2. For older versions (9.23.0-9.24.33), computing hashes requires downloading
   multi-GB installers
3. Winget already has all hashes computed, so we can reuse them

The script only processes versions that:
- Have user-scope installers (9.23.0+)
- Have both x64 and arm64 hashes in winget
"""

import json
import re
from pathlib import Path

WINGET_DIR = Path("../winget-mendix/manifests/Mendix/MendixStudioPro")
BUCKET_DIR = Path("bucket")

def extract_hashes(winget_file):
    """Extract x64 and arm64 hashes from winget installer.yaml"""
    content = winget_file.read_text()

    # Find User-x64-Setup.exe hash
    x64_match = re.search(
        r'User-x64-Setup\.exe\s+InstallerSha256:\s+([A-F0-9]+)',
        content,
        re.IGNORECASE
    )

    # Find User-arm64-Setup.exe hash
    arm64_match = re.search(
        r'User-arm64-Setup\.exe\s+InstallerSha256:\s+([A-F0-9]+)',
        content,
        re.IGNORECASE
    )

    return (
        x64_match.group(1) if x64_match else None,
        arm64_match.group(1) if arm64_match else None
    )

def update_manifest(manifest_file, x64_hash, arm64_hash):
    """Update Scoop manifest with real hashes"""
    data = json.loads(manifest_file.read_text())

    if "architecture" in data:
        if "64bit" in data["architecture"] and x64_hash:
            data["architecture"]["64bit"]["hash"] = [x64_hash]

        if "arm64" in data["architecture"] and arm64_hash:
            data["architecture"]["arm64"]["hash"] = [arm64_hash]

    manifest_file.write_text(json.dumps(data, indent=2) + "\n")

def main():
    if not WINGET_DIR.exists():
        print(f"Error: {WINGET_DIR} not found")
        return

    count = 0
    for manifest_file in sorted(BUCKET_DIR.glob("mendix-studio-pro-*.json")):
        version = manifest_file.stem.replace("mendix-studio-pro-", "")
        winget_file = WINGET_DIR / version / "Mendix.MendixStudioPro.installer.yaml"

        if not winget_file.exists():
            continue

        x64_hash, arm64_hash = extract_hashes(winget_file)

        if x64_hash and arm64_hash:
            update_manifest(manifest_file, x64_hash, arm64_hash)
            count += 1
            print(f"Updated {version}")

    print(f"\nUpdated {count} manifests with real hashes")

if __name__ == "__main__":
    main()
