import { useState } from 'react'

interface Endpoint {
  method: string
  path: string
  title: string
  description: string
  parameters?: Array<{ name: string; type: string; description: string; required?: boolean }>
  requestBody?: { fields: Array<{ name: string; type: string; description: string; required?: boolean }> }
  response: { status: number; body: string }
  errorResponses?: Array<{ status: number; body: string }>
  example: string
}

function App() {
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null)

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedCommand(id)
    setTimeout(() => setCopiedCommand(null), 2000)
  }

  const endpoints: Endpoint[] = [
    {
      method: 'GET',
      path: '/health',
      title: 'Health Check',
      description: 'Check if the API server is running.',
      response: { status: 200, body: JSON.stringify({ status: 'ok' }, null, 2) },
      example: 'curl http://localhost:8080/health'
    },
    {
      method: 'POST',
      path: '/api/v1/php/install/{version}',
      title: 'Install PHP Version',
      description: 'Install a PHP version from Remi repository (RHEL) or ondrej PPA (Debian).',
      parameters: [
        { name: 'version', type: 'path parameter', description: 'PHP version to install (e.g., 8.2, 8.1, 8.3)', required: true }
      ],
      response: { status: 200, body: JSON.stringify({ message: 'PHP installed successfully', version: '8.2' }, null, 2) },
      errorResponses: [
        { status: 500, body: JSON.stringify({ error: 'failed to install PHP packages: ...' }, null, 2) }
      ],
      example: 'curl -X POST http://localhost:8080/api/v1/php/install/8.2'
    },
    {
      method: 'GET',
      path: '/api/v1/php/versions',
      title: 'List PHP Versions',
      description: 'List all installed PHP versions.',
      response: { status: 200, body: JSON.stringify({ versions: ['8.2', '8.1', '8.3'] }, null, 2) },
      errorResponses: [
        { status: 500, body: JSON.stringify({ error: 'failed to list PHP versions: ...' }, null, 2) }
      ],
      example: 'curl http://localhost:8080/api/v1/php/versions'
    },
    {
      method: 'GET',
      path: '/api/v1/pools',
      title: 'List All Pools',
      description: 'List all PHP-FPM pools.',
      response: {
        status: 200,
        body: JSON.stringify([
          {
            User: 'john',
            PHPVersion: '8.2',
            Status: 'active',
            ConfigPath: '/etc/php-fpm.d/john.conf',
            SocketPath: '/var/run/php-fpm/john.sock'
          },
          {
            User: 'jane',
            PHPVersion: '8.1',
            Status: 'active',
            ConfigPath: '/etc/php/8.1/fpm/pool.d/jane.conf',
            SocketPath: '/var/run/php/php8.1-jane.sock'
          }
        ], null, 2)
      },
      errorResponses: [
        { status: 500, body: JSON.stringify({ error: 'failed to list pools from database: ...' }, null, 2) }
      ],
      example: 'curl http://localhost:8080/api/v1/pools'
    },
    {
      method: 'POST',
      path: '/api/v1/pools',
      title: 'Create Pool',
      description: 'Create a new PHP-FPM pool for a user.',
      requestBody: {
        fields: [
          { name: 'username', type: 'string', description: 'System username to create pool for', required: true },
          { name: 'php_version', type: 'string', description: 'PHP version to use (default: 8.2)', required: false }
        ]
      },
      response: { status: 201, body: JSON.stringify({ message: 'Pool created successfully', username: 'john' }, null, 2) },
      errorResponses: [
        { status: 400, body: JSON.stringify({ error: 'username is required' }, null, 2) },
        { status: 400, body: JSON.stringify({ error: 'Invalid request body' }, null, 2) },
        { status: 500, body: JSON.stringify({ error: 'user john does not exist: user: unknown user john' }, null, 2) },
        { status: 500, body: JSON.stringify({ error: 'failed to save pool to database: ...' }, null, 2) }
      ],
      example: `curl -X POST http://localhost:8080/api/v1/pools \\
  -H "Content-Type: application/json" \\
  -d '{
    "username": "john",
    "php_version": "8.2"
  }'`
    },
    {
      method: 'GET',
      path: '/api/v1/pools/{username}',
      title: 'Get Pool Information',
      description: 'Get pool information for a specific user.',
      parameters: [
        { name: 'username', type: 'path parameter', description: 'Username to get pool for', required: true }
      ],
      response: {
        status: 200,
        body: JSON.stringify({
          User: 'john',
          PHPVersion: '8.2',
          Status: 'active',
          ConfigPath: '/etc/php-fpm.d/john.conf',
          SocketPath: '/var/run/php-fpm/john.sock'
        }, null, 2)
      },
      errorResponses: [
        { status: 404, body: JSON.stringify({ error: 'Pool not found' }, null, 2) },
        { status: 500, body: JSON.stringify({ error: 'failed to list pools from database: ...' }, null, 2) }
      ],
      example: 'curl http://localhost:8080/api/v1/pools/john'
    },
    {
      method: 'DELETE',
      path: '/api/v1/pools/{username}',
      title: 'Delete Pool',
      description: 'Delete a PHP-FPM pool for a user.',
      parameters: [
        { name: 'username', type: 'path parameter', description: 'Username to delete pool for', required: true }
      ],
      response: { status: 200, body: JSON.stringify({ message: 'Pool deleted successfully', username: 'john' }, null, 2) },
      errorResponses: [
        { status: 500, body: JSON.stringify({ error: 'pool for user john not found' }, null, 2) },
        { status: 500, body: JSON.stringify({ error: 'failed to delete pool from database: ...' }, null, 2) }
      ],
      example: 'curl -X DELETE http://localhost:8080/api/v1/pools/john'
    }
  ]

  const getMethodColor = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'bg-blue-100 text-blue-800 border-blue-300',
      POST: 'bg-green-100 text-green-800 border-green-300',
      PUT: 'bg-yellow-100 text-yellow-800 border-yellow-300',
      DELETE: 'bg-red-100 text-red-800 border-red-300',
      PATCH: 'bg-purple-100 text-purple-800 border-purple-300'
    }
    return colors[method] || 'bg-gray-100 text-gray-800 border-gray-300'
  }

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'text-green-600'
    if (status >= 400 && status < 500) return 'text-yellow-600'
    if (status >= 500) return 'text-red-600'
    return 'text-gray-600'
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Lightweight PHP REST API</h1>
          <p className="text-lg text-gray-600">Complete API documentation and reference</p>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Base URL Section */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8 border border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Base URL</h2>
          <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
            <code className="text-lg font-mono text-gray-800">http://localhost:8080</code>
          </div>
          <p className="mt-4 text-gray-600">
            Default server runs on <code className="bg-gray-100 px-2 py-1 rounded text-sm">0.0.0.0:8080</code>. 
            You can change the host and port using CLI flags:
          </p>
          <div className="mt-3 bg-gray-50 rounded-lg p-4 border border-gray-200">
            <code className="font-mono text-gray-800">./lightweight-php server --host 127.0.0.1 --port 8080</code>
          </div>
        </div>

        {/* Authentication Notice */}
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-8 rounded">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-yellow-700">
                <strong>Note:</strong> Currently, the API does not require authentication. 
                <strong className="ml-1">In production, you should add authentication/authorization.</strong>
              </p>
            </div>
          </div>
        </div>

        {/* Endpoints Section */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-6">Endpoints</h2>
          <div className="space-y-8">
            {endpoints.map((endpoint, index) => (
              <div key={index} className="bg-white rounded-lg shadow-md overflow-hidden border border-gray-200">
                <div className="p-6">
                  {/* Endpoint Header */}
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <span className={`px-3 py-1 rounded-md font-semibold text-sm border ${getMethodColor(endpoint.method)}`}>
                          {endpoint.method}
                        </span>
                        <code className="text-lg font-mono text-gray-800 bg-gray-50 px-3 py-1 rounded border border-gray-200">
                          {endpoint.path}
                        </code>
                      </div>
                      <h3 className="text-xl font-semibold text-gray-900 mt-2">{endpoint.title}</h3>
                      <p className="text-gray-600 mt-1">{endpoint.description}</p>
                    </div>
                  </div>

                  {/* Parameters */}
                  {endpoint.parameters && endpoint.parameters.length > 0 && (
                    <div className="mt-6">
                      <h4 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">Parameters</h4>
                      <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {endpoint.parameters.map((param, idx) => (
                              <tr key={idx}>
                                <td className="px-4 py-3 whitespace-nowrap">
                                  <code className="text-sm font-mono text-gray-800">{param.name}</code>
                                  {param.required && <span className="ml-2 text-xs text-red-600">required</span>}
                                </td>
                                <td className="px-4 py-3 text-sm text-gray-600">{param.type}</td>
                                <td className="px-4 py-3 text-sm text-gray-600">{param.description}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Request Body */}
                  {endpoint.requestBody && (
                    <div className="mt-6">
                      <h4 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">Request Body</h4>
                      <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                          <thead className="bg-gray-50">
                            <tr>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Field</th>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                            </tr>
                          </thead>
                          <tbody className="bg-white divide-y divide-gray-200">
                            {endpoint.requestBody.fields.map((field, idx) => (
                              <tr key={idx}>
                                <td className="px-4 py-3 whitespace-nowrap">
                                  <code className="text-sm font-mono text-gray-800">{field.name}</code>
                                  {field.required && <span className="ml-2 text-xs text-red-600">required</span>}
                                </td>
                                <td className="px-4 py-3 text-sm text-gray-600">{field.type}</td>
                                <td className="px-4 py-3 text-sm text-gray-600">{field.description}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Response */}
                  <div className="mt-6">
                    <h4 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">
                      Response <span className={`font-normal ${getStatusColor(endpoint.response.status)}`}>({endpoint.response.status})</span>
                    </h4>
                    <div className="relative">
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
                        <code>{endpoint.response.body}</code>
                      </pre>
                      <button
                        onClick={() => copyToClipboard(endpoint.response.body, `response-${index}`)}
                        className="absolute top-2 right-2 bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-xs transition-colors"
                      >
                        {copiedCommand === `response-${index}` ? 'Copied!' : 'Copy'}
                      </button>
                    </div>
                  </div>

                  {/* Error Responses */}
                  {endpoint.errorResponses && endpoint.errorResponses.length > 0 && (
                    <div className="mt-6">
                      <h4 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">Error Responses</h4>
                      <div className="space-y-4">
                        {endpoint.errorResponses.map((error, idx) => (
                          <div key={idx}>
                            <span className={`text-sm font-medium ${getStatusColor(error.status)}`}>
                              {error.status} {error.status === 400 ? 'Bad Request' : error.status === 404 ? 'Not Found' : 'Internal Server Error'}
                            </span>
                            <div className="relative mt-2">
                              <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
                                <code>{error.body}</code>
                              </pre>
                              <button
                                onClick={() => copyToClipboard(error.body, `error-${index}-${idx}`)}
                                className="absolute top-2 right-2 bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-xs transition-colors"
                              >
                                {copiedCommand === `error-${index}-${idx}` ? 'Copied!' : 'Copy'}
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Example */}
                  <div className="mt-6">
                    <h4 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">Example</h4>
                    <div className="relative">
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
                        <code>{endpoint.example}</code>
                      </pre>
                      <button
                        onClick={() => copyToClipboard(endpoint.example, `example-${index}`)}
                        className="absolute top-2 right-2 bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-xs transition-colors"
                      >
                        {copiedCommand === `example-${index}` ? 'Copied!' : 'Copy'}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Status Codes Section */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8 border border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Response Status Codes</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 rounded-md font-semibold text-sm bg-green-100 text-green-800 border border-green-300">200</span>
              <span className="text-gray-700">OK - Request successful</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 rounded-md font-semibold text-sm bg-green-100 text-green-800 border border-green-300">201</span>
              <span className="text-gray-700">Created - Resource created successfully</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 rounded-md font-semibold text-sm bg-yellow-100 text-yellow-800 border border-yellow-300">400</span>
              <span className="text-gray-700">Bad Request - Invalid request parameters or body</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 rounded-md font-semibold text-sm bg-yellow-100 text-yellow-800 border border-yellow-300">404</span>
              <span className="text-gray-700">Not Found - Resource not found</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="px-3 py-1 rounded-md font-semibold text-sm bg-red-100 text-red-800 border border-red-300">500</span>
              <span className="text-gray-700">Internal Server Error - Server error occurred</span>
            </div>
          </div>
        </div>

        {/* Error Format Section */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8 border border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Error Response Format</h2>
          <p className="text-gray-600 mb-4">All error responses follow this format:</p>
          <div className="relative">
            <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
              <code>{JSON.stringify({ error: 'Error message description' }, null, 2)}</code>
            </pre>
            <button
              onClick={() => copyToClipboard(JSON.stringify({ error: 'Error message description' }, null, 2), 'error-format')}
              className="absolute top-2 right-2 bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-xs transition-colors"
            >
              {copiedCommand === 'error-format' ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>

        {/* Complete Workflow Section */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8 border border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Complete Workflow</h2>
          <div className="space-y-6">
            {[
              { step: 1, title: 'Install PHP 8.2', command: 'curl -X POST http://localhost:8080/api/v1/php/install/8.2' },
              { step: 2, title: 'List installed PHP versions', command: 'curl http://localhost:8080/api/v1/php/versions' },
              { step: 3, title: 'Create a pool for user "john"', command: `curl -X POST http://localhost:8080/api/v1/pools \\\n  -H "Content-Type: application/json" \\\n  -d '{"username": "john", "php_version": "8.2"}'` },
              { step: 4, title: 'Get pool information', command: 'curl http://localhost:8080/api/v1/pools/john' },
              { step: 5, title: 'List all pools', command: 'curl http://localhost:8080/api/v1/pools' },
              { step: 6, title: 'Delete pool', command: 'curl -X DELETE http://localhost:8080/api/v1/pools/john' }
            ].map((item) => (
              <div key={item.step} className="border-l-4 border-blue-500 pl-4">
                <div className="flex items-center gap-2 mb-2">
                  <span className="flex items-center justify-center w-8 h-8 rounded-full bg-blue-100 text-blue-800 font-semibold text-sm">
                    {item.step}
                  </span>
                  <h3 className="text-lg font-semibold text-gray-900">{item.title}</h3>
                </div>
                <div className="relative mt-2">
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm font-mono">
                    <code>{item.command}</code>
                  </pre>
                  <button
                    onClick={() => copyToClipboard(item.command, `workflow-${item.step}`)}
                    className="absolute top-2 right-2 bg-gray-700 hover:bg-gray-600 text-white px-3 py-1 rounded text-xs transition-colors"
                  >
                    {copiedCommand === `workflow-${item.step}` ? 'Copied!' : 'Copy'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Notes Section */}
        <div className="bg-blue-50 border-l-4 border-blue-400 p-6 rounded mb-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Notes</h2>
          <ul className="space-y-2 text-gray-700">
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>All timestamps are in ISO 8601 format (UTC)</span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>Pool socket paths differ between RHEL and Debian systems</span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>PHP-FPM services are automatically reloaded after pool creation/deletion</span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>The user must exist in the system before creating a pool</span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>Pool configurations are stored in:
                <ul className="ml-6 mt-1 space-y-1">
                  <li>• RHEL: <code className="bg-blue-100 px-1 py-0.5 rounded text-sm">/etc/php-fpm.d/&#123;username&#125;.conf</code></li>
                  <li>• Debian: <code className="bg-blue-100 px-1 py-0.5 rounded text-sm">/etc/php/&#123;version&#125;/fpm/pool.d/&#123;username&#125;.conf</code></li>
                </ul>
              </span>
            </li>
          </ul>
        </div>

        {/* Database Section */}
        <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Database</h2>
          <p className="text-gray-600 mb-3">
            The application uses SQLite database located at:
          </p>
          <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
            <code className="font-mono text-gray-800">/var/lib/lightweight-php/lightweight-php.db</code>
          </div>
          <p className="mt-3 text-gray-600">
            All pools and PHP installations are tracked in the database for persistence and querying.
          </p>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-gray-600">Lightweight PHP REST API Documentation</p>
        </div>
      </footer>
    </div>
  )
}

export default App
