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
	"lightweight-php/templates"
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

// GetDatabase returns the database instance (for API access)
func (pm *PoolManager) GetDatabase() *db.Database {
	return pm.db
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
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("failed to lookup user: %w", err)
	}
	uid := u.Uid
	gid := u.Gid

	// Create pool configuration using template
	config, err := pm.generatePoolConfig(username, uid, gid, socketPath, phpVersion)
	if err != nil {
		return fmt.Errorf("failed to generate pool config: %w", err)
	}

	// Write pool configuration
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	// Create socket directory
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Ensure PHP version exists in database (required for foreign key constraint)
	phpVersionRecord, err := pm.db.GetPHPVersion(phpVersion)
	if err != nil {
		return fmt.Errorf("failed to check PHP version: %w", err)
	}
	if phpVersionRecord == nil {
		// PHP version not registered, create it
		detector := system.NewOSDetector()
		osFamily, _ := detector.Detect()
		var osFamilyStr string
		if osFamily == system.OSRHEL {
			osFamilyStr = "rhel"
		} else {
			osFamilyStr = "debian"
		}
		if err := pm.db.CreatePHPVersion(phpVersion, providerType, osFamilyStr); err != nil {
			return fmt.Errorf("failed to register PHP version: %w", err)
		}
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

func (pm *PoolManager) UpdatePoolConfig(username string, settings map[string]interface{}) error {
	// Get pool from database
	dbPool, err := pm.db.GetPool(username)
	if err != nil {
		return fmt.Errorf("failed to get pool from database: %w", err)
	}
	if dbPool == nil {
		return fmt.Errorf("pool for user %s not found", username)
	}

	// Get user info for group name
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("failed to lookup user: %w", err)
	}

	// Get group name
	groupName := username
	if u.Gid != "" {
		g, err := user.LookupGroupId(u.Gid)
		if err == nil {
			groupName = g.Name
		}
	}

	// Load template
	templateContent, err := templates.LoadTemplate("pool.conf.tmpl")
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Create template data with defaults
	data := templates.DefaultPoolConfigData(username, groupName, dbPool.SocketPath)

	// Apply custom settings
	if err := applyPoolSettings(data, settings); err != nil {
		return fmt.Errorf("failed to apply settings: %w", err)
	}

	// Render template
	config, err := templates.RenderPoolConfig(templateContent, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Write updated configuration
	if err := os.WriteFile(dbPool.ConfigPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	// Reload PHP-FPM
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
		providerTypeEnum = provider.ProviderRemi
	}

	phpProvider, err := pm.providerFactory.CreateProvider(providerTypeEnum)
	if err == nil {
		serviceName := phpProvider.GetServiceName(dbPool.PHPVersion)
		if err := pm.reloadFPMService(serviceName); err != nil {
			return fmt.Errorf("failed to reload PHP-FPM: %w", err)
		}
	}

	return nil
}

func applyPoolSettings(data *templates.PoolConfigData, settings map[string]interface{}) error {
	for key, value := range settings {
		switch key {
		case "max_children":
			if v, ok := value.(float64); ok {
				data.MaxChildren = int(v)
			}
		case "start_servers":
			if v, ok := value.(float64); ok {
				data.StartServers = int(v)
			}
		case "min_spare_servers":
			if v, ok := value.(float64); ok {
				data.MinSpareServers = int(v)
			}
		case "max_spare_servers":
			if v, ok := value.(float64); ok {
				data.MaxSpareServers = int(v)
			}
		case "max_requests":
			if v, ok := value.(float64); ok {
				data.MaxRequests = int(v)
			}
		case "process_manager":
			if v, ok := value.(string); ok {
				data.ProcessManager = v
			}
		case "memory_limit":
			if v, ok := value.(string); ok {
				data.MemoryLimit = v
			}
		case "max_execution_time":
			if v, ok := value.(string); ok {
				data.MaxExecutionTime = v
			} else if v, ok := value.(float64); ok {
				data.MaxExecutionTime = fmt.Sprintf("%.0f", v)
			}
		case "upload_max_filesize":
			if v, ok := value.(string); ok {
				data.UploadMaxFilesize = v
			}
		case "post_max_size":
			if v, ok := value.(string); ok {
				data.PostMaxSize = v
			}
		case "display_errors":
			if v, ok := value.(string); ok {
				data.DisplayErrors = v
			} else if v, ok := value.(bool); ok {
				if v {
					data.DisplayErrors = "on"
				} else {
					data.DisplayErrors = "off"
				}
			}
		case "log_errors":
			if v, ok := value.(string); ok {
				data.LogErrors = v
			} else if v, ok := value.(bool); ok {
				if v {
					data.LogErrors = "on"
				} else {
					data.LogErrors = "off"
				}
			}
		case "date_timezone":
			if v, ok := value.(string); ok {
				data.DateTimezone = v
			}
		case "sendmail_path":
			if v, ok := value.(string); ok {
				data.SendmailPath = v
			}
		case "process_idle_timeout":
			if v, ok := value.(string); ok {
				data.ProcessIdleTimeout = v
			} else if v, ok := value.(float64); ok {
				data.ProcessIdleTimeout = fmt.Sprintf("%.0f", v)
			}
		case "listen_mode":
			if v, ok := value.(string); ok {
				data.ListenMode = v
			}
		}
	}
	return nil
}

func (pm *PoolManager) generatePoolConfig(username, uid, gid, socketPath, phpVersion string) (string, error) {
	// Get user group name
	u, err := user.LookupId(uid)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user: %w", err)
	}
	
	groupName := username
	if u.Gid != "" {
		g, err := user.LookupGroupId(u.Gid)
		if err == nil {
			groupName = g.Name
		}
	}

	// Load template
	templateContent, err := templates.LoadTemplate("pool.conf.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Create template data with defaults
	data := templates.DefaultPoolConfigData(username, groupName, socketPath)

	// Render template
	config, err := templates.RenderPoolConfig(templateContent, data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return config, nil
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
