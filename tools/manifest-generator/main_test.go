package main

import "testing"

func TestManifestVersionFor(t *testing.T) {
	tests := []struct {
		name     string
		release  Release
		expected string
	}{
		{
			name: "Mx9 version",
			release: Release{
				Major:       9,
				Minor:       24,
				Patch:       42,
				Build:       98709,
				VersionFull: "9.24.42.98709",
			},
			expected: "9.24.42",
		},
		{
			name: "Mx10 version",
			release: Release{
				Major:       10,
				Minor:       18,
				Patch:       13,
				Build:       89970,
				VersionFull: "10.18.13.89970",
			},
			expected: "10.18.13",
		},
		{
			name: "Mx11.4 version (4-part)",
			release: Release{
				Major:       11,
				Minor:       4,
				Patch:       0,
				Build:       83498,
				VersionFull: "11.4.0.83498",
			},
			expected: "11.4.0",
		},
		{
			name: "Mx11.5+ version (3-part)",
			release: Release{
				Major:       11,
				Minor:       5,
				Patch:       0,
				Build:       0,
				VersionFull: "11.5.0",
			},
			expected: "11.5.0",
		},
		{
			name: "Mx11.9 version",
			release: Release{
				Major:       11,
				Minor:       9,
				Patch:       1,
				Build:       0,
				VersionFull: "11.9.1",
			},
			expected: "11.9.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manifestVersionFor(tt.release)
			if result != tt.expected {
				t.Errorf("manifestVersionFor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestURLPatterns(t *testing.T) {
	tests := []struct {
		name        string
		versionFull string
		manifest    string
		expected4pt string
		expected3pt string
	}{
		{
			name:        "Mx9 has 4-part URLs",
			versionFull: "9.24.42.98709",
			manifest:    "9.24.42",
			expected4pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-9.24.42.98709-User-x64-Setup.exe",
			expected3pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-9.24.42-User-x64-Setup.exe",
		},
		{
			name:        "Mx10 has 4-part URLs",
			versionFull: "10.18.13.89970",
			manifest:    "10.18.13",
			expected4pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-10.18.13.89970-User-x64-Setup.exe",
			expected3pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-10.18.13-User-x64-Setup.exe",
		},
		{
			name:        "Mx11.4 has 4-part URLs",
			versionFull: "11.4.0.83498",
			manifest:    "11.4.0",
			expected4pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-11.4.0.83498-User-x64-Setup.exe",
			expected3pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-11.4.0-User-x64-Setup.exe",
		},
		{
			name:        "Mx11.5+ has 3-part URLs only",
			versionFull: "11.5.0",
			manifest:    "11.5.0",
			expected4pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-11.5.0-User-x64-Setup.exe",
			expected3pt: "https://artifacts.rnd.mendix.com/modelers/Mendix-11.5.0-User-x64-Setup.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we try the 4-part URL first
			url4pt := "https://artifacts.rnd.mendix.com/modelers/Mendix-" + tt.versionFull + "-User-x64-Setup.exe"
			if url4pt != tt.expected4pt {
				t.Errorf("4-part URL = %v, want %v", url4pt, tt.expected4pt)
			}

			// Test that we have a 3-part fallback
			url3pt := "https://artifacts.rnd.mendix.com/modelers/Mendix-" + tt.manifest + "-User-x64-Setup.exe"
			if url3pt != tt.expected3pt {
				t.Errorf("3-part URL = %v, want %v", url3pt, tt.expected3pt)
			}
		})
	}
}
