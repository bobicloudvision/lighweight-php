package manager

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"lightweight-php/db"
	"lightweight-php/provider"
	"lightweight-php/system"
)

type Pool struct {
	User       string
	PHPVersion string
	Provider   string
	Status     string
	ConfigPath string
	SocketPath string
}

type PoolManager struct {
	osFamily       system.OSFamily
	fpmDir         string
	db             *db.Database
	providerFactory *provider.ProviderFactory
}

func NewPoolManager() (*PoolManager, error) {
	database, err := db.NewDatabase("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	detector := system.NewOSDetector()
	osFamily, _ := detector.Detect()

	var fpmDir string
	if osFamily == system.OSRHEL {
		fpmDir = "/etc/php-fpm.d"
	} else {
		fpmDir = "/etc/php/*/fpm/pool.d"
	}

	providerFactory, err := provider.NewProviderFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider factory: %w", err)
	}

	return &PoolManager{
		osFamily:        osFamily,
		fpmDir:          fpmDir,
		db:              database,
		providerFactory: providerFactory,
	}, nil
}

func (pm *PoolManager) CreatePool(username, phpVersion, providerType string) error {
	// Verify user exists
	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user %s does not exist: %w", username, err)
	}

	// Validate provider type
	var providerTypeEnum provider.ProviderType
	switch providerType {
	case "remi":
		providerTypeEnum = provider.ProviderRemi
	case "lsphp":
		providerTypeEnum = provider.ProviderLiteSpeed
	case "alt-php":
		providerTypeEnum = provider.ProviderAltPHP
	case "docker":
		providerTypeEnum = provider.ProviderDocker
	default:
		return fmt.Errorf("invalid provider type: %s. Supported: remi, lsphp, alt-php, docker", providerType)
	}

	// Get provider instance
	phpProvider, err := pm.providerFactory.CreateProvider(providerTypeEnum)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Get paths from provider
	socketPath := phpProvider.GetSocketPath(username, phpVersion)
	configPath := phpProvider.GetConfigPath(username, phpVersion)

	// Get pool directory from config path
	poolDir := filepath.Dir(configPath)

	// Create pool directory if it doesn't exist
	if err := os.MkdirAll(poolDir, 0755); err != nil {
		return fmt.Errorf("failed to create pool directory: %w", err)
	}

	// Check if pool already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("pool for user %s with PHP %s and provider %s already exists", username, phpVersion, providerType)
	}

	// Get user info
	u, _ := user.Lookup(username)
	uid := u.Uid
	gid := u.Gid

	// Create pool configuration
	config := pm.generatePoolConfig(username, uid, gid, socketPath, phpVersion)

	// Write pool configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	// Create socket directory
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Save to database
	if err := pm.db.CreatePool(username, phpVersion, providerType, socketPath, configPath); err != nil {
		// Rollback: remove config file if database save fails
		os.Remove(configPath)
		return fmt.Errorf("failed to save pool to database: %w", err)
	}

	// Reload PHP-FPM using provider's service name
	serviceName := phpProvider.GetServiceName(phpVersion)
	if err := pm.reloadFPMService(serviceName); err != nil {
		return fmt.Errorf("failed to reload PHP-FPM: %w", err)
	}

	return nil
}

func (pm *PoolManager) DeletePool(username string) error {
	// Get pool from database to find config file
	dbPool, err := pm.db.GetPool(username)
	if err != nil {
		return fmt.Errorf("failed to get pool from database: %w", err)
	}
	if dbPool == nil {
		return fmt.Errorf("pool for user %s not found", username)
	}

	// Remove config file
	if _, err := os.Stat(dbPool.ConfigPath); err == nil {
		if err := os.Remove(dbPool.ConfigPath); err != nil {
			return fmt.Errorf("failed to remove pool file: %w", err)
		}
	}

	// Get provider to reload service
	var providerTypeEnum provider.ProviderType
	switch dbPool.Provider {
	case "remi":
		providerTypeEnum = provider.ProviderRemi
	case "lsphp":
		providerTypeEnum = provider.ProviderLiteSpeed
	case "alt-php":
		providerTypeEnum = provider.ProviderAltPHP
	case "docker":
		providerTypeEnum = provider.ProviderDocker
	default:
		// If provider is unknown, skip reload
		providerTypeEnum = provider.ProviderRemi
	}

	phpProvider, err := pm.providerFactory.CreateProvider(providerTypeEnum)
	if err == nil {
		serviceName := phpProvider.GetServiceName(dbPool.PHPVersion)
		pm.reloadFPMService(serviceName)
	}

	// Remove from database
	if err := pm.db.DeletePool(username); err != nil {
		return fmt.Errorf("failed to delete pool from database: %w", err)
	}

	return nil
}

func (pm *PoolManager) ListPools() ([]Pool, error) {
	dbPools, err := pm.db.ListPools()
	if err != nil {
		return nil, fmt.Errorf("failed to list pools from database: %w", err)
	}

	pools := make([]Pool, 0, len(dbPools))
	for _, dbPool := range dbPools {
		pools = append(pools, Pool{
			User:       dbPool.Username,
			PHPVersion: dbPool.PHPVersion,
			Provider:   dbPool.Provider,
			Status:     dbPool.Status,
			ConfigPath: dbPool.ConfigPath,
			SocketPath: dbPool.SocketPath,
		})
	}

	return pools, nil
}

func (pm *PoolManager) generatePoolConfig(username, uid, gid, socketPath, phpVersion string) string {
	config := fmt.Sprintf(`[%s]
user = %s
group = %s
listen = %s
listen.owner = %s
listen.group = %s
listen.mode = 0660

pm = dynamic
pm.max_children = 50
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 35
pm.max_requests = 500

php_admin_value[sendmail_path] = /usr/sbin/sendmail -t -i -f www@my.domain.com
php_flag[display_errors] = off
php_admin_value[error_log] = /var/log/fpm-php.%s.log
php_admin_flag[log_errors] = on
php_admin_value[memory_limit] = 128M
`, username, username, username, socketPath, username, username, username)

	return config
}

func (pm *PoolManager) reloadFPMService(serviceName string) error {
	cmd := exec.Command("systemctl", "reload", serviceName)
	if err := cmd.Run(); err != nil {
		// Try alternative method
		cmd = exec.Command("systemctl", "reload-or-restart", serviceName)
		return cmd.Run()
	}
	return nil
}
