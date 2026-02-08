// Use relative URLs in development (via Vite proxy) or absolute URL in production
const API_BASE_URL = import.meta.env.PROD 
  ? 'http://95.217.213.18:8981'
  : '' // Empty string means use relative URLs (goes through Vite proxy)

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
        ...options,
      })

      const data = await response.json()

      if (!response.ok) {
        return {
          error: data.error || `HTTP ${response.status}: ${response.statusText}`,
        }
      }

      return { data }
    } catch (error) {
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
