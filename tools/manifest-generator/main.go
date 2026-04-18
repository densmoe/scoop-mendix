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
}

func processRelease(release Release, bucketDir string, skipSHA, dryRun bool) string {
	manifestVersion := manifestVersionFor(release)
	manifestPath := filepath.Join(bucketDir, fmt.Sprintf("mendix-studio-pro-%s.json", manifestVersion))

	if manifestComplete(manifestPath) {
		return fmt.Sprintf("Skipping %s (complete)", manifestVersion)
	}

	// Try full version first, fall back to 3-part for Mx11.5+
	shortRelease := release
	shortRelease.VersionFull = manifestVersion

	userX64URL := fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-x64-Setup.exe", release.VersionFull)
	userARM64URL := fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-arm64-Setup.exe", release.VersionFull)

	// Check if URLs exist, fall back to 3-part if needed
	if !urlExists(userX64URL) {
		if release.VersionFull != manifestVersion {
			userX64URL = fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-x64-Setup.exe", manifestVersion)
			userARM64URL = fmt.Sprintf("https://artifacts.rnd.mendix.com/modelers/Mendix-%s-User-arm64-Setup.exe", manifestVersion)
			if !urlExists(userX64URL) {
				return fmt.Sprintf("%s: user x64 installer not found", manifestVersion)
			}
		} else {
			return fmt.Sprintf("%s: user x64 installer not found", manifestVersion)
		}
	}

	// x64 is required, ARM64 is optional
	var userX64SHA, userARM64SHA string
	if skipSHA {
		userX64SHA = "SHA256_PLACEHOLDER"
		userARM64SHA = "SHA256_PLACEHOLDER"
	} else {
		var err error
		userX64SHA, err = fetchSHA256(userX64URL)
		if err != nil {
			return fmt.Sprintf("%s: failed to get x64 hash: %v", manifestVersion, err)
		}

		// ARM64 is optional - if it doesn't exist, we'll handle it in the template
		userARM64SHA, err = fetchSHA256(userARM64URL)
		if err != nil {
			userARM64SHA = "" // ARM64 installer doesn't exist for this version
		}
	}

	if dryRun {
		return fmt.Sprintf("%s: would create manifest", manifestVersion)
	}

	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		return fmt.Sprintf("%s: failed to create directory: %v", manifestVersion, err)
	}

	data := ScoopManifestData{
		Version:         manifestVersion,
		UserX64URL:      userX64URL,
		UserX64SHA256:   userX64SHA,
		UserARM64URL:    userARM64URL,
		UserARM64SHA256: userARM64SHA,
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Sprintf("%s: failed to create manifest: %v", manifestVersion, err)
	}
	defer f.Close()

	if err := scoopManifestTemplate.Execute(f, data); err != nil {
		return fmt.Sprintf("%s: failed to write manifest: %v", manifestVersion, err)
	}

	return fmt.Sprintf("%s: created manifest", manifestVersion)
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
