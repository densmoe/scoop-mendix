package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	bucketDir := flag.String("bucket-dir", "../../bucket", "Directory for Scoop bucket manifests")
	skipSHA := flag.Bool("skip-sha", false, "Skip SHA256 computation")
	dryRun := flag.Bool("dry-run", false, "Preview only")
	versionTypes := flag.String("version-types", "LTS,MTS,Stable", "Comma-separated version types")
	minMajor := flag.Int("min-major", 10, "Minimum major version")
	maxVersions := flag.Int("max-versions", 0, "Limit to first N versions (0 = all)")
	workers := flag.Int("workers", 5, "Number of parallel workers")
	flag.Parse()

	types := strings.Split(*versionTypes, ",")

	client, err := NewMarketplaceClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	releases, err := client.FetchReleases(types, *minMajor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching releases: %v\n", err)
		os.Exit(1)
	}

	// Filter out versions that already have complete manifests
	var needsManifests []Release
	for _, release := range releases {
		manifestVersion := manifestVersionFor(release)
		manifestPath := filepath.Join(*bucketDir, fmt.Sprintf("mendix-studio-pro-%s.json", manifestVersion))
		if !manifestComplete(manifestPath) {
			needsManifests = append(needsManifests, release)
		}
	}

	if *maxVersions > 0 && len(needsManifests) > *maxVersions {
		needsManifests = needsManifests[:*maxVersions]
	}

	fmt.Printf("Found %d total releases, %d need manifests, processing %d with %d workers\n",
		len(releases), len(needsManifests), len(needsManifests), *workers)

	releases = needsManifests

	var wg sync.WaitGroup
	releaseChan := make(chan Release, len(releases))
	resultChan := make(chan string, len(releases))

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for release := range releaseChan {
				result := processRelease(release, *bucketDir, *skipSHA, *dryRun)
				resultChan <- result
			}
		}()
	}

	go func() {
		for _, release := range releases {
			releaseChan <- release
		}
		close(releaseChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		fmt.Println(result)
	}

	// Generate alias manifests (e.g., mendix-studio-pro-10.json -> latest 10.x)
	if !*dryRun {
		fmt.Println("\nGenerating alias manifests...")
		aliasCount := generateAliases(*bucketDir)
		fmt.Printf("Generated %d alias manifest(s)\n", aliasCount)
	}
}

