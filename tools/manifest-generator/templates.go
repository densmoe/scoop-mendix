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
  "innosetup": true,
  "post_install": [
    "$installPath = \"$env:LOCALAPPDATA\\Programs\\Mendix\"",
    "$version = Get-ChildItem $installPath | Where-Object { $_.Name -like '{{.Version}}*' } | Select-Object -First 1 -ExpandProperty Name",
    "$exePath = \"$installPath\\$version\\modeler\\Mendix.exe\"",
    "if (Test-Path $exePath) {",
    "  New-Item -ItemType SymbolicLink -Path \"$dir\\Mendix.exe\" -Target $exePath -Force | Out-Null",
    "} else {",
    "  Write-Host \"Warning: Mendix executable not found at $exePath\" -ForegroundColor Yellow",
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
    "script": "Get-ItemProperty 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*' | Where-Object { $_.DisplayName -like '*Mendix*{{.Version}}*' } | ForEach-Object { Start-Process -Wait ($_.UninstallString -replace '\"','') -ArgumentList '/VERYSILENT' }"
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
