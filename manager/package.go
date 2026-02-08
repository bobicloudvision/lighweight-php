package manager

import (
	"fmt"
	"os/exec"
	"strings"

	"lightweight-php/db"
	"lightweight-php/system"
)

type PackageManager struct {
	osFamily system.OSFamily
	db       *db.Database
}

func NewPackageManager() (*PackageManager, error) {
	database, err := db.NewDatabase("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	detector := system.NewOSDetector()
	osFamily, _ := detector.Detect()

	return &PackageManager{
		osFamily: osFamily,
		db:       database,
	}, nil
}

func (pm *PackageManager) InstallPHP(version string) error {
	// Normalize version (e.g., "8.2" -> "82" for package names)
	versionNum := strings.ReplaceAll(version, ".", "")

	if pm.osFamily == system.OSRHEL {
		return pm.installPHPRHEL(version, versionNum)
	} else {
		return pm.installPHPDebian(version, versionNum)
	}
}

func (pm *PackageManager) installPHPRHEL(version, versionNum string) error {
	// Check if Remi repository is installed
	if err := pm.ensureRemiRepo(); err != nil {
		return fmt.Errorf("failed to setup Remi repository: %w", err)
	}

	// Enable Remi repository for the specific PHP version
	enableCmd := exec.Command("yum", "config-manager", "--enable", fmt.Sprintf("remi-php%s", versionNum))
	if enableCmd.Run() != nil {
		// Try with dnf if yum is not available
		enableCmd = exec.Command("dnf", "config-manager", "--enable", fmt.Sprintf("remi-php%s", versionNum))
		if err := enableCmd.Run(); err != nil {
			return fmt.Errorf("failed to enable Remi repository: %w", err)
		}
	}

	// Install PHP and PHP-FPM
	packages := []string{
		fmt.Sprintf("php%s", versionNum),
		fmt.Sprintf("php%s-php-fpm", versionNum),
		fmt.Sprintf("php%s-php-cli", versionNum),
		fmt.Sprintf("php%s-php-common", versionNum),
	}

	var installCmd *exec.Cmd
	if pm.hasCommand("dnf") {
		installCmd = exec.Command("dnf", append([]string{"install", "-y"}, packages...)...)
	} else {
		installCmd = exec.Command("yum", append([]string{"install", "-y"}, packages...)...)
	}

	installCmd.Stdout = nil
	installCmd.Stderr = nil
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install PHP packages: %w", err)
	}

	// Enable and start PHP-FPM service
	serviceName := fmt.Sprintf("php%s-php-fpm", versionNum)
	enableService := exec.Command("systemctl", "enable", serviceName)
	enableService.Run()

	startService := exec.Command("systemctl", "start", serviceName)
	if err := startService.Run(); err != nil {
		return fmt.Errorf("failed to start PHP-FPM service: %w", err)
	}

	// Save to database
	if err := pm.db.CreatePHPVersion(version, "remi", "rhel"); err != nil {
		// Log error but don't fail installation
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (pm *PackageManager) installPHPDebian(version, versionNum string) error {
	// Update package list
	updateCmd := exec.Command("apt-get", "update")
	updateCmd.Stdout = nil
	updateCmd.Stderr = nil
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install prerequisites
	prereqCmd := exec.Command("apt-get", "install", "-y", "software-properties-common", "apt-transport-https", "lsb-release", "ca-certificates", "gnupg2")
	prereqCmd.Stdout = nil
	prereqCmd.Stderr = nil
	prereqCmd.Run()

	// Add ondrej/php PPA (standard for multiple PHP versions on Debian/Ubuntu)
	// Note: Remi repository is primarily for RHEL, so we use ondrej PPA for Debian
	addRepoScript := `add-apt-repository -y ppa:ondrej/php 2>/dev/null || echo "deb https://ppa.launchpadcontent.net/ondrej/php/ubuntu $(lsb_release -sc) main" > /etc/apt/sources.list.d/ondrej-php.list`
	addRepoCmd := exec.Command("sh", "-c", addRepoScript)
	addRepoCmd.Run()

	// Add GPG key
	addKeyScript := `curl -fsSL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x14AA40EC0831756756D7F66C4F4EA0AAE5267A6C" | gpg --dearmor -o /etc/apt/trusted.gpg.d/ondrej-php.gpg 2>/dev/null || apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 14AA40EC0831756756D7F66C4F4EA0AAE5267A6C 2>/dev/null`
	addKeyCmd := exec.Command("sh", "-c", addKeyScript)
	addKeyCmd.Run()

	// Update again after adding repository
	updateCmd.Run()

	// Install PHP version
	packages := []string{
		fmt.Sprintf("php%s", version),
		fmt.Sprintf("php%s-fpm", version),
		fmt.Sprintf("php%s-cli", version),
		fmt.Sprintf("php%s-common", version),
	}

	installCmd := exec.Command("apt-get", "install", "-y")
	installCmd.Args = append(installCmd.Args, packages...)
	installCmd.Stdout = nil
	installCmd.Stderr = nil
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install PHP packages: %w", err)
	}

	// Enable and start PHP-FPM service
	serviceName := fmt.Sprintf("php%s-fpm", version)
	enableService := exec.Command("systemctl", "enable", serviceName)
	enableService.Run()

	startService := exec.Command("systemctl", "start", serviceName)
	if err := startService.Run(); err != nil {
		return fmt.Errorf("failed to start PHP-FPM service: %w", err)
	}

	// Save to database
	if err := pm.db.CreatePHPVersion(version, "ondrej", "debian"); err != nil {
		// Log error but don't fail installation
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (pm *PackageManager) ensureRemiRepo() error {
	if pm.osFamily != system.OSRHEL {
		return nil // Not needed for Debian
	}

	// Check if remi-release is installed
	checkCmd := exec.Command("rpm", "-q", "remi-release")
	if checkCmd.Run() == nil {
		return nil // Already installed
	}

	// Detect RHEL version
	releaseCmd := exec.Command("rpm", "-q", "--qf", "%{VERSION}", "redhat-release-server")
	output, err := releaseCmd.Output()
	if err != nil {
		// Try alternative
		releaseCmd = exec.Command("cat", "/etc/redhat-release")
		output, err = releaseCmd.Output()
	}

	var repoURL string
	releaseStr := strings.TrimSpace(string(output))
	if strings.Contains(releaseStr, "8") || strings.Contains(releaseStr, "9") {
		repoURL = "https://rpms.remirepo.net/enterprise/remi-release-8.rpm"
	} else if strings.Contains(releaseStr, "7") {
		repoURL = "https://rpms.remirepo.net/enterprise/remi-release-7.rpm"
	} else {
		repoURL = "https://rpms.remirepo.net/enterprise/remi-release-8.rpm" // Default
	}

	installCmd := exec.Command("yum", "install", "-y", repoURL)
	if installCmd.Run() != nil {
		installCmd = exec.Command("dnf", "install", "-y", repoURL)
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install Remi repository: %w", err)
		}
	}

	return nil
}

func (pm *PackageManager) ListInstalledPHP() ([]string, error) {
	// Try to get from database first
	dbVersions, err := pm.db.ListPHPVersions()
	if err == nil && len(dbVersions) > 0 {
		versions := make([]string, 0, len(dbVersions))
		for _, v := range dbVersions {
			if v.Status == "active" {
				versions = append(versions, v.Version)
			}
		}
		if len(versions) > 0 {
			return versions, nil
		}
	}

	// Fallback to system detection
	var versions []string

	if pm.osFamily == system.OSRHEL {
		// Check for installed PHP packages
		cmd := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\n")
		if pm.hasCommand("dnf") {
			cmd = exec.Command("dnf", "list", "installed", "php*-php-fpm")
		} else {
			cmd = exec.Command("yum", "list", "installed", "php*-php-fpm")
		}
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "php") && strings.Contains(line, "fpm") {
					// Extract version from package name
					parts := strings.Fields(line)
					if len(parts) > 0 {
						versions = append(versions, parts[0])
					}
				}
			}
		}
	} else {
		// Check /etc/php directory
		entries, err := exec.Command("ls", "/etc/php").Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(entries)), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, " ") {
					versions = append(versions, line)
				}
			}
		}
	}

	return versions, nil
}

