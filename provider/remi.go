package provider

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"lightweight-php/db"
	"lightweight-php/system"
)

// RemiProvider implements PHPProvider for Remi repository (RHEL) and ondrej PPA (Debian)
type RemiProvider struct {
	db       *db.Database
	osFamily system.OSFamily
}

func NewRemiProvider(database *db.Database, osFamily system.OSFamily) (*RemiProvider, error) {
	return &RemiProvider{
		db:       database,
		osFamily: osFamily,
	}, nil
}

func (p *RemiProvider) GetProviderType() string {
	return string(ProviderRemi)
}

func (p *RemiProvider) GetServiceName(version string) string {
	versionNum := strings.ReplaceAll(version, ".", "")
	if p.osFamily == system.OSRHEL {
		return fmt.Sprintf("php%s-php-fpm", versionNum)
	}
	return fmt.Sprintf("php%s-fpm", version)
}

func (p *RemiProvider) GetSocketPath(username, version string) string {
	if p.osFamily == system.OSRHEL {
		versionNum := strings.ReplaceAll(version, ".", "")
		return fmt.Sprintf("/var/opt/remi/php%s/run/php-fpm/%s.sock", versionNum, username)
	}
	return fmt.Sprintf("/var/run/php/php%s-%s.sock", version, username)
}

func (p *RemiProvider) GetConfigPath(username, version string) string {
	if p.osFamily == system.OSRHEL {
		versionNum := strings.ReplaceAll(version, ".", "")
		return filepath.Join("/etc/opt/remi", fmt.Sprintf("php%s", versionNum), "php-fpm.d", fmt.Sprintf("%s.conf", username))
	}
	return filepath.Join("/etc/php", version, "fpm/pool.d", fmt.Sprintf("%s.conf", username))
}

func (p *RemiProvider) InstallPHP(version string) error {
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
		return p.installPHPRHEL(version, versionNum)
	} else {
		return p.installPHPDebian(version, versionNum)
	}
}

