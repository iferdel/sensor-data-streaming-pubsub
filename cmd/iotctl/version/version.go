package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"

	"golang.org/x/mod/semver"
)

const repoOwner = "iferdel"
const repoName = "sensor-data-streaming-pubsub"

type VersionInfo struct {
	CurrentVersion   string
	LatestVersion    string
	IsOutdated       bool
	IsUpdateRequired bool
	FailtedToFetch   error
}

func FetchUpdateInfo(CurrentVersion string) VersionInfo {
	latest, err := getLatestVersion()
	if err != nil {
		return VersionInfo{
			FailtedToFetch: err,
		}
	}
	isUpdateRequired := isUpdateRequired(CurrentVersion, latest)
	isOutdated := isOutdated(CurrentVersion, latest)
	return VersionInfo{
		CurrentVersion:   CurrentVersion,
		LatestVersion:    latest,
		IsOutdated:       isOutdated,
		IsUpdateRequired: isUpdateRequired,
		FailtedToFetch:   nil,
	}
}

func (v *VersionInfo) PromptUpdateIfAvailable() {
	if v.IsOutdated {
		fmt.Fprintln(os.Stderr, "A new version of the iot CLI is available")
		fmt.Fprintln(os.Stderr, "Please run the following command to update:")
		fmt.Fprintln(os.Stderr, "  iotctl upgrade")
		fmt.Fprintln(os.Stderr, "or")
		fmt.Fprintf(os.Stderr, "  go install github.com/%s/%s@%s\n\n", repoOwner, repoName, v.LatestVersion)
	}
}

func isUpdateRequired(current, latest string) bool {
	latestMajorMinor := semver.MajorMinor(latest)
	currentMajorMinor := semver.MajorMinor(current)
	return semver.Compare(currentMajorMinor, latestMajorMinor) < 0
}

func isOutdated(current, latest string) bool {
	return semver.Compare(current, latest) < 0
}

func getLatestVersion() (string, error) {
	// it uses a default goproxy unless another one is used in the current system.
	goproxyDefault := "https://proxy.golang.org"
	goproxy := goproxyDefault
	cmd := exec.Command("go", "env", "GOPROXY")
	output, err := cmd.Output()
	if err == nil {
		goproxy = strings.TrimSpace(string(output))
	}

	// if in the goproxy command we have a different url than the default, append the default to this slice
	proxies := strings.Split(goproxy, ",")
	if !slices.Contains(proxies, goproxyDefault) {
		proxies = append(proxies, goproxyDefault)
	}

	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		proxy = strings.TrimRight(proxy, "/")
		if proxy == "direct" || proxy == "off" {
			continue
		}

		// with each new release (version) we create a new latest. So its a one-to-one relationship
		url := fmt.Sprintf("%s/github.com/%s/%s/@latest", proxy, repoOwner, repoName)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var version struct{ Version string }
		if err = json.Unmarshal(body, &version); err != nil {
			continue
		}

		return version.Version, nil
	}

	return "", fmt.Errorf("Failed to fetch latest version")

}
