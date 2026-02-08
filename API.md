# Lightweight PHP REST API Documentation

## Base URL

```
http://localhost:8080
```

Default server runs on `0.0.0.0:8080`. You can change the host and port using CLI flags:
```bash
./lightweight-php server --host 127.0.0.1 --port 8080
```

## Authentication

Currently, the API does not require authentication. **Note:** In production, you should add authentication/authorization.

## Endpoints

### Health Check

#### GET /health

Check if the API server is running.

**Response:**
```json
{
  "status": "ok"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

### PHP Version Management

#### POST /api/v1/php/install/{version}

Install a PHP version from Remi repository (RHEL) or ondrej PPA (Debian).

**Parameters:**
- `version` (path parameter) - PHP version to install (e.g., `8.2`, `8.1`, `8.3`)
- `provider` (query parameter, optional) - PHP provider type: `remi`, `lsphp`, `alt-php`, or `docker` (default: `remi`)

**Response:**
```json
{
  "message": "PHP installed successfully",
  "version": "8.2",
  "provider": "remi"
}
```

**Examples:**
```bash
# Install with default provider (remi)
curl -X POST http://localhost:8080/api/v1/php/install/8.2

# Install with specific provider
curl -X POST http://localhost:8080/api/v1/php/install/8.2?provider=lsphp
curl -X POST http://localhost:8080/api/v1/php/install/8.2?provider=alt-php
curl -X POST http://localhost:8080/api/v1/php/install/8.2?provider=docker
```

**Error Response (500):**
```json
{
  "error": "failed to install PHP packages: ..."
}
```

---

#### GET /api/v1/php/versions

List all installed PHP versions.

**Response:**
```json
{
  "versions": ["8.2", "8.1", "8.3"]
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/php/versions
```

**Error Response (500):**
```json
{
  "error": "failed to list PHP versions: ..."
}
```

---

#### GET /api/v1/php/available

List all available PHP versions that can be installed from the repository (Remi for RHEL, ondrej PPA for Debian).

**Response:**
```json
{
  "versions": ["8.3", "8.2", "8.1", "8.0", "7.4", "7.3", "7.2"]
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/php/available
```

**Notes:**
- This endpoint queries the package repositories to find available PHP versions
- For RHEL systems, it checks Remi repository for available PHP versions
- For Debian/Ubuntu systems, it checks ondrej PPA for available PHP versions
- If repository queries fail, it returns a list of commonly available versions
- The list may vary depending on your system's repository configuration

**Error Response (500):**
```json
{
  "error": "failed to list available PHP versions: ..."
}
```

---

### Pool Management

#### GET /api/v1/pools

List all PHP-FPM pools.

**Response:**
```json
[
  {
    "User": "john",
    "PHPVersion": "8.2",
    "Provider": "remi",
    "Status": "active",
    "ConfigPath": "/etc/php-fpm.d/john.conf",
    "SocketPath": "/var/run/php-fpm/john.sock"
  },
  {
    "User": "jane",
    "PHPVersion": "8.3",
    "Provider": "lsphp",
    "Status": "active",
    "ConfigPath": "/usr/local/lsws/lsphp83/etc/php-fpm.d/jane.conf",
    "SocketPath": "/var/run/lsphp/jane.sock"
  }
]
```

**Example:**
```bash
curl http://localhost:8080/api/v1/pools
```

**Error Response (500):**
```json
{
  "error": "failed to list pools from database: ..."
}
```

---

#### POST /api/v1/pools

Create a new PHP-FPM pool for a user. **Provider must be specified** to determine which PHP backend to use.

**Request Body:**
```json
{
  "username": "john",
  "php_version": "8.2",
  "provider": "remi"
}
```

**Fields:**
- `username` (required) - System username to create pool for
- `php_version` (optional) - PHP version to use (default: `8.2`)
- `provider` (optional) - PHP provider type: `remi`, `lsphp`, `alt-php`, or `docker` (default: `remi`)

**Response (201):**
```json
{
  "message": "Pool created successfully",
  "username": "john"
}
```

**Examples:**
```bash
# Create pool with default provider (remi)
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "php_version": "8.2"
  }'

# Create pool with LiteSpeed PHP provider
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "username": "jane",
    "php_version": "8.3",
    "provider": "lsphp"
  }'