func (p *RemiProvider) installPHPRHEL(version, versionNum string) error {
	// Check if Remi repository is installed
	if err := p.ensureRemiRepo(); err != nil {
		return fmt.Errorf("failed to setup Remi repository: %w", err)
	}

	// Try to enable Remi repository for the specific PHP version (non-fatal if it fails)
	repoName := fmt.Sprintf("remi-php%s", versionNum)
	
	// Check if repository exists before trying to enable it
	var repoExists bool
	if p.hasCommand("dnf") {
		checkCmd := exec.Command("dnf", "repolist", "--all", "--quiet")
		output, err := checkCmd.Output()
		if err == nil {
			repoExists = strings.Contains(string(output), repoName)
		}
	} else {
		checkCmd := exec.Command("yum", "repolist", "all", "-q")
		output, err := checkCmd.Output()
		if err == nil {
			repoExists = strings.Contains(string(output), repoName)
		}
	}
	
	// Only try to enable if repository exists
	if repoExists {
		if p.hasCommand("yum-config-manager") {
			enableCmd := exec.Command("yum-config-manager", "--enable", repoName)
			enableCmd.Stdout = nil
			enableCmd.Stderr = nil
			enableCmd.Run() // Ignore errors
		} else if p.hasCommand("dnf") {
			enableCmd := exec.Command("dnf", "config-manager", "--enable", repoName)
			enableCmd.Stdout = nil
			enableCmd.Stderr = nil
			enableCmd.Run() // Ignore errors
		}
	}

	// Install PHP and PHP-FPM with repository enabled
	packages := []string{
		fmt.Sprintf("php%s-php-fpm", versionNum),
		fmt.Sprintf("php%s-php-cli", versionNum),
		fmt.Sprintf("php%s-php-common", versionNum),
	}

	var installCmd *exec.Cmd
	var stderr bytes.Buffer
	
	// Try installation - use --enablerepo only if repository exists
	if repoExists {
		if p.hasCommand("dnf") {
			installCmd = exec.Command("dnf", append([]string{"install", "-y", fmt.Sprintf("--enablerepo=%s", repoName)}, packages...)...)
		} else {
			installCmd = exec.Command("yum", append([]string{"install", "-y", fmt.Sprintf("--enablerepo=%s", repoName)}, packages...)...)
		}
	} else {
		if p.hasCommand("dnf") {
			installCmd = exec.Command("dnf", append([]string{"install", "-y"}, packages...)...)
		} else {
			installCmd = exec.Command("yum", append([]string{"install", "-y"}, packages...)...)
		}
	}

	installCmd.Stderr = &stderr
	installCmd.Stdout = nil

	if err := installCmd.Run(); err != nil {
		if repoExists {
			stderr.Reset()
			if p.hasCommand("dnf") {
				installCmd = exec.Command("dnf", append([]string{"install", "-y"}, packages...)...)
			} else {
				installCmd = exec.Command("yum", append([]string{"install", "-y"}, packages...)...)
			}
			installCmd.Stderr = &stderr
			installCmd.Stdout = nil
			
			if err := installCmd.Run(); err != nil {
				errorMsg := strings.TrimSpace(stderr.String())
				if errorMsg == "" {
					errorMsg = err.Error()
				}
				return fmt.Errorf("failed to install PHP packages: %s", errorMsg)
			}
		} else {
			errorMsg := strings.TrimSpace(stderr.String())
			if errorMsg == "" {
				errorMsg = err.Error()
			}
			return fmt.Errorf("failed to install PHP packages: %s", errorMsg)
		}
	}

	// Enable and start PHP-FPM service
	serviceName := p.GetServiceName(version)
	enableService := exec.Command("systemctl", "enable", serviceName)
	enableService.Run()

	startService := exec.Command("systemctl", "start", serviceName)
	if err := startService.Run(); err != nil {
		return fmt.Errorf("failed to start PHP-FPM service: %w", err)
	}

	// Save to database
	if err := p.db.CreatePHPVersion(version, "remi", "rhel"); err != nil {
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (p *RemiProvider) installPHPDebian(version, versionNum string) error {
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

	// Add ondrej/php PPA
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
	serviceName := p.GetServiceName(version)
	enableService := exec.Command("systemctl", "enable", serviceName)
	enableService.Run()

	startService := exec.Command("systemctl", "start", serviceName)
	if err := startService.Run(); err != nil {
		return fmt.Errorf("failed to start PHP-FPM service: %w", err)
	}

	// Save to database
	if err := p.db.CreatePHPVersion(version, "ondrej", "debian"); err != nil {
		fmt.Printf("Warning: failed to save PHP version to database: %v\n", err)
	}

	return nil
}

func (p *RemiProvider) ensureRemiRepo() error {
	if p.osFamily != system.OSRHEL {
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
		releaseCmd = exec.Command("cat", "/etc/redhat-release")
		output, err = releaseCmd.Output()
	}

	releaseStr := strings.TrimSpace(string(output))
	var epelURL, remiURL string
	var useDnf bool

	// Determine EPEL and Remi URLs based on RHEL version
	if strings.Contains(releaseStr, "10") {
		epelURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm"
		remiURL = "https://rpms.remirepo.net/enterprise/remi-release-10.rpm"
		useDnf = true
	} else if strings.Contains(releaseStr, "9") {
		epelURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm"
		remiURL = "https://rpms.remirepo.net/enterprise/remi-release-9.rpm"
		useDnf = true
	} else if strings.Contains(releaseStr, "8") {
		epelURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm"
		remiURL = "https://rpms.remirepo.net/enterprise/remi-release-8.rpm"
		useDnf = true
	} else if strings.Contains(releaseStr, "7") {
		epelURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm"
		remiURL = "https://rpms.remirepo.net/enterprise/remi-release-7.rpm"
		useDnf = false
	} else {
		epelURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm"
		remiURL = "https://rpms.remirepo.net/enterprise/remi-release-9.rpm"
		useDnf = true
	}

	// Install EPEL first (required for Remi)
	checkEpelCmd := exec.Command("rpm", "-q", "epel-release")
	if checkEpelCmd.Run() != nil {
		var epelCmd *exec.Cmd
		if useDnf {
			epelCmd = exec.Command("dnf", "install", "-y", epelURL)
		} else {
			epelCmd = exec.Command("yum", "install", "-y", epelURL)
		}
		epelCmd.Stdout = nil
		epelCmd.Stderr = nil
		if err := epelCmd.Run(); err != nil {
			return fmt.Errorf("failed to install EPEL repository: %w", err)
		}
	}

	// Install Remi repository
	var remiCmd *exec.Cmd
	if useDnf {
		remiCmd = exec.Command("dnf", "install", "-y", remiURL)
	} else {
		remiCmd = exec.Command("yum", "install", "-y", remiURL)
	}
	remiCmd.Stdout = nil
	remiCmd.Stderr = nil
	if err := remiCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Remi repository: %w", err)
	}

	return nil
}

func (p *RemiProvider) ListInstalledPHP() ([]string, error) {
	// Try to get from database first
	dbVersions, err := p.db.ListPHPVersions()
	if err == nil && dbVersions != nil && len(dbVersions) > 0 {
		versions := make([]string, 0, len(dbVersions))
		for _, v := range dbVersions {
			if v.Status == "active" && v.PackageManager == "remi" {
				versions = append(versions, v.Version)
			}
		}
		if len(versions) > 0 {
			return versions, nil
		}
	}

	// Fallback to system detection
	versions := make([]string, 0)

	if p.osFamily == system.OSRHEL {
		// Check for installed PHP packages
		if p.hasCommand("dnf") {
			cmd := exec.Command("dnf", "list", "installed", "php*-php-fpm")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "php") && strings.Contains(line, "fpm") {
						parts := strings.Fields(line)
						if len(parts) > 0 {
							pkgName := parts[0]
							if strings.HasPrefix(pkgName, "php") {
								versionNum := strings.TrimPrefix(pkgName, "php")
								if idx := strings.Index(versionNum, "-php-fpm"); idx > 0 {
									versionNum = versionNum[:idx]
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
		} else {
			cmd := exec.Command("yum", "list", "installed", "php*-php-fpm")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "php") && strings.Contains(line, "fpm") {
						parts := strings.Fields(line)
						if len(parts) > 0 {
							pkgName := parts[0]
							if strings.HasPrefix(pkgName, "php") {
								versionNum := strings.TrimPrefix(pkgName, "php")
								if idx := strings.Index(versionNum, "-php-fpm"); idx > 0 {
									versionNum = versionNum[:idx]
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

	if versions == nil {
		versions = []string{}
	}
	return versions, nil
}

func (p *RemiProvider) ListAvailablePHP() ([]string, error) {
	if p.osFamily == system.OSRHEL {
		return p.listAvailablePHPRHEL()
	} else {
		return p.listAvailablePHPDebian()
	}
}

func (p *RemiProvider) listAvailablePHPRHEL() ([]string, error) {
	versions := []string{
		"8.3",
		"8.2",
		"8.1",
		"8.0",
		"7.4",
	}
	return versions, nil
}

func (p *RemiProvider) listAvailablePHPDebian() ([]string, error) {
	versions := []string{
		"8.3",
		"8.2",
		"8.1",
		"8.0",
		"7.4",
	}
	return versions, nil
}

func (p *RemiProvider) hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// installPHPRHEL and installPHPDebian methods will be moved here from package.go
// For now, keeping the structure - these will be implemented in the next step
