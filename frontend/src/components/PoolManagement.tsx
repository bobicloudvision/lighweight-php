import { useState, useEffect } from 'react'
import { apiService } from '../services/api'
import type { Pool, Provider, PoolConfig } from '../services/api'

export default function PoolManagement() {
  const [pools, setPools] = useState<Pool[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [installedVersions, setInstalledVersions] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [selectedPool, setSelectedPool] = useState<Pool | null>(null)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [showConfigEditor, setShowConfigEditor] = useState(false)
  const [poolConfig, setPoolConfig] = useState<PoolConfig>({})
  const [updatingConfig, setUpdatingConfig] = useState(false)
  
  // Create form state
  const [username, setUsername] = useState('')
  const [phpVersion, setPhpVersion] = useState('')
  const [provider, setProvider] = useState('remi')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    loadProviders()
    loadPools()
    loadInstalledVersions()
  }, [])

  useEffect(() => {
    // Reload versions when provider changes (but not on initial mount)
    if (providers.length > 0) {
      loadInstalledVersions()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [provider])

  const loadProviders = async () => {
    const result = await apiService.getProviders()
    if (result.data && result.data.providers) {
      const providersList = Array.isArray(result.data.providers) ? result.data.providers : []
      setProviders(providersList)
      // Set default provider to 'remi' if available
      if (providersList.length > 0) {
        const remiProvider = providersList.find(p => p.type === 'remi')
        if (remiProvider) {
          setProvider('remi')
        } else {
          setProvider(providersList[0].type)
        }
      }
    }
  }

  const loadPools = async () => {
    setLoading(true)
    setError(null)
    const result = await apiService.getPools()
    if (result.data) {
      setPools(Array.isArray(result.data) ? result.data : [])
    } else {
      setError(result.error || 'Failed to load pools')
      setPools([])
    }
    setLoading(false)
  }

  const loadInstalledVersions = async () => {
    try {
      // Try to get provider-specific versions first
      const providerResult = await apiService.getProviderVersions(provider)
      if (providerResult.data && providerResult.data.versions) {
        const versions = Array.isArray(providerResult.data.versions) ? providerResult.data.versions : []
        setInstalledVersions(versions)
        // Set default version if available and not already set
        if (versions.length > 0 && !phpVersion) {
          setPhpVersion(versions[0])
        }
        return
      }
    } catch {
      // Fallback to all versions if provider-specific fails
    }

    // Fallback to all installed versions
    const result = await apiService.getPhpVersions()
    if (result.data && result.data.versions) {
      const versions = Array.isArray(result.data.versions) ? result.data.versions : []
      setInstalledVersions(versions)
      // Set default version if available and not already set
      if (versions.length > 0 && !phpVersion) {
        setPhpVersion(versions[0])
      }
    } else {
      setInstalledVersions([])
    }
  }

  const handleCreatePool = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!username.trim()) {
      setError('Username is required')
      return
    }

    if (!phpVersion) {
      setError('PHP version is required')
      return
    }

    setCreating(true)
    setError(null)
    setSuccess(null)

    const result = await apiService.createPool(username.trim(), phpVersion, provider)
    
    if (result.data) {
      setSuccess(`Pool for user "${username}" created successfully!`)
      setUsername('')
      setPhpVersion('')
      setProvider('remi')
      setShowCreateForm(false)
      await loadPools()
      await loadInstalledVersions()
    } else {
      setError(result.error || 'Failed to create pool')
    }

    setCreating(false)
  }

  const handleDeletePool = async (username: string) => {
    if (!confirm(`Are you sure you want to delete the pool for user "${username}"?`)) {
      return
    }

    setDeleting(username)
    setError(null)
    setSuccess(null)

    const result = await apiService.deletePool(username)
    
    if (result.data) {
      setSuccess(`Pool for user "${username}" deleted successfully!`)
      await loadPools()
      if (selectedPool?.User === username) {
        setSelectedPool(null)
      }
    } else {
      setError(result.error || 'Failed to delete pool')
    }

    setDeleting(null)
  }

  const handleViewPool = async (username: string) => {
    const result = await apiService.getPool(username)
    if (result.data) {
      setSelectedPool(result.data)
      setShowConfigEditor(false)
    } else {
      setError(result.error || 'Failed to load pool details')
    }
  }

  const handleEditPool = async (username: string) => {
    const result = await apiService.getPool(username)
    if (result.data) {
      setSelectedPool(result.data)
      setShowConfigEditor(true)
      setError(null)
      setSuccess(null)
    } else {
      setError(result.error || 'Failed to load pool details')
    }
  }

  const handleUpdateConfig = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedPool) return

    setUpdatingConfig(true)
    setError(null)
    setSuccess(null)

    const result = await apiService.updatePoolConfig(selectedPool.User, poolConfig)
    
    if (result.data) {
      setSuccess(`Pool configuration updated successfully for ${selectedPool.User}!`)
      setShowConfigEditor(false)
      setPoolConfig({})
      await loadPools()
      await handleViewPool(selectedPool.User)
    } else {
      setError(result.error || 'Failed to update pool configuration')
    }

    setUpdatingConfig(false)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold text-gray-900 mb-2">Pool Management</h2>
          <p className="text-gray-600">Create and manage PHP-FPM pools</p>
        </div>
        <button
          onClick={() => {
            setShowCreateForm(!showCreateForm)
            setError(null)
            setSuccess(null)
            if (!showCreateForm) {
              // Reload versions when opening form
              loadInstalledVersions()
            }
          }}
          className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          {showCreateForm ? 'Cancel' : 'Create Pool'}
        </button>
      </div>

      {/* Create Pool Form */}
      {showCreateForm && (
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <h3 className="text-xl font-semibold text-gray-900 mb-4">Create New Pool</h3>
          
          <form onSubmit={handleCreatePool} className="space-y-4">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-2">
                Username <span className="text-red-600">*</span>
              </label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="john"
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                disabled={creating}
              />
              <p className="mt-2 text-sm text-gray-500">
                System username to create pool for (user must exist)
              </p>
            </div>

            <div>
              <label htmlFor="php-version-pool" className="block text-sm font-medium text-gray-700 mb-2">
                PHP Version <span className="text-red-600">*</span>
              </label>
              {installedVersions.length > 0 ? (
                <select
                  id="php-version-pool"
                  value={phpVersion}
                  onChange={(e) => setPhpVersion(e.target.value)}
                  required
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
                  disabled={creating}
                >
                  <option value="">Select PHP version</option>
                  {installedVersions.map((version) => (
                    <option key={version} value={version}>
                      PHP {version}
                    </option>
                  ))}
                </select>
              ) : (
                <div className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500">
                  No PHP versions installed. Please install a PHP version first.
                </div>
              )}
              <p className="mt-2 text-sm text-gray-500">
                Select an installed PHP version for this pool
              </p>
            </div>

            <div>
              <label htmlFor="provider-pool" className="block text-sm font-medium text-gray-700 mb-2">
                Provider
              </label>
              <select
                id="provider-pool"
                value={provider}
                onChange={(e) => {
                  setProvider(e.target.value)
                  setPhpVersion('') // Reset PHP version when provider changes
                }}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
                disabled={creating}
              >
                {providers.map((p) => (
                  <option key={p.type} value={p.type}>
                    {p.name} {p.status === 'stub' && '(Stub)'}
                  </option>
                ))}
              </select>
              <p className="mt-2 text-sm text-gray-500">
                PHP provider to use (default: remi)
              </p>
            </div>

            <div className="flex gap-4">
              <button
                type="submit"
                disabled={creating}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                {creating ? (
                  <>
                    <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Creating...
                  </>
                ) : (
                  'Create Pool'
                )}
              </button>
            </div>
          </form>
        </div>
      )}

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

      {/* Pools List */}
      <div className="bg-white rounded-lg shadow-md border border-gray-200">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-xl font-semibold text-gray-900">All Pools</h3>
            <button
              onClick={loadPools}
              disabled={loading}
              className="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:bg-gray-50 disabled:text-gray-400 transition-colors flex items-center gap-2"
            >
              <svg className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Refresh
            </button>
          </div>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <svg className="animate-spin h-8 w-8 text-gray-400 mx-auto" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            <p className="mt-2 text-gray-500">Loading pools...</p>
          </div>
        ) : pools.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <p className="text-lg">No pools found</p>
            <p className="text-sm mt-2">Create a pool using the button above.</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {pools.map((pool) => (
              <div key={pool.User} className="p-6 hover:bg-gray-50 transition-colors">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-4">
                      <div>
                        <h4 className="text-lg font-semibold text-gray-900">{pool.User}</h4>
                        <div className="mt-2 flex flex-wrap gap-4 text-sm text-gray-600">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">PHP:</span>
                            <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded">{pool.PHPVersion}</span>
                          </div>
                          {pool.Provider && (
                            <div className="flex items-center gap-2">
                              <span className="font-medium">Provider:</span>
                              <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded">{pool.Provider}</span>
                            </div>
                          )}
                          <div className="flex items-center gap-2">
                            <span className="font-medium">Status:</span>
                            <span className={`px-2 py-1 rounded ${
                              pool.Status === 'active' 
                                ? 'bg-green-100 text-green-800' 
                                : 'bg-gray-100 text-gray-800'
                            }`}>
                              {pool.Status}
                            </span>
                          </div>
                        </div>
                        <div className="mt-3 space-y-1 text-sm text-gray-500">
                          <p><span className="font-medium">Config:</span> <code className="bg-gray-100 px-2 py-0.5 rounded">{pool.ConfigPath}</code></p>
                          <p><span className="font-medium">Socket:</span> <code className="bg-gray-100 px-2 py-0.5 rounded">{pool.SocketPath}</code></p>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 ml-4">
                    <button
                      onClick={() => handleViewPool(pool.User)}
                      className="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
                    >
                      View Details
                    </button>
                    <button
                      onClick={() => handleDeletePool(pool.User)}
                      disabled={deleting === pool.User}
                      className="px-4 py-2 text-sm bg-red-100 text-red-700 rounded-lg hover:bg-red-200 disabled:bg-gray-100 disabled:text-gray-400 transition-colors flex items-center gap-2"
                    >
                      {deleting === pool.User ? (
                        <>
                          <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                          </svg>
                          Deleting...
                        </>
                      ) : (
                        'Delete'
                      )}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Pool Details Modal */}
      {selectedPool && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="text-2xl font-bold text-gray-900">Pool Details</h3>
                <button
                  onClick={() => setSelectedPool(null)}
                  className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-600">User</label>
                <p className="mt-1 text-lg font-semibold text-gray-900">{selectedPool.User}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-600">PHP Version</label>
                <p className="mt-1">
                  <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-lg font-semibold">
                    {selectedPool.PHPVersion}
                  </span>
                </p>
              </div>
              {selectedPool.Provider && (
                <div>
                  <label className="text-sm font-medium text-gray-600">Provider</label>
                  <p className="mt-1">
                    <span className="px-3 py-1 bg-purple-100 text-purple-800 rounded-lg font-semibold">
                      {selectedPool.Provider}
                    </span>
                  </p>
                </div>
              )}
              <div>
                <label className="text-sm font-medium text-gray-600">Status</label>
                <p className="mt-1">
                  <span className={`px-3 py-1 rounded-lg font-semibold ${
                    selectedPool.Status === 'active'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}>
                    {selectedPool.Status}
                  </span>
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-600">Config Path</label>
                <p className="mt-1">
                  <code className="bg-gray-100 px-3 py-2 rounded-lg text-sm block">{selectedPool.ConfigPath}</code>
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-600">Socket Path</label>
                <p className="mt-1">
                  <code className="bg-gray-100 px-3 py-2 rounded-lg text-sm block">{selectedPool.SocketPath}</code>
                </p>
              </div>
            </div>
            {!showConfigEditor ? (
              <div className="p-6 border-t border-gray-200 flex justify-between">
                <button
                  onClick={() => {
                    setShowConfigEditor(true)
                    setError(null)
                    setSuccess(null)
                  }}
                  className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                  Edit Configuration
                </button>
                <button
                  onClick={() => setSelectedPool(null)}
                  className="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
                >
                  Close
                </button>
              </div>
            ) : (
              <div className="p-6 border-t border-gray-200">
                <h4 className="text-lg font-semibold text-gray-900 mb-4">Edit Pool Configuration</h4>
                
                {error && (
                  <div className="mb-4 bg-red-50 border-l-4 border-red-400 p-4 rounded">
                    <p className="text-sm text-red-700">{error}</p>
                  </div>
                )}

                {success && (
                  <div className="mb-4 bg-green-50 border-l-4 border-green-400 p-4 rounded">
                    <p className="text-sm text-green-700">{success}</p>
                  </div>
                )}

                <form onSubmit={handleUpdateConfig} className="space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label htmlFor="max_children" className="block text-sm font-medium text-gray-700 mb-1">
                        Max Children
                      </label>
                      <input
                        type="number"
                        id="max_children"
                        value={poolConfig.max_children || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, max_children: parseInt(e.target.value) || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="50"
                      />
                    </div>
                    <div>
                      <label htmlFor="start_servers" className="block text-sm font-medium text-gray-700 mb-1">
                        Start Servers
                      </label>
                      <input
                        type="number"
                        id="start_servers"
                        value={poolConfig.start_servers || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, start_servers: parseInt(e.target.value) || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="5"
                      />
                    </div>
                    <div>
                      <label htmlFor="min_spare_servers" className="block text-sm font-medium text-gray-700 mb-1">
                        Min Spare Servers
                      </label>
                      <input
                        type="number"
                        id="min_spare_servers"
                        value={poolConfig.min_spare_servers || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, min_spare_servers: parseInt(e.target.value) || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="5"
                      />
                    </div>
                    <div>
                      <label htmlFor="max_spare_servers" className="block text-sm font-medium text-gray-700 mb-1">
                        Max Spare Servers
                      </label>
                      <input
                        type="number"
                        id="max_spare_servers"
                        value={poolConfig.max_spare_servers || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, max_spare_servers: parseInt(e.target.value) || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="35"
                      />
                    </div>
                    <div>
                      <label htmlFor="max_requests" className="block text-sm font-medium text-gray-700 mb-1">
                        Max Requests
                      </label>
                      <input
                        type="number"
                        id="max_requests"
                        value={poolConfig.max_requests || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, max_requests: parseInt(e.target.value) || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="500"
                      />
                    </div>
                    <div>
                      <label htmlFor="process_manager" className="block text-sm font-medium text-gray-700 mb-1">
                        Process Manager
                      </label>
                      <select
                        id="process_manager"
                        value={poolConfig.process_manager || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, process_manager: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
                      >
                        <option value="">Select...</option>
                        <option value="dynamic">Dynamic</option>
                        <option value="static">Static</option>
                        <option value="ondemand">Ondemand</option>
                      </select>
                    </div>
                    <div>
                      <label htmlFor="memory_limit" className="block text-sm font-medium text-gray-700 mb-1">
                        Memory Limit
                      </label>
                      <input
                        type="text"
                        id="memory_limit"
                        value={poolConfig.memory_limit || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, memory_limit: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="128M"
                      />
                    </div>
                    <div>
                      <label htmlFor="max_execution_time" className="block text-sm font-medium text-gray-700 mb-1">
                        Max Execution Time
                      </label>
                      <input
                        type="text"
                        id="max_execution_time"
                        value={poolConfig.max_execution_time || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, max_execution_time: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="300"
                      />
                    </div>
                    <div>
                      <label htmlFor="upload_max_filesize" className="block text-sm font-medium text-gray-700 mb-1">
                        Upload Max Filesize
                      </label>
                      <input
                        type="text"
                        id="upload_max_filesize"
                        value={poolConfig.upload_max_filesize || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, upload_max_filesize: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="64M"
                      />
                    </div>
                    <div>
                      <label htmlFor="post_max_size" className="block text-sm font-medium text-gray-700 mb-1">
                        Post Max Size
                      </label>
                      <input
                        type="text"
                        id="post_max_size"
                        value={poolConfig.post_max_size || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, post_max_size: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="64M"
                      />
                    </div>
                    <div>
                      <label htmlFor="display_errors" className="block text-sm font-medium text-gray-700 mb-1">
                        Display Errors
                      </label>
                      <select
                        id="display_errors"
                        value={typeof poolConfig.display_errors === 'boolean' ? (poolConfig.display_errors ? 'on' : 'off') : (poolConfig.display_errors || '')}
                        onChange={(e) => setPoolConfig({ ...poolConfig, display_errors: e.target.value === 'on' ? true : e.target.value === 'off' ? false : e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
                      >
                        <option value="">Select...</option>
                        <option value="on">On</option>
                        <option value="off">Off</option>
                      </select>
                    </div>
                    <div>
                      <label htmlFor="log_errors" className="block text-sm font-medium text-gray-700 mb-1">
                        Log Errors
                      </label>
                      <select
                        id="log_errors"
                        value={typeof poolConfig.log_errors === 'boolean' ? (poolConfig.log_errors ? 'on' : 'off') : (poolConfig.log_errors || '')}
                        onChange={(e) => setPoolConfig({ ...poolConfig, log_errors: e.target.value === 'on' ? true : e.target.value === 'off' ? false : e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white"
                      >
                        <option value="">Select...</option>
                        <option value="on">On</option>
                        <option value="off">Off</option>
                      </select>
                    </div>
                    <div>
                      <label htmlFor="date_timezone" className="block text-sm font-medium text-gray-700 mb-1">
                        Date Timezone
                      </label>
                      <input
                        type="text"
                        id="date_timezone"
                        value={poolConfig.date_timezone || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, date_timezone: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="UTC"
                      />
                    </div>
                    <div>
                      <label htmlFor="sendmail_path" className="block text-sm font-medium text-gray-700 mb-1">
                        Sendmail Path
                      </label>
                      <input
                        type="text"
                        id="sendmail_path"
                        value={poolConfig.sendmail_path || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, sendmail_path: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="/usr/sbin/sendmail"
                      />
                    </div>
                    <div>
                      <label htmlFor="process_idle_timeout" className="block text-sm font-medium text-gray-700 mb-1">
                        Process Idle Timeout
                      </label>
                      <input
                        type="text"
                        id="process_idle_timeout"
                        value={poolConfig.process_idle_timeout || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, process_idle_timeout: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="10s"
                      />
                    </div>
                    <div>
                      <label htmlFor="listen_mode" className="block text-sm font-medium text-gray-700 mb-1">
                        Listen Mode
                      </label>
                      <input
                        type="text"
                        id="listen_mode"
                        value={poolConfig.listen_mode || ''}
                        onChange={(e) => setPoolConfig({ ...poolConfig, listen_mode: e.target.value || undefined })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="0660"
                      />
                    </div>
                  </div>
                  <div className="flex justify-end gap-3 pt-4">
                    <button
                      type="button"
                      onClick={() => {
                        setShowConfigEditor(false)
                        setPoolConfig({})
                        setError(null)
                        setSuccess(null)
                      }}
                      className="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
                      disabled={updatingConfig}
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={updatingConfig}
                      className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
                    >
                      {updatingConfig ? (
                        <>
                          <svg className="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                          </svg>
                          Updating...
                        </>
                      ) : (
                        'Update Configuration'
                      )}
                    </button>
                  </div>
                </form>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
