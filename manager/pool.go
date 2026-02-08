package manager

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"lightweight-php/db"
	"lightweight-php/system"
)

type Pool struct {
	User       string
	PHPVersion string
	Status     string
	ConfigPath string
	SocketPath string
}

type PoolManager struct {
	osFamily system.OSFamily
	fpmDir   string
	db       *db.Database
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

	return &PoolManager{
		osFamily: osFamily,
		fpmDir:   fpmDir,
		db:       database,
	}, nil
}

func (pm *PoolManager) CreatePool(username, phpVersion string) error {
	// Verify user exists
	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user %s does not exist: %w", username, err)
	}

	// Determine PHP-FPM config directory based on version
	var poolDir string
	if pm.osFamily == system.OSRHEL {
		poolDir = fmt.Sprintf("/etc/php-fpm.d")
	} else {
		poolDir = fmt.Sprintf("/etc/php/%s/fpm/pool.d", phpVersion)
	}

	// Create pool directory if it doesn't exist
	if err := os.MkdirAll(poolDir, 0755); err != nil {
		return fmt.Errorf("failed to create pool directory: %w", err)
	}

	poolFile := filepath.Join(poolDir, fmt.Sprintf("%s.conf", username))

	// Check if pool already exists
	if _, err := os.Stat(poolFile); err == nil {
		return fmt.Errorf("pool for user %s already exists", username)
	}

	// Get user info
	u, _ := user.Lookup(username)
	uid := u.Uid
	gid := u.Gid

	// Determine socket path
	var socketPath string
	if pm.osFamily == system.OSRHEL {
		socketPath = fmt.Sprintf("/var/run/php-fpm/%s.sock", username)
	} else {
		socketPath = fmt.Sprintf("/var/run/php/php%s-%s.sock", phpVersion, username)
	}

	// Create pool configuration
	config := pm.generatePoolConfig(username, uid, gid, socketPath, phpVersion)

	// Write pool configuration
	if err := os.WriteFile(poolFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	// Create socket directory
	socketDir := filepath.Dir(socketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Save to database
	if err := pm.db.CreatePool(username, phpVersion, socketPath, poolFile); err != nil {
		// Rollback: remove config file if database save fails
		os.Remove(poolFile)
		return fmt.Errorf("failed to save pool to database: %w", err)
	}

	// Reload PHP-FPM
	if err := pm.reloadFPM(phpVersion); err != nil {
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

	// Remove from database
	if err := pm.db.DeletePool(username); err != nil {
		return fmt.Errorf("failed to delete pool from database: %w", err)
	}

	// Reload PHP-FPM (try all versions)
	if pm.osFamily == system.OSDebian {
		entries, _ := os.ReadDir("/etc/php")
		for _, entry := range entries {
			if entry.IsDir() {
				pm.reloadFPM(entry.Name())
			}
		}
	} else {
		pm.reloadFPM("")
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

func (pm *PoolManager) reloadFPM(phpVersion string) error {
	var cmd *exec.Cmd
	if pm.osFamily == system.OSRHEL {
		cmd = exec.Command("systemctl", "reload", "php-fpm")
	} else {
		serviceName := fmt.Sprintf("php%s-fpm", phpVersion)
		cmd = exec.Command("systemctl", "reload", serviceName)
	}

	if err := cmd.Run(); err != nil {
		// Try alternative method
		if pm.osFamily == system.OSRHEL {
			cmd = exec.Command("systemctl", "reload-or-restart", "php-fpm")
		} else {
			serviceName := fmt.Sprintf("php%s-fpm", phpVersion)
			cmd = exec.Command("systemctl", "reload-or-restart", serviceName)
		}
		return cmd.Run()
	}

	return nil
}
