package provider

import (
	"fmt"
	"path/filepath"
	"strings"

	"lightweight-php/db"
	"lightweight-php/system"
)

// AltPHPProvider implements PHPProvider for Alternative PHP (alt-php)
type AltPHPProvider struct {
	db       *db.Database
	osFamily system.OSFamily
}

func NewAltPHPProvider(database *db.Database, osFamily system.OSFamily) (*AltPHPProvider, error) {
	return &AltPHPProvider{
		db:       database,
		osFamily: osFamily,
	}, nil
}

func (p *AltPHPProvider) GetProviderType() string {
	return string(ProviderAltPHP)
}

func (p *AltPHPProvider) GetServiceName(version string) string {
	versionNum := strings.ReplaceAll(version, ".", "")
	return fmt.Sprintf("alt-php%s-php-fpm", versionNum)
}

func (p *AltPHPProvider) GetSocketPath(username, version string) string {
	versionNum := strings.ReplaceAll(version, ".", "")
	return fmt.Sprintf("/var/run/alt-php%s/%s.sock", versionNum, username)
}

func (p *AltPHPProvider) GetConfigPath(username, version string) string {
	versionNum := strings.ReplaceAll(version, ".", "")
	return filepath.Join("/etc/opt/alt/php%s", versionNum, "php-fpm.d", fmt.Sprintf("%s.conf", username))
}

func (p *AltPHPProvider) InstallPHP(version string) error {
	// TODO: Implement Alt-PHP installation
	return fmt.Errorf("Alt-PHP provider not yet implemented")
}

func (p *AltPHPProvider) ListInstalledPHP() ([]string, error) {
	// TODO: Implement listing installed Alt-PHP versions
	return []string{}, nil
}

func (p *AltPHPProvider) ListAvailablePHP() ([]string, error) {
	// TODO: Return available Alt-PHP versions
	return []string{"8.3", "8.2", "8.1", "8.0", "7.4"}, nil
}