func (pm *PackageManager) ListAvailablePHP() ([]string, error) {
	if pm.osFamily == system.OSRHEL {
		return pm.listAvailablePHPRHEL()
	} else {
		return pm.listAvailablePHPDebian()
	}
}

func (pm *PackageManager) listAvailablePHPRHEL() ([]string, error) {
	// Ensure Remi repo is available
	if err := pm.ensureRemiRepo(); err != nil {
		// If repo setup fails, return common versions
		return []string{"8.3", "8.2", "8.1", "8.0", "7.4"}, nil
	}

	var versions []string
	var cmd *exec.Cmd

	// Try to list available remi-php repositories
	if pm.hasCommand("dnf") {
		cmd = exec.Command("dnf", "repolist", "--enabled", "--disabled")
	} else {
		cmd = exec.Command("yum", "repolist", "all")
	}

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "remi-php") {
				// Extract version from repository name like "remi-php82"
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "remi-php") {
						versionNum := strings.TrimPrefix(part, "remi-php")
						if len(versionNum) >= 2 {
							// Convert "82" to "8.2"
							if len(versionNum) == 2 {
								version := fmt.Sprintf("%s.%s", versionNum[0:1], versionNum[1:])
								versions = append(versions, version)
							} else if len(versionNum) == 3 {
								// Handle versions like "81" or "74"
								version := fmt.Sprintf("%s.%s", versionNum[0:1], versionNum[1:])
								versions = append(versions, version)
							}
						}
					}
				}
			}
		}
	}

	// If we didn't find any, try querying available packages
	if len(versions) == 0 {
		// Query for available PHP-FPM packages
		if pm.hasCommand("dnf") {
			cmd = exec.Command("dnf", "list", "available", "php*-php-fpm")
		} else {
			cmd = exec.Command("yum", "list", "available", "php*-php-fpm")
		}
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "php") && strings.Contains(line, "fpm") {
					// Extract version from package name
					parts := strings.Fields(line)
					if len(parts) > 0 {
						pkgName := parts[0]
						// Extract version from package name like "php82-php-fpm"
						if strings.HasPrefix(pkgName, "php") {
							versionNum := strings.TrimPrefix(pkgName, "php")
							if idx := strings.Index(versionNum, "-php-fpm"); idx > 0 {
								versionNum = versionNum[:idx]
								if len(versionNum) >= 2 {
									version := fmt.Sprintf("%s.%s", versionNum[0:1], versionNum[1:])
									// Avoid duplicates
									found := false
									for _, v := range versions {
										if v == version {
											found = true
											break
										}
									}
									if !found {
										versions = append(versions, version)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// If still no versions found, return common Remi versions
	if len(versions) == 0 {
		versions = []string{"8.3", "8.2", "8.1", "8.0", "7.4", "7.3", "7.2"}
	}

	return versions, nil
}

func (pm *PackageManager) listAvailablePHPDebian() ([]string, error) {
	var versions []string

	// Try to query available PHP versions from repository
	cmd := exec.Command("apt-cache", "search", "--names-only", "^php[0-9]\\.[0-9]-fpm$")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "php") && strings.Contains(line, "-fpm") {
				// Extract version from package name like "php8.2-fpm"
				parts := strings.Fields(line)
				if len(parts) > 0 {
					pkgName := parts[0]
					// Remove "php" prefix and "-fpm" suffix
					if strings.HasPrefix(pkgName, "php") && strings.HasSuffix(pkgName, "-fpm") {
						version := strings.TrimPrefix(pkgName, "php")
						version = strings.TrimSuffix(version, "-fpm")
						// Validate version format (X.Y)
						if len(version) >= 3 && strings.Contains(version, ".") {
							// Avoid duplicates
							found := false
							for _, v := range versions {
								if v == version {
									found = true
									break
								}
							}
							if !found {
								versions = append(versions, version)
							}
						}
					}
				}
			}
		}
	}

	// If no versions found, try alternative method
	if len(versions) == 0 {
		cmd = exec.Command("apt-cache", "pkgnames", "php")
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "php") && strings.Contains(line, "-fpm") {
					// Extract version
					parts := strings.Split(line, "-")
					if len(parts) >= 2 && strings.HasPrefix(parts[0], "php") {
						version := strings.TrimPrefix(parts[0], "php")
						if len(version) >= 3 && strings.Contains(version, ".") {
							found := false
							for _, v := range versions {
								if v == version {
									found = true
									break
								}
							}
							if !found {
								versions = append(versions, version)
							}
						}
					}
				}
			}
		}
	}

	// If still no versions found, return common ondrej PPA versions
	if len(versions) == 0 {
		versions = []string{"8.3", "8.2", "8.1", "8.0", "7.4", "7.3", "7.2", "7.1", "7.0", "5.6"}
	}

	return versions, nil
}

func (pm *PackageManager) hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
