package main

import "text/template"

var scoopManifestTemplate = template.Must(template.New("scoop").Parse(`{
  "version": "{{.Version}}",
  "description": "Low-code application development platform",
  "homepage": "https://www.mendix.com/studio-pro/",
  "license": "Proprietary",
  "architecture": {
    "64bit": {
      "url": [
        "{{.UserX64URL}}"
      ],
      "hash": [
        "{{.UserX64SHA256}}"
      ],
      "installer": {
        "script": [
          "$installerPath = \"$dir\\Mendix-{{.Version}}-User-x64-Setup.exe\"",
          "Start-Process -Wait -FilePath $installerPath -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', '/DIR=\"$dir\\app\"'"
        ]
      }
    },
    "arm64": {
      "url": [
        "{{.UserARM64URL}}"
      ],
      "hash": [
        "{{.UserARM64SHA256}}"
      ],
      "installer": {
        "script": [
          "$installerPath = \"$dir\\Mendix-{{.Version}}-User-arm64-Setup.exe\"",
          "Start-Process -Wait -FilePath $installerPath -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', '/DIR=\"$dir\\app\"'"
        ]
      }
    }
  },
  "bin": "app\\bin\\mendix.exe",
  "shortcuts": [
    [
      "app\\bin\\mendix.exe",
      "Mendix Studio Pro {{.Version}}"
    ]
  ],
  "checkver": {
    "url": "https://marketplace.mendix.com/link/studiopro/",
    "jsonpath": "$.version"
  },
  "notes": "Multiple versions can be installed side-by-side by installing with version-specific names."
}
`))

type ScoopManifestData struct {
	Version        string
	UserX64URL     string
	UserX64SHA256  string
	UserARM64URL   string
	UserARM64SHA256 string
}
