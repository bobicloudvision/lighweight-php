export const API_BASE_URL = 'http://95.217.213.18:8989'

export interface Pool {
  User: string
  PHPVersion: string
  Status: string
  ConfigPath: string
  SocketPath: string
}

export interface ApiResponse<T> {
  data?: T
  error?: string
}

class ApiService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        mode: 'cors',
        ...options,
      })

      // Check if response is ok before trying to parse JSON
      if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`
        try {
          const errorData = await response.json()
          if (errorData.error) {
            errorMessage = errorData.error
          }
        } catch {
          // If response is not JSON, use the status text
        }
        return { error: errorMessage }
      }

      // Try to parse JSON, but handle cases where response might not be JSON
      let data: T
      try {
        const contentType = response.headers.get('content-type')
        if (contentType && contentType.includes('application/json')) {
          data = await response.json()
        } else {
          const text = await response.text()
          // Try to parse as JSON anyway
          try {
            data = JSON.parse(text) as T
          } catch {
            return { error: `Invalid response format: ${text.substring(0, 100)}` }
          }
        }
      } catch (parseError) {
        return {
          error: 'Failed to parse response as JSON',
        }
      }

      return { data }
    } catch (error) {
      // Handle CORS and network errors
      if (error instanceof TypeError) {
        if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
          return {
            error: `CORS or network error: Unable to connect to ${API_BASE_URL}. Please ensure the server is running and CORS is properly configured.`,
          }
        }
        if (error.message.includes('CORS')) {
          return {
            error: `CORS error: The server at ${API_BASE_URL} is not allowing requests from this origin. Please configure CORS headers on the server.`,
          }
        }
      }
      
      return {
        error: error instanceof Error ? error.message : 'Unknown error occurred',
      }
    }
  }

  async checkHealth(): Promise<ApiResponse<{ status: string }>> {
    return this.request<{ status: string }>('/health')
  }

  async installPhpVersion(version: string): Promise<ApiResponse<{ message: string; version: string }>> {
    return this.request<{ message: string; version: string }>(
      `/api/v1/php/install/${version}`,
      { method: 'POST' }
    )
  }

  async getPhpVersions(): Promise<ApiResponse<{ versions: string[] }>> {
    return this.request<{ versions: string[] }>('/api/v1/php/versions')
  }

  async getAvailablePhpVersions(): Promise<ApiResponse<{ versions: string[] }>> {
    return this.request<{ versions: string[] }>('/api/v1/php/available')
  }

  async getPools(): Promise<ApiResponse<Pool[]>> {
    return this.request<Pool[]>('/api/v1/pools')
  }

  async getPool(username: string): Promise<ApiResponse<Pool>> {
    return this.request<Pool>(`/api/v1/pools/${username}`)
  }

  async createPool(username: string, phpVersion: string = '8.2'): Promise<ApiResponse<{ message: string; username: string }>> {
    return this.request<{ message: string; username: string }>(
      '/api/v1/pools',
      {
        method: 'POST',
        body: JSON.stringify({ username, php_version: phpVersion }),
      }
    )
  }

  async deletePool(username: string): Promise<ApiResponse<{ message: string; username: string }>> {
    return this.request<{ message: string; username: string }>(
      `/api/v1/pools/${username}`,
      { method: 'DELETE' }
    )
  }
}

export const apiService = new ApiService()
