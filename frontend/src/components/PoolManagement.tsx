import { useState, useEffect } from 'react'
import { apiService } from '../services/api'
import type { Pool } from '../services/api'

export default function PoolManagement() {
  const [pools, setPools] = useState<Pool[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [selectedPool, setSelectedPool] = useState<Pool | null>(null)
  const [deleting, setDeleting] = useState<string | null>(null)
  
  // Create form state
  const [username, setUsername] = useState('')
  const [phpVersion, setPhpVersion] = useState('8.2')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    loadPools()
  }, [])

  const loadPools = async () => {
    setLoading(true)
    setError(null)
    const result = await apiService.getPools()
    if (result.data) {
      setPools(result.data)
    } else {
      setError(result.error || 'Failed to load pools')
    }
    setLoading(false)
  }

  const handleCreatePool = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!username.trim()) {
      setError('Username is required')
      return
    }

    setCreating(true)
    setError(null)
    setSuccess(null)

    const result = await apiService.createPool(username.trim(), phpVersion)
    
    if (result.data) {
      setSuccess(`Pool for user "${username}" created successfully!`)
      setUsername('')
      setPhpVersion('8.2')
      setShowCreateForm(false)
      await loadPools()
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
    } else {
      setError(result.error || 'Failed to load pool details')
    }
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
                PHP Version
              </label>
              <input
                type="text"
                id="php-version-pool"
                value={phpVersion}
                onChange={(e) => setPhpVersion(e.target.value)}
                placeholder="8.2"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                disabled={creating}
              />
              <p className="mt-2 text-sm text-gray-500">
                PHP version to use (default: 8.2)
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
            <div className="p-6 border-t border-gray-200 flex justify-end">
              <button
                onClick={() => setSelectedPool(null)}
                className="px-6 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
