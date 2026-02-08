package provider

import (
	"fmt"
	"path/filepath"

	"lightweight-php/db"
	"lightweight-php/system"
)

// DockerProvider implements PHPProvider for Docker-hosted PHP
type DockerProvider struct {
	db       *db.Database
	osFamily system.OSFamily
}

func NewDockerProvider(database *db.Database, osFamily system.OSFamily) (*DockerProvider, error) {
	return &DockerProvider{
		db:       database,
		osFamily: osFamily,
	}, nil
}

func (p *DockerProvider) GetProviderType() string {
	return string(ProviderDocker)
}

func (p *DockerProvider) GetServiceName(version string) string {
	// Docker containers don't use systemd services directly
	return fmt.Sprintf("php-%s-fpm", version)
}

func (p *DockerProvider) GetSocketPath(username, version string) string {
	// Docker socket path (could be TCP or Unix socket)
	return fmt.Sprintf("/var/run/docker/php-%s-%s.sock", version, username)
}

func (p *DockerProvider) GetConfigPath(username, version string) string {
	// Docker configuration path
	return filepath.Join("/etc/docker/php", version, fmt.Sprintf("%s.conf", username))
}

func (p *DockerProvider) InstallPHP(version string) error {
	// TODO: Implement Docker PHP installation
	return fmt.Errorf("Docker PHP provider not yet implemented")
}

func (p *DockerProvider) ListInstalledPHP() ([]string, error) {
	// TODO: Implement listing installed Docker PHP versions
	// This would query Docker containers
	return []string{}, nil
}

func (p *DockerProvider) ListAvailablePHP() ([]string, error) {
	// TODO: Return available Docker PHP versions
	return []string{"8.3", "8.2", "8.1", "8.0", "7.4"}, nil
}
