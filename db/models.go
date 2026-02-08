package db

import (
	"database/sql"
	"time"
)

type PHPVersion struct {
	ID             int64
	Version        string
	InstalledAt    time.Time
	Status         string
	PackageManager string
	OSFamily       string
}

type Pool struct {
	ID         int64
	Username   string
	PHPVersion string
	SocketPath string
	ConfigPath string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (db *Database) CreatePHPVersion(version, packageManager, osFamily string) error {
	_, err := db.Exec(
		"INSERT INTO php_versions (version, package_manager, os_family, status) VALUES (?, ?, ?, 'active')",
		version, packageManager, osFamily,
	)
	return err
}

func (db *Database) GetPHPVersion(version string) (*PHPVersion, error) {
	var pv PHPVersion
	var installedAt sql.NullTime
	err := db.QueryRow(
		"SELECT id, version, installed_at, status, package_manager, os_family FROM php_versions WHERE version = ?",
		version,
	).Scan(&pv.ID, &pv.Version, &installedAt, &pv.Status, &pv.PackageManager, &pv.OSFamily)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if installedAt.Valid {
		pv.InstalledAt = installedAt.Time
	}

	return &pv, nil
}

func (db *Database) ListPHPVersions() ([]PHPVersion, error) {
	rows, err := db.Query("SELECT id, version, installed_at, status, package_manager, os_family FROM php_versions ORDER BY version DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []PHPVersion
	for rows.Next() {
		var pv PHPVersion
		var installedAt sql.NullTime
		if err := rows.Scan(&pv.ID, &pv.Version, &installedAt, &pv.Status, &pv.PackageManager, &pv.OSFamily); err != nil {
			return nil, err
		}
		if installedAt.Valid {
			pv.InstalledAt = installedAt.Time
		}
		versions = append(versions, pv)
	}

	return versions, rows.Err()
}

func (db *Database) CreatePool(username, phpVersion, socketPath, configPath string) error {
	_, err := db.Exec(
		`INSERT INTO pools (username, php_version, socket_path, config_path, status) 
		 VALUES (?, ?, ?, ?, 'active')
		 ON CONFLICT(username, php_version) DO UPDATE SET
		 socket_path = excluded.socket_path,
		 config_path = excluded.config_path,
		 updated_at = CURRENT_TIMESTAMP`,
		username, phpVersion, socketPath, configPath,
	)
	return err
}

func (db *Database) GetPool(username string) (*Pool, error) {
	var p Pool
	var createdAt, updatedAt sql.NullTime
	err := db.QueryRow(
		`SELECT id, username, php_version, socket_path, config_path, status, created_at, updated_at 
		 FROM pools WHERE username = ? ORDER BY created_at DESC LIMIT 1`,
		username,
	).Scan(&p.ID, &p.Username, &p.PHPVersion, &p.SocketPath, &p.ConfigPath, &p.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}

	return &p, nil
}

func (db *Database) GetPoolByUsernameAndVersion(username, phpVersion string) (*Pool, error) {
	var p Pool
	var createdAt, updatedAt sql.NullTime
	err := db.QueryRow(
		`SELECT id, username, php_version, socket_path, config_path, status, created_at, updated_at 
		 FROM pools WHERE username = ? AND php_version = ?`,
		username, phpVersion,
	).Scan(&p.ID, &p.Username, &p.PHPVersion, &p.SocketPath, &p.ConfigPath, &p.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		p.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		p.UpdatedAt = updatedAt.Time
	}

	return &p, nil
}

func (db *Database) ListPools() ([]Pool, error) {
	rows, err := db.Query(
		`SELECT id, username, php_version, socket_path, config_path, status, created_at, updated_at 
		 FROM pools ORDER BY username, created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []Pool
	for rows.Next() {
		var p Pool
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Username, &p.PHPVersion, &p.SocketPath, &p.ConfigPath, &p.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			p.UpdatedAt = updatedAt.Time
		}
		pools = append(pools, p)
	}

	return pools, rows.Err()
}

func (db *Database) DeletePool(username string) error {
	result, err := db.Exec("DELETE FROM pools WHERE username = ?", username)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (db *Database) UpdatePoolStatus(username, status string) error {
	_, err := db.Exec(
		"UPDATE pools SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE username = ?",
		status, username,
	)
	return err
}
