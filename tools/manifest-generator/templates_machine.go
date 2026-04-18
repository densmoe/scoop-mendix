package main

import "text/template"

var scoopMachineManifestTemplate = template.Must(template.New("scoop-machine").Parse(`{
  "version": "{{.Version}}",
  "description": "Low-code application development platform (machine-scope, requires admin)",
  "homepage": "https://www.mendix.com/studio-pro/",
  "license": "Proprietary",
  "url": "{{.MachineURL}}",
  "hash": "{{.MachineSHA256}}",
  "pre_install": [
    "$installerPath = Get-ChildItem \"$dir\\Mendix-*-Setup.exe\" | Select-Object -First 1 -ExpandProperty FullName",
    "if ($installerPath) {",
    "  Write-Host \"Running Mendix installer (requires admin)...\" -ForegroundColor Cyan",
    "  Start-Process -Wait -FilePath $installerPath -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART' -Verb RunAs",
    "  Write-Host \"Installation complete\" -ForegroundColor Green",
    "} else {",
    "  Write-Host \"Installer not found\" -ForegroundColor Red",
    "  exit 1",
    "}"
  ],
  "post_install": [
    "$installPath = \"$env:ProgramFiles\\Mendix\"",
    "$versionDir = Get-ChildItem $installPath -ErrorAction SilentlyContinue | Where-Object { $_.Name -like '{{.Version}}*' } | Select-Object -First 1",
    "if ($versionDir) {",
    "  Write-Host \"Mendix Studio Pro {{.Version}} installed successfully (machine-scope)\" -ForegroundColor Green",
    "  Write-Host \"Location: $installPath\\$($versionDir.Name)\" -ForegroundColor Cyan",
    "  Write-Host \"Executable: $installPath\\$($versionDir.Name)\\modeler\\studiopro.exe\" -ForegroundColor Cyan",
    "} else {",
    "  Write-Host \"Warning: Could not verify Mendix installation\" -ForegroundColor Yellow",
    "}"
  ],
  "uninstaller": {
    "script": [
      "$uninstallKey = Get-ItemProperty 'HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\*' -ErrorAction SilentlyContinue | Where-Object { $_.DisplayName -like '*Mendix*{{.Version}}*' } | Select-Object -First 1",
      "if ($uninstallKey -and $uninstallKey.UninstallString) {",
      "  $uninstallCmd = $uninstallKey.UninstallString -replace '\"',''",
      "  if (Test-Path $uninstallCmd) {",
      "    Write-Host \"Uninstalling Mendix Studio Pro {{.Version}}...\" -ForegroundColor Cyan",
      "    Start-Process -Wait -FilePath $uninstallCmd -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART' -Verb RunAs",
      "  }",
      "}"
    ]
  },
  "checkver": {
    "url": "https://marketplace.mendix.com/link/studiopro/",
    "jsonpath": "$.version"
  },
  "notes": [
    "Machine-scope installation (requires admin rights).",
    "Installs to %ProgramFiles%\\Mendix\\<version>.",
    "Multiple versions can coexist.",
    "",
    "For user-scope installation (no admin), use: mendix-studio-pro-{{.Version}}"
  ]
}
`))

type ScoopMachineManifestData struct {
	Version        string
	MachineURL     string
	MachineSHA256  string
}
