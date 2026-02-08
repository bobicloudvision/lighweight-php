import { useState, useEffect } from 'react'
import { apiService, API_BASE_URL } from '../services/api'

export default function CorsWarning() {
  const [showWarning, setShowWarning] = useState(false)

  useEffect(() => {
    const checkCors = async () => {
      const result = await apiService.checkHealth()
      if (result.error && (result.error.includes('CORS') || result.error.includes('Unable to connect'))) {
        setShowWarning(true)
      } else {
        setShowWarning(false)
      }
    }

    checkCors()
    const interval = setInterval(checkCors, 10000) // Check every 10 seconds
    return () => clearInterval(interval)
  }, [])

  if (!showWarning) return null

  return (
    <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-6 rounded">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium text-red-800">CORS Configuration Required</h3>
          <div className="mt-2 text-sm text-red-700">
            <p>
              The API server at <code className="bg-red-100 px-2 py-1 rounded text-xs">{API_BASE_URL}</code> is not configured to allow requests from this origin.
            </p>
            <p className="mt-2">
              <strong>To fix this, configure CORS headers on your backend server:</strong>
            </p>
            <ul className="mt-2 list-disc list-inside space-y-1">
              <li>Add <code className="bg-red-100 px-1 py-0.5 rounded text-xs">Access-Control-Allow-Origin: *</code> header (or specific origin)</li>
              <li>Add <code className="bg-red-100 px-1 py-0.5 rounded text-xs">Access-Control-Allow-Methods: GET, POST, DELETE, OPTIONS</code></li>
              <li>Add <code className="bg-red-100 px-1 py-0.5 rounded text-xs">Access-Control-Allow-Headers: Content-Type</code></li>
              <li>Handle OPTIONS preflight requests</li>
            </ul>
            <p className="mt-2 text-xs">
              <strong>Note:</strong> If using HTTPS-Only Mode in Firefox, ensure your server supports HTTPS or disable HTTPS-Only Mode for this site.
            </p>
          </div>
        </div>
        <div className="ml-4 flex-shrink-0">
          <button
            onClick={() => setShowWarning(false)}
            className="text-red-400 hover:text-red-600"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  )
}
