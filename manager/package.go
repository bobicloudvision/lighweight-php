package manager

import (
	"fmt"

	"lightweight-php/provider"
)

type PackageManager struct {
	providerFactory *provider.ProviderFactory
	defaultProvider provider.PHPProvider
}

func NewPackageManager() (*PackageManager, error) {
	factory, err := provider.NewProviderFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider factory: %w", err)
	}

	defaultProvider, err := factory.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get default provider: %w", err)
	}

	return &PackageManager{
		providerFactory: factory,
		defaultProvider: defaultProvider,
	}, nil
}

// InstallPHP installs PHP using the default provider (remi)
func (pm *PackageManager) InstallPHP(version string) error {
	return pm.defaultProvider.InstallPHP(version)
}

// InstallPHPWithProvider installs PHP using a specific provider
func (pm *PackageManager) InstallPHPWithProvider(version string, providerType provider.ProviderType) error {
	phpProvider, err := pm.providerFactory.CreateProvider(providerType)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	return phpProvider.InstallPHP(version)
}

func (pm *PackageManager) ListInstalledPHP() ([]string, error) {
	return pm.defaultProvider.ListInstalledPHP()
}

func (pm *PackageManager) ListAvailablePHP() ([]string, error) {
	return pm.defaultProvider.ListAvailablePHP()
}

// GetProvider returns the default provider
func (pm *PackageManager) GetProvider() provider.PHPProvider {
	return pm.defaultProvider
}

// GetProviderByType returns a provider by type
func (pm *PackageManager) GetProviderByType(providerType provider.ProviderType) (provider.PHPProvider, error) {
	return pm.providerFactory.CreateProvider(providerType)
}
