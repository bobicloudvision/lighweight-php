# FPM Pool Configuration Templates

This directory contains templates for PHP-FPM pool configurations.

## Template System

The template system uses Go's `text/template` package to generate PHP-FPM pool configuration files. Templates are embedded into the binary using Go's `embed` directive.

## Available Templates

- `pool.conf.tmpl` - Default PHP-FPM pool configuration template

## Template Variables

The following variables are available in the pool configuration template:

### Basic Settings
- `PoolName` - Name of the pool (usually the username)
- `Username` - System username
- `Group` - System group name
- `SocketPath` - Unix socket path for the pool
- `ListenMode` - Socket file permissions (default: "0660")

### Process Manager Settings
- `ProcessManager` - Process manager type (default: "dynamic")
- `MaxChildren` - Maximum number of child processes (default: 50)
- `StartServers` - Number of servers to start initially (default: 5)
- `MinSpareServers` - Minimum number of idle servers (default: 5)
- `MaxSpareServers` - Maximum number of idle servers (default: 35)
- `MaxRequests` - Maximum requests per child process (default: 500)
- `ProcessIdleTimeout` - Process idle timeout (optional)

### PHP Settings
- `SendmailPath` - Sendmail path (optional)
- `DisplayErrors` - Display errors flag (default: "off")
- `ErrorLog` - Error log file path
- `LogErrors` - Log errors flag (default: "on")
- `MemoryLimit` - PHP memory limit (default: "128M")
- `MaxExecutionTime` - Maximum execution time in seconds (optional)
- `UploadMaxFilesize` - Maximum upload file size (optional)
- `PostMaxSize` - Maximum POST data size (optional)
- `DateTimezone` - Default timezone (optional)

## Customizing Templates

To customize the pool configuration:

1. Edit `pool.conf.tmpl` in this directory
2. Rebuild the application: `go build -o lightweight-php .`
3. The template will be embedded into the binary

## Template Syntax

The templates use Go's `text/template` syntax:

- `{{.VariableName}}` - Insert variable value
- `{{- if .VariableName}}...{{- end}}` - Conditional block (only renders if variable is set)
- `{{- }}` - Trim whitespace

## Example

```ini
[{{.PoolName}}]
user = {{.Username}}
group = {{.Group}}
listen = {{.SocketPath}}
pm = {{.ProcessManager}}
pm.max_children = {{.MaxChildren}}
```

## Adding New Templates

To add a new template:

1. Create a new `.tmpl` file in this directory
2. Add it to the `LoadTemplate` function in `template.go`:
   ```go
   templates := map[string]string{
       "pool.conf.tmpl": defaultPoolTemplate,
       "custom.tmpl": customTemplate,
   }
   ```
3. Use `//go:embed custom.tmpl` to embed the new template
4. Rebuild the application
