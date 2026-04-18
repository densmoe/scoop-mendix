#!/bin/bash
# Update Scoop manifest hashes from winget manifests

WINGET_DIR="../winget-mendix/manifests/Mendix/MendixStudioPro"
BUCKET_DIR="bucket"

if [ ! -d "$WINGET_DIR" ]; then
  echo "Error: winget-mendix directory not found"
  exit 1
fi

count=0
for manifest in "$BUCKET_DIR"/mendix-studio-pro-*.json; do
  version=$(basename "$manifest" | sed 's/mendix-studio-pro-\(.*\)\.json/\1/')
  winget_file="$WINGET_DIR/$version/Mendix.MendixStudioPro.installer.yaml"

  if [ ! -f "$winget_file" ]; then
    continue
  fi

  # Extract hashes from winget manifest
  x64_hash=$(grep -A 1 "User-x64-Setup.exe" "$winget_file" | grep "InstallerSha256:" | awk '{print $2}')
  arm64_hash=$(grep -A 1 "User-arm64-Setup.exe" "$winget_file" | grep "InstallerSha256:" | awk '{print $2}')

  if [ -n "$x64_hash" ] && [ -n "$arm64_hash" ]; then
    # Update the Scoop manifest
    sed -i.bak \
      -e "s/\"SHA256_PLACEHOLDER\"/\"$x64_hash\"/" \
      "$manifest"

    # Replace the second placeholder for arm64
    awk -v arm="$arm64_hash" '
      /SHA256_PLACEHOLDER/ {
        if (!done) {
          done=1;
          next;
        }
        gsub(/SHA256_PLACEHOLDER/, arm);
      }
      {print}
    ' "$manifest" > "$manifest.tmp" && mv "$manifest.tmp" "$manifest"

    rm -f "$manifest.bak"
    ((count++))
    echo "Updated $version"
  fi
done

echo "Updated $count manifests with real hashes"
