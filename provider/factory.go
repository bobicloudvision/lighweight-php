package provider

import (
	"fmt"

	"lightweight-php/db"
	"lightweight-php/system"
)

// ProviderFactory creates PHP providers based on type
type ProviderFactory struct {
	db       *db.Database
	osFamily system.OSFamily
}

func NewProviderFactory() (*ProviderFactory, error) {
	database, err := db.NewDatabase("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	detector := system.NewOSDetector()
	osFamily, _ := detector.Detect()

	return &ProviderFactory{
		db:       database,
		osFamily: osFamily,
	}, nil
}

// CreateProvider creates a PHP provider based on the provider type
func (f *ProviderFactory) CreateProvider(providerType ProviderType) (PHPProvider, error) {
	switch providerType {
	case ProviderRemi:
		return NewRemiProvider(f.db, f.osFamily)
	case ProviderLiteSpeed:
		return NewLiteSpeedProvider(f.db, f.osFamily)
	case ProviderAltPHP:
		return NewAltPHPProvider(f.db, f.osFamily)
	case ProviderDocker:
		return NewDockerProvider(f.db, f.osFamily)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetDefaultProvider returns the default provider (remi for RHEL, ondrej for Debian)
func (f *ProviderFactory) GetDefaultProvider() (PHPProvider, error) {
	return f.CreateProvider(ProviderRemi)
}
