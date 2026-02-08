package provider

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"lightweight-php/db"
	"lightweight-php/system"
)

// LiteSpeedProvider implements PHPProvider for LiteSpeed PHP (lsphp)
type LiteSpeedProvider struct {
	db       *db.Database
	osFamily system.OSFamily
}

func NewLiteSpeedProvider(database *db.Database, osFamily system.OSFamily) (*LiteSpeedProvider, error) {
	return &LiteSpeedProvider{
		db:       database,
		osFamily: osFamily,
	}, nil
}

func (p *LiteSpeedProvider) GetProviderType() string {
	return string(ProviderLiteSpeed)
}

func (p *LiteSpeedProvider) GetServiceName(version string) string {
	// LiteSpeed uses lsws service, not individual PHP services
	return "lsws"
}

func (p *LiteSpeedProvider) GetSocketPath(username, version string) string {
	// LiteSpeed socket path format: /tmp/lsphp82-username.sock
	versionNum := strings.ReplaceAll(version, ".", "")
	return fmt.Sprintf("/tmp/lsphp%s-%s.sock", versionNum, username)
}

func (p *LiteSpeedProvider) GetConfigPath(username, version string) string {
	// LiteSpeed configuration path
	return filepath.Join("/usr/local/lsws/conf", fmt.Sprintf("%s-%s.conf", username, version))
}

func (p *LiteSpeedProvider) InstallPHP(version string) error {
	// Validate minimum PHP version (7.4)
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid PHP version format: %s", version)
	}
	
	major, minor := parts[0], parts[1]
	if major < "7" || (major == "7" && minor < "4") {
		return fmt.Errorf("PHP version %s is not supported. Minimum version is 7.4", version)
	}

	versionNum := strings.ReplaceAll(version, ".", "")

	if p.osFamily == system.OSRHEL {
		return p.installLSPHPRHEL(version, versionNum)
	} else {
		return p.installLSPHPDebian(version, versionNum)
	}
}

func (p *LiteSpeedProvider) installLSPHPRHEL(version, versionNum string) error {
	// LiteSpeed PHP packages for RHEL are typically: lsphp82, lsphp82-common, lsphp82-process, etc.
	packages := []string{
		fmt.Sprintf("lsphp%s", versionNum),
		fmt.Sprintf("lsphp%s-common", versionNum),
		fmt.Sprintf("lsphp%s-process", versionNum),
	}

	var installCmd *exec.Cmd
	if p.hasCommand("dnf") {
		installCmd = exec.Command("dnf", append([]string{"install", "-y"}, packages...)...)
	} else {
		installCmd = exec.Command("yum", append([]string{"install", "-y"}, packages...)...)
	}

	installCmd.Stdout = nil
	installCmd.Stderr = nil
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install LiteSpeed PHP packages: %w", err)
	}

	// Save to database
	if err := p.db.CreatePHPVersion(version, "lsphp", "rhel"); err != nil {
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (p *LiteSpeedProvider) installLSPHPDebian(version, versionNum string) error {
	// LiteSpeed PHP packages for Debian
	packages := []string{
		fmt.Sprintf("lsphp%s", versionNum),
		fmt.Sprintf("lsphp%s-common", versionNum),
		fmt.Sprintf("lsphp%s-process", versionNum),
	}

	// Update package list
	updateCmd := exec.Command("apt-get", "update")
	updateCmd.Stdout = nil
	updateCmd.Stderr = nil
	updateCmd.Run()

	installCmd := exec.Command("apt-get", "install", "-y")
	installCmd.Args = append(installCmd.Args, packages...)
	installCmd.Stdout = nil
	installCmd.Stderr = nil
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install LiteSpeed PHP packages: %w", err)
	}

	// Save to database
	if err := p.db.CreatePHPVersion(version, "lsphp", "debian"); err != nil {
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (p *LiteSpeedProvider) ListInstalledPHP() ([]string, error) {
	// Try to get from database first
	dbVersions, err := p.db.ListPHPVersions()
	if err == nil && dbVersions != nil && len(dbVersions) > 0 {
		versions := make([]string, 0, len(dbVersions))
		for _, v := range dbVersions {
			if v.Status == "active" && v.PackageManager == "lsphp" {
				versions = append(versions, v.Version)
			}
		}
		if len(versions) > 0 {
			return versions, nil
		}
	}

	// Fallback: Check for installed lsphp packages
	versions := make([]string, 0)

	if p.osFamily == system.OSRHEL {
		var cmd *exec.Cmd
		if p.hasCommand("dnf") {
			cmd = exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\n")
		} else {
			cmd = exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\n")
		}
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "lsphp") && strings.Contains(line, "common") {
					// Extract version from package name like "lsphp82-common"
					parts := strings.Split(line, "-")
					if len(parts) > 0 {
						pkgName := parts[0]
						if strings.HasPrefix(pkgName, "lsphp") {
							versionNum := strings.TrimPrefix(pkgName, "lsphp")
							if len(versionNum) >= 2 {
								version := fmt.Sprintf("%s.%s", versionNum[0:1], versionNum[1:])
								versions = append(versions, version)
							}
						}
					}
				}
			}
		}
	} else {
		// Check for installed lsphp packages via dpkg
		cmd := exec.Command("dpkg", "-l")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "lsphp") && strings.Contains(line, "common") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						pkgName := parts[1]
						if strings.HasPrefix(pkgName, "lsphp") {
							parts2 := strings.Split(pkgName, "-")
							if len(parts2) > 0 {
								versionNum := strings.TrimPrefix(parts2[0], "lsphp")
								if len(versionNum) >= 2 {
									version := fmt.Sprintf("%s.%s", versionNum[0:1], versionNum[1:])
									versions = append(versions, version)
								}
							}
						}
					}
				}
			}
		}
	}

	if versions == nil {
		versions = []string{}
	}
	return versions, nil
}

func (p *LiteSpeedProvider) ListAvailablePHP() ([]string, error) {
	// Hardcoded list of available LiteSpeed PHP versions
	versions := []string{
		"8.3",
		"8.2",
		"8.1",
		"8.0",
		"7.4",
	}
	return versions, nil
}

func (p *LiteSpeedProvider) hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
