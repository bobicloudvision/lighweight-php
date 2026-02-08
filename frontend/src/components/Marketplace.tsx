import { useState, useEffect } from 'react'
import { apiService, type Provider } from '../services/api'

interface PhpVersion {
  version: string
  name: string
  description: string
  status: 'available' | 'installing' | 'installed'
  releaseDate?: string
  eolDate?: string
}

// Metadata for PHP versions (descriptions, release dates, etc.)
const VERSION_METADATA: Record<string, Omit<PhpVersion, 'version' | 'status'>> = {
  '8.4': {
    name: 'PHP 8.4',
    description: 'Latest stable release with modern features and performance improvements',
    releaseDate: '2024-11-21',
  },
  '8.3': {
    name: 'PHP 8.3',
    description: 'Stable release with improved JIT compiler and new features',
    releaseDate: '2023-11-23',
  },
  '8.2': {
    name: 'PHP 8.2',
    description: 'Stable release with read-only classes and performance enhancements',
    releaseDate: '2022-12-08',
  },
  '8.1': {
    name: 'PHP 8.1',
    description: 'Stable release with enums, fibers, and intersection types',
    releaseDate: '2021-11-25',
    eolDate: '2024-11-25',
  },
  '8.0': {
    name: 'PHP 8.0',
    description: 'Major release with JIT compiler and union types',
    releaseDate: '2020-11-26',
    eolDate: '2023-11-26',
  },
  '7.4': {
    name: 'PHP 7.4',
    description: 'Last PHP 7.x release with typed properties and arrow functions',
    releaseDate: '2019-11-28',
    eolDate: '2022-11-28',
  },
  '7.3': {
    name: 'PHP 7.3',
    description: 'PHP 7.3 release with flexible heredoc/nowdoc syntax',
    releaseDate: '2018-12-06',
    eolDate: '2021-12-06',
  },
  '7.2': {
    name: 'PHP 7.2',
    description: 'PHP 7.2 release with improved performance',
    releaseDate: '2017-11-30',
    eolDate: '2020-11-30',
  },
}

// Helper to get metadata for a version, or create default if not found
const getVersionMetadata = (version: string): Omit<PhpVersion, 'version' | 'status'> => {
  const metadata = VERSION_METADATA[version]
  if (metadata) {
    return metadata
  }
  // Default metadata for versions not in our list
  return {
    name: `PHP ${version}`,
    description: `PHP ${version} from repository`,
  }
}

