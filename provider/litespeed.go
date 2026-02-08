package provider

import (
	"fmt"
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
	// LiteSpeed socket path format
	return fmt.Sprintf("/tmp/lsphp%s-%s.sock", strings.ReplaceAll(version, ".", ""), username)
}

func (p *LiteSpeedProvider) GetConfigPath(username, version string) string {
	// LiteSpeed configuration path
	return filepath.Join("/usr/local/lsws/conf", fmt.Sprintf("%s-%s.conf", username, version))
}

func (p *LiteSpeedProvider) InstallPHP(version string) error {
	// TODO: Implement LiteSpeed PHP installation
	return fmt.Errorf("LiteSpeed PHP provider not yet implemented")
}

func (p *LiteSpeedProvider) ListInstalledPHP() ([]string, error) {
	// TODO: Implement listing installed LiteSpeed PHP versions
	return []string{}, nil
}

func (p *LiteSpeedProvider) ListAvailablePHP() ([]string, error) {
	// TODO: Return available LiteSpeed PHP versions
	return []string{"8.3", "8.2", "8.1", "8.0", "7.4"}, nil
}
