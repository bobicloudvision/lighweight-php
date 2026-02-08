import { useEffect, useState } from 'react'
import { apiService } from '../services/api'

export default function Dashboard() {
  const [healthStatus, setHealthStatus] = useState<'checking' | 'ok' | 'error'>('checking')
  const [poolsCount, setPoolsCount] = useState<number | null>(null)
  const [phpVersionsCount, setPhpVersionsCount] = useState<number | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      // Check health
      const health = await apiService.checkHealth()
      setHealthStatus(health.data ? 'ok' : 'error')

      // Get pools count
      const pools = await apiService.getPools()
      if (pools.data) {
        setPoolsCount(pools.data.length)
      }

      // Get PHP versions count
      const versions = await apiService.getPhpVersions()
      if (versions.data) {
        setPhpVersionsCount(versions.data.versions.length)
      }
    }

    fetchData()
    const interval = setInterval(fetchData, 30000) // Refresh every 30 seconds
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold text-gray-900 mb-2">Dashboard</h2>
        <p className="text-gray-600">Overview of your Lightweight PHP API</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Health Status Card */}
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">API Status</p>
              <p className="text-2xl font-bold mt-2">
                {healthStatus === 'checking' && <span className="text-gray-400">Checking...</span>}
                {healthStatus === 'ok' && <span className="text-green-600">Online</span>}
                {healthStatus === 'error' && <span className="text-red-600">Offline</span>}
              </p>
            </div>
            <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
              healthStatus === 'ok' ? 'bg-green-100' : 
              healthStatus === 'error' ? 'bg-red-100' : 
              'bg-gray-100'
            }`}>
              {healthStatus === 'ok' && (
                <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              )}
              {healthStatus === 'error' && (
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              )}
              {healthStatus === 'checking' && (
                <svg className="w-6 h-6 text-gray-400 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              )}
            </div>
          </div>
        </div>

        {/* PHP Versions Card */}
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">PHP Versions</p>
              <p className="text-2xl font-bold mt-2 text-gray-900">
                {phpVersionsCount !== null ? phpVersionsCount : '...'}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
            </div>
          </div>
        </div>

        {/* Pools Card */}
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Active Pools</p>
              <p className="text-2xl font-bold mt-2 text-gray-900">
                {poolsCount !== null ? poolsCount : '...'}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-purple-100 flex items-center justify-center">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-blue-50 border-l-4 border-blue-400 p-4 rounded">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-sm text-blue-700">
              <strong>Base URL:</strong> <code className="bg-blue-100 px-2 py-1 rounded text-sm">http://localhost:8989</code>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