func processRelease(release Release, bucketDir string, skipSHA, dryRun bool) string {
	manifestVersion := manifestVersionFor(release)
	manifestPath := filepath.Join(bucketDir, fmt.Sprintf("mendix-studio-pro-%s.json", manifestVersion))
	machinePath := filepath.Join(bucketDir, fmt.Sprintf("mendix-studio-pro-%s-machine.json", manifestVersion))

	// Check if both user and machine manifests are complete
	userComplete := manifestComplete(manifestPath)
	machineComplete := manifestComplete(machinePath)

	if userComplete && machineComplete {
		return fmt.Sprintf("Skipping %s (complete)", manifestVersion)
	}

	// Try full version first, fall back to 3-part for Mx11.5+
	shortRelease := release
	shortRelease.VersionFull = manifestVersion

	// User-scope installers
	userX64URL := fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-x64-Setup.exe", release.VersionFull)
	userARM64URL := fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-arm64-Setup.exe", release.VersionFull)

	// Machine-scope installer (x64 only, no arch suffix)
	machineURL := fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-Setup.exe", release.VersionFull)

	// Check if URLs exist, fall back to 3-part if needed
	if !urlExists(userX64URL) {
		if release.VersionFull != manifestVersion {
			userX64URL = fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-x64-Setup.exe", manifestVersion)
			userARM64URL = fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-arm64-Setup.exe", manifestVersion)
			machineURL = fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-Setup.exe", manifestVersion)
			if !urlExists(userX64URL) {
				return fmt.Sprintf("%s: user x64 installer not found", manifestVersion)
			}
		} else {
			return fmt.Sprintf("%s: user x64 installer not found", manifestVersion)
		}
	}

	// Fetch SHA256 hashes
	var userX64SHA, userARM64SHA, machineSHA string
	if skipSHA {
		userX64SHA = "SHA256_PLACEHOLDER"
		userARM64SHA = "SHA256_PLACEHOLDER"
		machineSHA = "SHA256_PLACEHOLDER"
	} else {
		var err error

		// User x64 is required
		if !userComplete {
			userX64SHA, err = fetchSHA256(userX64URL)
			if err != nil {
				return fmt.Sprintf("%s: failed to get user x64 hash: %v", manifestVersion, err)
			}

			// ARM64 is optional
			userARM64SHA, err = fetchSHA256(userARM64URL)
			if err != nil {
				userARM64SHA = "" // ARM64 installer doesn't exist for this version
			}
		}

		// Machine installer
		if !machineComplete {
			machineSHA, err = fetchSHA256(machineURL)
			if err != nil {
				return fmt.Sprintf("%s: failed to get machine hash: %v", manifestVersion, err)
			}
		}
	}

	if dryRun {
		return fmt.Sprintf("%s: would create manifest", manifestVersion)
	}

	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		return fmt.Sprintf("%s: failed to create directory: %v", manifestVersion, err)
	}

	results := []string{}

	// Write user-scope manifest
	if !userComplete {
		userData := ScoopManifestData{
			Version:         manifestVersion,
			UserX64URL:      userX64URL,
			UserX64SHA256:   userX64SHA,
			UserARM64URL:    userARM64URL,
			UserARM64SHA256: userARM64SHA,
		}

		f, err := os.Create(manifestPath)
		if err != nil {
			return fmt.Sprintf("%s: failed to create user manifest: %v", manifestVersion, err)
		}

		if err := scoopManifestTemplate.Execute(f, userData); err != nil {
			f.Close()
			return fmt.Sprintf("%s: failed to write user manifest: %v", manifestVersion, err)
		}
		f.Close()
		results = append(results, "user")
	}

	// Write machine-scope manifest
	if !machineComplete {
		machineData := ScoopMachineManifestData{
			Version:       manifestVersion,
			MachineURL:    machineURL,
			MachineSHA256: machineSHA,
		}

		f, err := os.Create(machinePath)
		if err != nil {
			return fmt.Sprintf("%s: failed to create machine manifest: %v", manifestVersion, err)
		}

		if err := scoopMachineManifestTemplate.Execute(f, machineData); err != nil {
			f.Close()
			return fmt.Sprintf("%s: failed to write machine manifest: %v", manifestVersion, err)
		}
		f.Close()
		results = append(results, "machine")
	}

	if len(results) == 0 {
		return fmt.Sprintf("%s: already complete", manifestVersion)
	}

	return fmt.Sprintf("%s: created %s", manifestVersion, strings.Join(results, " + "))
}

func manifestVersionFor(r Release) string {
	return fmt.Sprintf("%d.%d.%d", r.Major, r.Minor, r.Patch)
}

func manifestComplete(manifestPath string) bool {
	if _, err := os.Stat(manifestPath); err != nil {
		return false
	}

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return false
	}

	s := string(content)
	if strings.Contains(s, "PLACEHOLDER") {
		return false
	}
	if !strings.Contains(s, "\"hash\"") {
		return false
	}

	return true
}

func urlExists(url string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check both status code and content length
	// Mendix CDN returns 200 with tiny error pages for missing files
	if resp.StatusCode != 200 {
		return false
	}

	// Valid installers are at least 100MB, anything smaller is likely an error page
	if resp.ContentLength > 0 && resp.ContentLength < 100*1024*1024 {
		return false
	}

	return true
}

func fetchSHA256(url string) (string, error) {
	// Try the .sha256 sidecar file first (fast, available for newer versions)
	sha, err := fetchSHA256FromSidecar(url)
	if err == nil {
		return sha, nil
	}

	// Fall back to downloading the installer and computing the hash
	fmt.Printf("  computing SHA256 for %s (no .sha256 file)\n", filepath.Base(url))
	return computeSHA256FromURL(url)
}

func fetchSHA256FromSidecar(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url + ".sha256")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("SHA256 file not found (HTTP %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(strings.TrimSpace(string(body)))
	if len(parts) == 0 {
		return "", fmt.Errorf("empty SHA256 file")
	}

	return strings.ToLower(parts[0]), nil
}

