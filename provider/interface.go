package provider

// PHPProvider defines the interface for different PHP installation providers
type PHPProvider interface {
	// InstallPHP installs a PHP version
	InstallPHP(version string) error
	
	// ListInstalledPHP returns list of installed PHP versions
	ListInstalledPHP() ([]string, error)
	
	// ListAvailablePHP returns list of available PHP versions
	ListAvailablePHP() ([]string, error)
	
	// GetProviderType returns the provider type name (e.g., "remi", "lsphp", "alt-php", "docker")
	GetProviderType() string
	
	// GetServiceName returns the systemd service name for a PHP version
	GetServiceName(version string) string
	
	// GetSocketPath returns the socket path for a PHP-FPM pool
	GetSocketPath(username, version string) string
	
	// GetConfigPath returns the pool configuration file path
	GetConfigPath(username, version string) string
}

// ProviderType represents different PHP provider types
type ProviderType string

const (
	ProviderRemi    ProviderType = "remi"
	ProviderLiteSpeed ProviderType = "lsphp"
	ProviderAltPHP  ProviderType = "alt-php"
	ProviderDocker  ProviderType = "docker"
)
