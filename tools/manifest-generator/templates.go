package main

import "text/template"

var scoopManifestTemplate = template.Must(template.New("scoop").Parse(`{
  "version": "{{.Version}}",
  "description": "Low-code application development platform",
  "homepage": "https://www.mendix.com/studio-pro/",
  "license": "Proprietary",
  "architecture": {
    "64bit": {
      "url": "{{.UserX64URL}}",
      "hash": "{{.UserX64SHA256}}"
    },
    "arm64": {
      "url": "{{.UserARM64URL}}",
      "hash": "{{.UserARM64SHA256}}"
    }
  },
  "pre_install": [
    "$installerPath = Get-ChildItem \"$dir\\Mendix-*-User-*.exe\" | Select-Object -First 1 -ExpandProperty FullName",
    "if ($installerPath) {",
    "  Write-Host \"Running Mendix installer...\" -ForegroundColor Cyan",
    "  Start-Process -Wait -FilePath $installerPath -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART'",
    "  Write-Host \"Installation complete\" -ForegroundColor Green",
    "} else {",
    "  Write-Host \"Installer not found\" -ForegroundColor Red",
    "  exit 1",
    "}"
  ],
  "post_install": [
    "$installPath = \"$env:LOCALAPPDATA\\Programs\\Mendix\"",
    "$version = Get-ChildItem $installPath -ErrorAction SilentlyContinue | Where-Object { $_.Name -like '{{.Version}}*' } | Select-Object -First 1 -ExpandProperty Name",
    "if ($version) {",
    "  $exePath = \"$installPath\\$version\\modeler\\Mendix.exe\"",
    "  if (Test-Path $exePath) {",
    "    New-Item -ItemType SymbolicLink -Path \"$dir\\Mendix.exe\" -Target $exePath -Force -ErrorAction SilentlyContinue | Out-Null",
    "    Write-Host \"Created shim to $exePath\" -ForegroundColor Green",
    "  } else {",
    "    Write-Host \"Warning: Mendix executable not found at $exePath\" -ForegroundColor Yellow",
    "  }",
    "} else {",
    "  Write-Host \"Warning: Could not find Mendix installation for version {{.Version}}\" -ForegroundColor Yellow",
    "}"
  ],
  "bin": "Mendix.exe",
  "shortcuts": [
    [
      "Mendix.exe",
      "Mendix Studio Pro {{.Version}}"
    ]
  ],
  "uninstaller": {
    "script": [
      "$uninstallKey = Get-ItemProperty 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*' -ErrorAction SilentlyContinue | Where-Object { $_.DisplayName -like '*Mendix*{{.Version}}*' } | Select-Object -First 1",
      "if ($uninstallKey -and $uninstallKey.UninstallString) {",
      "  $uninstallCmd = $uninstallKey.UninstallString -replace '\"',''",
      "  if (Test-Path $uninstallCmd) {",
      "    Write-Host \"Uninstalling Mendix Studio Pro {{.Version}}...\" -ForegroundColor Cyan",
      "    Start-Process -Wait -FilePath $uninstallCmd -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART'",
      "  }",
      "}"
    ]
  },
  "checkver": {
    "url": "https://marketplace.mendix.com/link/studiopro/",
    "jsonpath": "$.version"
  },
  "notes": "Mendix Studio Pro installs to %LOCALAPPDATA%\\Programs\\Mendix\\<version>. Multiple versions can coexist."
}
`))

type ScoopManifestData struct {
	Version         string
	UserX64URL      string
	UserX64SHA256   string
	UserARM64URL    string
	UserARM64SHA256 string
}
