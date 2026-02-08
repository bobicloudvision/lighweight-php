package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed pool.conf.tmpl
var defaultPoolTemplate string

// PoolConfigData holds the data for pool configuration template
type PoolConfigData struct {
	PoolName            string
	Username            string
	Group               string
	SocketPath          string
	ListenMode          string
	ProcessManager      string
	MaxChildren         int
	StartServers         int
	MinSpareServers      int
	MaxSpareServers      int
	MaxRequests          int
	ProcessIdleTimeout   string
	SendmailPath         string
	DisplayErrors        string
	ErrorLog             string
	LogErrors            string
	MemoryLimit          string
	MaxExecutionTime     string
	UploadMaxFilesize    string
	PostMaxSize          string
	DateTimezone         string
}

// DefaultPoolConfigData returns default values for pool configuration
func DefaultPoolConfigData(username, group, socketPath string) *PoolConfigData {
	return &PoolConfigData{
		PoolName:          username,
		Username:          username,
		Group:             group,
		SocketPath:        socketPath,
		ListenMode:        "0660",
		ProcessManager:    "dynamic",
		MaxChildren:       50,
		StartServers:      5,
		MinSpareServers:   5,
		MaxSpareServers:   35,
		MaxRequests:       500,
		ProcessIdleTimeout: "",
		SendmailPath:      "/usr/sbin/sendmail -t -i -f www@my.domain.com",
		DisplayErrors:      "off",
		ErrorLog:          fmt.Sprintf("/var/log/fpm-php.%s.log", username),
		LogErrors:         "on",
		MemoryLimit:       "128M",
		MaxExecutionTime:  "",
		UploadMaxFilesize: "",
		PostMaxSize:       "",
		DateTimezone:      "",
	}
}

// RenderPoolConfig renders the pool configuration template with the provided data
func RenderPoolConfig(templateContent string, data *PoolConfigData) (string, error) {
	tmpl, err := template.New("pool").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// LoadTemplate loads a template from embedded files
func LoadTemplate(name string) (string, error) {
	templates := map[string]string{
		"pool.conf.tmpl": defaultPoolTemplate,
	}

	content, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	return content, nil
}
