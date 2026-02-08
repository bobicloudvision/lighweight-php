# Architecture: PHP Provider System

## Overview

The application uses a provider-based architecture to support multiple PHP installation backends:
- **Remi** (default): Remi repository for RHEL, ondrej PPA for Debian
- **LiteSpeed** (lsphp): LiteSpeed PHP
- **Alt-PHP**: Alternative PHP
- **Docker**: Docker-hosted PHP containers

## Architecture Components

### 1. Provider Interface (`provider/interface.go`)

Defines the `PHPProvider` interface that all providers must implement:

```go
type PHPProvider interface {
    InstallPHP(version string) error
    ListInstalledPHP() ([]string, error)
    ListAvailablePHP() ([]string, error)
    GetProviderType() string
    GetServiceName(version string) string
    GetSocketPath(username, version string) string
    GetConfigPath(username, version string) string
}
```

### 2. Provider Factory (`provider/factory.go`)

Creates and manages PHP providers:

- `NewProviderFactory()`: Initializes the factory
- `CreateProvider(ProviderType)`: Creates a specific provider
- `GetDefaultProvider()`: Returns the default provider (Remi)

### 3. Provider Implementations

#### Remi Provider (`provider/remi.go`)
- **Status**: ✅ Fully implemented
- **Supports**: RHEL (via Remi repo) and Debian (via ondrej PPA)
- **Features**: EPEL installation, repository management, package installation

#### LiteSpeed Provider (`provider/litespeed.go`)
- **Status**: 🚧 Stub implementation
- **Supports**: LiteSpeed Web Server PHP
- **TODO**: Implement installation and management

#### Alt-PHP Provider (`provider/altphp.go`)
- **Status**: 🚧 Stub implementation
- **Supports**: Alternative PHP packages
- **TODO**: Implement installation and management

#### Docker Provider (`provider/docker.go`)
- **Status**: 🚧 Stub implementation
- **Supports**: Docker-hosted PHP containers
- **TODO**: Implement Docker container management

### 4. Package Manager (`manager/package.go`)

Acts as a facade that uses the provider system:

```go
type PackageManager struct {
    providerFactory *provider.ProviderFactory
    defaultProvider provider.PHPProvider
}
```

Methods:
- `InstallPHP(version)`: Uses default provider
- `InstallPHPWithProvider(version, providerType)`: Uses specific provider
- `ListInstalledPHP()`: Lists installed versions
- `ListAvailablePHP()`: Lists available versions

### 5. Pool Manager Integration

The `PoolManager` should be updated to use providers for:
- Getting socket paths: `provider.GetSocketPath(username, version)`
- Getting config paths: `provider.GetConfigPath(username, version)`
- Getting service names: `provider.GetServiceName(version)`

## Usage Examples

### Install PHP with default provider (Remi)
```go
pm, _ := manager.NewPackageManager()
pm.InstallPHP("8.2")
```

### Install PHP with specific provider
```go
pm, _ := manager.NewPackageManager()
pm.InstallPHPWithProvider("8.2", provider.ProviderLiteSpeed)
```

### Get provider-specific paths
```go
provider, _ := pm.GetProviderByType(provider.ProviderLiteSpeed)
socketPath := provider.GetSocketPath("username", "8.2")
configPath := provider.GetConfigPath("username", "8.2")
```

## Database Schema

The `php_versions` table tracks the provider type:
- `package_manager` field stores: "remi", "lsphp", "alt-php", "docker"

## API Extensions Needed

The REST API should support provider selection:

```
POST /api/v1/php/install/{version}?provider=lsphp
POST /api/v1/php/install/{version}?provider=alt-php
POST /api/v1/php/install/{version}?provider=docker
```

## Future Enhancements

1. **LiteSpeed Provider**: Implement lsphp installation and management
2. **Alt-PHP Provider**: Implement alt-php package installation
3. **Docker Provider**: Implement Docker container management
4. **Provider Selection**: Add API endpoints to list available providers
5. **Pool Provider Association**: Store provider type with each pool in database
