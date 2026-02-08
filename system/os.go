package system

import (
	"os"
	"os/exec"
	"strings"
)

type OSFamily string

const (
	OSRHEL   OSFamily = "rhel"
	OSDebian OSFamily = "debian"
)

type OSDetector struct{}

func NewOSDetector() *OSDetector {
	return &OSDetector{}
}

func (d *OSDetector) Detect() (OSFamily, error) {
	// Check for /etc/redhat-release (RHEL, CentOS, Rocky, etc.)
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return OSRHEL, nil
	}

	// Check for /etc/debian_version (Debian, Ubuntu, etc.)
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return OSDebian, nil
	}

	// Try to detect via lsb_release
	output, err := exec.Command("lsb_release", "-is").Output()
	if err == nil {
		distro := strings.ToLower(strings.TrimSpace(string(output)))
		if strings.Contains(distro, "redhat") || strings.Contains(distro, "centos") ||
			strings.Contains(distro, "rocky") || strings.Contains(distro, "alma") ||
			strings.Contains(distro, "fedora") {
			return OSRHEL, nil
		}
		if strings.Contains(distro, "debian") || strings.Contains(distro, "ubuntu") {
			return OSDebian, nil
		}
	}

	// Default to RHEL if uncertain
	return OSRHEL, nil
}