export default function Marketplace() {
  const [availableVersions, setAvailableVersions] = useState<string[]>([])
  const [installedVersions, setInstalledVersions] = useState<string[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [selectedProvider, setSelectedProvider] = useState<string>('remi')
  const [loading, setLoading] = useState(true)
  const [installing, setInstalling] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [filter, setFilter] = useState<'all' | 'installed' | 'available'>('all')

  useEffect(() => {
    loadData()
  }, [])

  const handleProviderChange = (providerType: string) => {
    setSelectedProvider(providerType)
    if (providerType === 'remi') {
      loadData()
    } else {
      loadProviderVersions(providerType)
    }
  }

  const loadData = async () => {
    setLoading(true)
    setError(null)
    
    // Load providers, available versions, and installed versions in parallel
    const [providersResult, availableResult, installedResult] = await Promise.all([
      apiService.getProviders(),
      apiService.getAvailablePhpVersions(),
      apiService.getPhpVersions(),
    ])

    if (providersResult.data && providersResult.data.providers) {
      const providersList = Array.isArray(providersResult.data.providers) ? providersResult.data.providers : []
      setProviders(providersList)
      // Set default provider to 'remi' if available and not already set
      if (providersList.length > 0 && !selectedProvider) {
        const remiProvider = providersList.find(p => p.type === 'remi')
        if (remiProvider) {
          setSelectedProvider('remi')
        } else {
          setSelectedProvider(providersList[0].type)
        }
      }
    }

    if (availableResult.data && availableResult.data.versions) {
      setAvailableVersions(Array.isArray(availableResult.data.versions) ? availableResult.data.versions : [])
    } else {
      setError(availableResult.error || 'Failed to load available versions')
      setAvailableVersions([])
    }

    if (installedResult.data && installedResult.data.versions) {
      const versionsList = Array.isArray(installedResult.data.versions) ? installedResult.data.versions : []
      // Extract version strings from PhpVersion objects (backward compatible)
      const versionStrings = versionsList.map(v => typeof v === 'string' ? v : v.version)
      setInstalledVersions(versionStrings)
    } else {
      setInstalledVersions([])
    }

    setLoading(false)
  }

  const loadProviderVersions = async (provider?: string) => {
    const providerToUse = provider || selectedProvider
    setLoading(true)
    setError(null)

    const [availableResult, installedResult] = await Promise.all([
      apiService.getProviderAvailableVersions(providerToUse),
      apiService.getProviderVersions(providerToUse),
    ])

    if (availableResult.data && availableResult.data.versions) {
      setAvailableVersions(Array.isArray(availableResult.data.versions) ? availableResult.data.versions : [])
    } else {
      setError(availableResult.error || `Failed to load available versions for ${selectedProvider}`)
      setAvailableVersions([])
    }

    if (installedResult.data && installedResult.data.versions) {
      // Provider-specific versions are still string arrays
      const versionsList = Array.isArray(installedResult.data.versions) ? installedResult.data.versions : []
      setInstalledVersions(versionsList)
    } else {
      setInstalledVersions([])
    }

    setLoading(false)
  }

  const handleInstall = async (version: string) => {
    setInstalling(version)
    setError(null)
    setSuccess(null)

    const result = selectedProvider === 'remi'
      ? await apiService.installPhpVersion(version, selectedProvider)
      : await apiService.installPhpVersionWithProvider(selectedProvider, version)
    
    if (result.data) {
      setSuccess(`PHP ${version} installed successfully using ${selectedProvider}!`)
      if (selectedProvider === 'remi') {
        await loadData()
      } else {
        await loadProviderVersions()
      }
    } else {
      setError(result.error || `Failed to install PHP ${version}`)
    }

    setInstalling(null)
  }

  const getVersionStatus = (version: string): 'available' | 'installing' | 'installed' => {
    if (installing === version) return 'installing'
    if (installedVersions.includes(version)) return 'installed'
    return 'available'
  }

  const getStatusBadge = (status: 'available' | 'installing' | 'installed') => {
    switch (status) {
      case 'installed':
        return (
          <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-xs font-semibold flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            Installed
          </span>
        )
      case 'installing':
        return (
          <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-xs font-semibold flex items-center gap-1">
            <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Installing...
          </span>
        )
      default:
        return (
          <span className="px-3 py-1 bg-gray-100 text-gray-800 rounded-full text-xs font-semibold">
            Available
          </span>
        )
    }
  }

  // Merge available versions with metadata
  const versions: PhpVersion[] = availableVersions
    .map(version => {
      const metadata = getVersionMetadata(version)
      return {
        version,
        ...metadata,
        status: getVersionStatus(version),
      }
    })
    .sort((a, b) => {
      // Sort by version number (descending - newest first)
      const aParts = a.version.split('.').map(Number)
      const bParts = b.version.split('.').map(Number)
      for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
        const aVal = aParts[i] || 0
        const bVal = bParts[i] || 0
        if (bVal !== aVal) return bVal - aVal
      }
      return 0
    })

  const filteredVersions = versions.filter(v => {
    if (filter === 'installed') return v.status === 'installed'
    if (filter === 'available') return v.status === 'available'
    return true
  })

  const isEol = (version: PhpVersion) => {
    if (!version.eolDate) return false
    return new Date(version.eolDate) < new Date()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold text-gray-900 mb-2">PHP Marketplace</h2>
          <p className="text-gray-600">Browse and install available PHP versions from repository</p>
        </div>
        <button
          onClick={() => selectedProvider === 'remi' ? loadData() : loadProviderVersions()}
          disabled={loading}
          className="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:bg-gray-50 disabled:text-gray-400 transition-colors flex items-center gap-2"
        >
          <svg className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      {/* Provider Selection */}
      {providers.length > 0 && (
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <label htmlFor="provider-select" className="block text-sm font-medium text-gray-700 mb-3">
            PHP Provider
          </label>
          <div className="flex gap-4 items-center">
            <select
              id="provider-select"
              value={selectedProvider}
              onChange={(e) => handleProviderChange(e.target.value)}
              className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
            >
              {providers.map((provider) => (
                <option key={provider.type} value={provider.type}>
                  {provider.name} {provider.status === 'stub' && '(Stub)'}
                </option>
              ))}
            </select>
            {providers.find(p => p.type === selectedProvider) && (
              <p className="text-sm text-gray-600">
                {providers.find(p => p.type === selectedProvider)?.description}
              </p>
            )}
          </div>
        </div>
      )}

      {/* Filter Tabs */}
      <div className="bg-white rounded-lg shadow-md p-4 border border-gray-200">
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium text-gray-700">Filter:</span>
          <div className="flex gap-2">
            {(['all', 'available', 'installed'] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
                  filter === f
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {f === 'all' ? 'All Versions' : f === 'available' ? 'Available' : 'Installed'}
                {f === 'installed' && installedVersions.length > 0 && (
                  <span className="ml-2 px-2 py-0.5 bg-white bg-opacity-20 rounded-full">
                    {installedVersions.length}
                  </span>
                )}
                {f === 'available' && availableVersions.length > 0 && (
                  <span className="ml-2 px-2 py-0.5 bg-white bg-opacity-20 rounded-full">
                    {availableVersions.length - installedVersions.length}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Messages */}
      {error && (
        <div className="bg-red-50 border-l-4 border-red-400 p-4 rounded">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {success && (
        <div className="bg-green-50 border-l-4 border-green-400 p-4 rounded">
          <p className="text-sm text-green-700">{success}</p>
        </div>
      )}

      {/* Versions Grid */}
      {loading ? (
        <div className="text-center py-12">
          <svg className="animate-spin h-8 w-8 text-gray-400 mx-auto" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <p className="mt-2 text-gray-500">Loading versions from repository...</p>
        </div>
      ) : filteredVersions.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
          <p className="text-lg">No versions found</p>
          <p className="text-sm mt-2">Try selecting a different filter.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredVersions.map((version) => (
            <div
              key={version.version}
              className={`bg-white rounded-lg shadow-md border-2 overflow-hidden transition-all hover:shadow-lg ${
                version.status === 'installed'
                  ? 'border-green-200'
                  : version.status === 'installing'
                  ? 'border-blue-200'
                  : 'border-gray-200'
              }`}
            >
              {/* Header */}
              <div className={`p-6 ${
                version.status === 'installed'
                  ? 'bg-gradient-to-br from-green-50 to-green-100'
                  : version.status === 'installing'
                  ? 'bg-gradient-to-br from-blue-50 to-blue-100'
                  : 'bg-gradient-to-br from-gray-50 to-gray-100'
              }`}>
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-2xl font-bold text-gray-900">{version.name}</h3>
                    <p className="text-sm text-gray-600 mt-1">Version {version.version}</p>
                  </div>
                  {getStatusBadge(version.status)}
                </div>
                
                {isEol(version) && (
                  <div className="mt-3 bg-yellow-100 border border-yellow-300 rounded-lg p-2">
                    <p className="text-xs text-yellow-800 font-medium flex items-center gap-1">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                      End of Life
                    </p>
                  </div>
                )}
              </div>

              {/* Content */}
              <div className="p-6">
                <p className="text-gray-600 text-sm mb-4">{version.description}</p>
                
                <div className="space-y-2 mb-4 text-xs text-gray-500">
                  {version.releaseDate && (
                    <div className="flex items-center gap-2">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                      <span>Released: {new Date(version.releaseDate).toLocaleDateString()}</span>
                    </div>
                  )}
                  {version.eolDate && (
                    <div className="flex items-center gap-2">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <span>EOL: {new Date(version.eolDate).toLocaleDateString()}</span>
                    </div>
                  )}
                </div>

                {/* Action Button */}
                <button
                  onClick={() => handleInstall(version.version)}
                  disabled={version.status === 'installed' || version.status === 'installing'}
                  className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors flex items-center justify-center gap-2 ${
                    version.status === 'installed'
                      ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                      : version.status === 'installing'
                      ? 'bg-blue-100 text-blue-600 cursor-not-allowed'
                      : 'bg-blue-600 text-white hover:bg-blue-700'
                  }`}
                >
                  {version.status === 'installed' ? (
                    <>
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Installed
                    </>
                  ) : version.status === 'installing' ? (
                    <>
                      <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Installing...
                    </>
                  ) : (
                    <>
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                      </svg>
                      Install
                    </>
                  )}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Info Box */}
      <div className="bg-blue-50 border-l-4 border-blue-400 p-4 rounded">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-sm text-blue-700">
              <strong>Note:</strong> Available versions are fetched from your system's repository based on the selected provider. 
              The default provider is "remi" (Remi repository for RHEL, ondrej PPA for Debian). 
              Other providers (lsphp, alt-php, docker) may have different available versions. 
              Installation may take a few minutes depending on your system and provider.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
