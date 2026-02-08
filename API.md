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

**Response:**
```json
{
  "message": "PHP installed successfully",
  "version": "8.2"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/php/install/8.2
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

### Pool Management

#### GET /api/v1/pools

List all PHP-FPM pools.

**Response:**
```json
[
  {
    "User": "john",
    "PHPVersion": "8.2",
    "Status": "active",
    "ConfigPath": "/etc/php-fpm.d/john.conf",
    "SocketPath": "/var/run/php-fpm/john.sock"
  },
  {
    "User": "jane",
    "PHPVersion": "8.1",
    "Status": "active",
    "ConfigPath": "/etc/php/8.1/fpm/pool.d/jane.conf",
    "SocketPath": "/var/run/php/php8.1-jane.sock"
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

Create a new PHP-FPM pool for a user.

**Request Body:**
```json
{
  "username": "john",
  "php_version": "8.2"
}
```

**Fields:**
- `username` (required) - System username to create pool for
- `php_version` (optional) - PHP version to use (default: `8.2`)

**Response (201):**
```json
{
  "message": "Pool created successfully",
  "username": "john"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "php_version": "8.2"
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

1. **Install PHP 8.2:**
```bash
curl -X POST http://localhost:8080/api/v1/php/install/8.2
```

2. **List installed PHP versions:**
```bash
curl http://localhost:8080/api/v1/php/versions
```

3. **Create a pool for user "john":**
```bash
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "php_version": "8.2"}'
```

4. **Get pool information:**
```bash
curl http://localhost:8080/api/v1/pools/john
```

5. **List all pools:**
```bash
curl http://localhost:8080/api/v1/pools
```

6. **Delete pool:**
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