func computeSHA256FromURL(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed (HTTP %d)", resp.StatusCode)
	}

	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", fmt.Errorf("download interrupted: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type versionInfo struct {
	major, minor, patch int
	fullVersion         string
	manifestPath        string
}

func generateAliases(bucketDir string) int {
	// Find all existing versioned manifests
	pattern := filepath.Join(bucketDir, "mendix-studio-pro-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding manifests: %v\n", err)
		return 0
	}

	var versions []versionInfo
	for _, path := range matches {
		name := filepath.Base(path)

		// Skip machine manifests (we only alias user-scope)
		if strings.Contains(name, "-machine.json") {
			continue
		}

		// Skip alias manifests (those without patch version)
		versionStr := strings.TrimPrefix(name, "mendix-studio-pro-")
		versionStr = strings.TrimSuffix(versionStr, ".json")

		parts := strings.Split(versionStr, ".")
		if len(parts) != 3 {
			continue // Skip non-standard names
		}

		var major, minor, patch int
		if _, err := fmt.Sscanf(versionStr, "%d.%d.%d", &major, &minor, &patch); err != nil {
			continue
		}

		versions = append(versions, versionInfo{
			major:        major,
			minor:        minor,
			patch:        patch,
			fullVersion:  versionStr,
			manifestPath: path,
		})
	}

	if len(versions) == 0 {
		return 0
	}

	// Find latest for each category
	latestOverall := versions[0]
	latestByMajor := make(map[int]versionInfo)
	latestByMinor := make(map[string]versionInfo)

	for _, v := range versions {
		// Latest overall
		if versionCompare(v, latestOverall) > 0 {
			latestOverall = v
		}

		// Latest for major version (e.g., 10, 11)
		if existing, ok := latestByMajor[v.major]; !ok || versionCompare(v, existing) > 0 {
			latestByMajor[v.major] = v
		}

		// Latest for minor version (e.g., 10.24, 11.9)
		minorKey := fmt.Sprintf("%d.%d", v.major, v.minor)
		if existing, ok := latestByMinor[minorKey]; !ok || versionCompare(v, existing) > 0 {
			latestByMinor[minorKey] = v
		}
	}

	aliasCount := 0

	// Create "mendix-studio-pro.json" -> latest overall
	if err := createAliasManifest(bucketDir, "mendix-studio-pro.json", latestOverall.manifestPath); err != nil {
		fmt.Fprintf(os.Stderr, "  Error creating latest alias: %v\n", err)
	} else {
		fmt.Printf("  mendix-studio-pro.json -> %s\n", latestOverall.fullVersion)
		aliasCount++
	}

	// Create "mendix-studio-pro-N.json" -> latest N.x
	for major, v := range latestByMajor {
		aliasName := fmt.Sprintf("mendix-studio-pro-%d.json", major)
		if err := createAliasManifest(bucketDir, aliasName, v.manifestPath); err != nil {
			fmt.Fprintf(os.Stderr, "  Error creating major alias %d: %v\n", major, err)
		} else {
			fmt.Printf("  mendix-studio-pro-%d.json -> %s\n", major, v.fullVersion)
			aliasCount++
		}
	}

	// Create "mendix-studio-pro-N.M.json" -> latest N.M.x
	for minorKey, v := range latestByMinor {
		aliasName := fmt.Sprintf("mendix-studio-pro-%s.json", minorKey)
		if err := createAliasManifest(bucketDir, aliasName, v.manifestPath); err != nil {
			fmt.Fprintf(os.Stderr, "  Error creating minor alias %s: %v\n", minorKey, err)
		} else {
			fmt.Printf("  mendix-studio-pro-%s.json -> %s\n", minorKey, v.fullVersion)
			aliasCount++
		}
	}

	return aliasCount
}

func versionCompare(a, b versionInfo) int {
	if a.major != b.major {
		return a.major - b.major
	}
	if a.minor != b.minor {
		return a.minor - b.minor
	}
	return a.patch - b.patch
}

func createAliasManifest(bucketDir, aliasName, sourcePath string) error {
	// Read the source manifest
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	// Write to alias path
	aliasPath := filepath.Join(bucketDir, aliasName)
	return os.WriteFile(aliasPath, content, 0644)
}