```

**Error Responses:**

**400 Bad Request:**
```json
{
  "error": "username is required"
}
```

```json
{
  "error": "Invalid request body"
}
```

**500 Internal Server Error:**
```json
{
  "error": "user john does not exist: user: unknown user john"
}
```

```json
{
  "error": "invalid provider type: invalid. Supported: remi, lsphp, alt-php, docker"
}
```

```json
{
  "error": "failed to save pool to database: ..."
}
```

---

#### GET /api/v1/pools/{username}

Get pool information for a specific user.

**Parameters:**
- `username` (path parameter) - Username to get pool for

**Response (200):**
```json
{
  "User": "john",
  "PHPVersion": "8.2",
  "Provider": "remi",
  "Status": "active",
  "ConfigPath": "/etc/php-fpm.d/john.conf",
  "SocketPath": "/var/run/php-fpm/john.sock"
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/pools/john
```

**Error Responses:**

**404 Not Found:**
```json
{
  "error": "Pool not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "failed to list pools from database: ..."
}
```

---

#### DELETE /api/v1/pools/{username}

Delete a PHP-FPM pool for a user.

**Parameters:**
- `username` (path parameter) - Username to delete pool for

**Response (200):**
```json
{
  "message": "Pool deleted successfully",
  "username": "john"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/pools/john
```

**Error Responses:**

**500 Internal Server Error:**
```json
{
  "error": "pool for user john not found"
}
```

```json
{
  "error": "failed to delete pool from database: ..."
}
```

---

### Provider Management

#### GET /api/v1/providers

List all available PHP providers.

**Response:**
```json
{
  "providers": [
    {
      "type": "remi",
      "name": "Remi Repository",
      "description": "Remi repository for RHEL, ondrej PPA for Debian",
      "status": "active"
    },
    {
      "type": "lsphp",
      "name": "LiteSpeed PHP",
      "description": "LiteSpeed Web Server PHP",
      "status": "stub"
    },
    {
      "type": "alt-php",
      "name": "Alternative PHP",
      "description": "Alternative PHP packages",
      "status": "stub"
    },
    {
      "type": "docker",
      "name": "Docker PHP",
      "description": "Docker-hosted PHP containers",
      "status": "stub"
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/providers
```

---

#### POST /api/v1/providers/{provider}/install/{version}

Install a PHP version using a specific provider.

**Parameters:**
- `provider` (path parameter) - Provider type: `remi`, `lsphp`, `alt-php`, or `docker`
- `version` (path parameter) - PHP version to install (e.g., `8.2`, `8.1`, `8.3`)

**Response:**
```json
{
  "message": "PHP installed successfully",
  "version": "8.2",
  "provider": "lsphp"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/providers/lsphp/install/8.2
curl -X POST http://localhost:8080/api/v1/providers/alt-php/install/8.2
curl -X POST http://localhost:8080/api/v1/providers/docker/install/8.2
```

**Error Response (400):**
```json
{
  "error": "Invalid provider: invalid-provider"
}
```

---

#### GET /api/v1/providers/{provider}/versions

List installed PHP versions for a specific provider.

**Parameters:**
- `provider` (path parameter) - Provider type: `remi`, `lsphp`, `alt-php`, or `docker`

**Response:**
```json
{
  "provider": "lsphp",
  "versions": ["8.2", "8.1", "8.0"]
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/providers/lsphp/versions
```

**Error Response (400):**
```json
{
  "error": "Invalid provider: invalid-provider"
}
```

---

#### GET /api/v1/providers/{provider}/available

List available PHP versions for a specific provider.

**Parameters:**
- `provider` (path parameter) - Provider type: `remi`, `lsphp`, `alt-php`, or `docker`

**Response:**
```json
{
  "provider": "lsphp",
  "versions": ["8.3", "8.2", "8.1", "8.0", "7.4"]
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/providers/lsphp/available
```

**Error Response (400):**
```json
{
  "error": "Invalid provider: invalid-provider"
}
```

---

## Response Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters or body
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error occurred

## Error Response Format

All error responses follow this format:
```json
{
  "error": "Error message description"
}
```

## Examples

### Complete Workflow

1. **List available providers:**
```bash
curl http://localhost:8080/api/v1/providers
```

2. **Check available PHP versions:**
```bash
curl http://localhost:8080/api/v1/php/available
# Or for specific provider:
curl http://localhost:8080/api/v1/providers/lsphp/available
```

3. **Install PHP 8.2 (default provider):**
```bash
curl -X POST http://localhost:8080/api/v1/php/install/8.2
```

4. **Install PHP 8.2 with specific provider:**
```bash
curl -X POST http://localhost:8080/api/v1/php/install/8.2?provider=lsphp
# Or using provider-specific endpoint:
curl -X POST http://localhost:8080/api/v1/providers/lsphp/install/8.2
```

5. **List installed PHP versions:**
```bash
curl http://localhost:8080/api/v1/php/versions
# Or for specific provider:
curl http://localhost:8080/api/v1/providers/lsphp/versions
```

4. **Create a pool for user "john":**
```bash
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "php_version": "8.2"}'
```

5. **Get pool information:**
```bash
curl http://localhost:8080/api/v1/pools/john
```

6. **List all pools:**
```bash
curl http://localhost:8080/api/v1/pools
```

7. **Delete pool:**
```bash
curl -X DELETE http://localhost:8080/api/v1/pools/john
```

## Notes

- All timestamps are in ISO 8601 format (UTC)
- Pool socket paths differ between RHEL and Debian systems
- PHP-FPM services are automatically reloaded after pool creation/deletion
- The user must exist in the system before creating a pool
- Pool configurations are stored in:
  - RHEL: `/etc/php-fpm.d/{username}.conf`
  - Debian: `/etc/php/{version}/fpm/pool.d/{username}.conf`

## Database

The application uses SQLite database located at:
```
/var/lib/lightweight-php/lightweight-php.db
```

All pools and PHP installations are tracked in the database for persistence and querying.
