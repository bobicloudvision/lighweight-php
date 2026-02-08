package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"lightweight-php/manager"
	"lightweight-php/provider"

	"github.com/gorilla/mux"
)

type Router struct {
	*mux.Router
	poolManager    *manager.PoolManager
	packageManager *manager.PackageManager
}

func NewRouter() (*Router, error) {
	poolMgr, err := manager.NewPoolManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pool manager: %w", err)
	}

	pkgMgr, err := manager.NewPackageManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize package manager: %w", err)
	}

	r := &Router{
		Router:         mux.NewRouter(),
		poolManager:    poolMgr,
		packageManager: pkgMgr,
	}
	r.setupRoutes()
	return r, nil
}

func (r *Router) setupRoutes() {
	// Pool management endpoints
	r.HandleFunc("/api/v1/pools", r.listPools).Methods("GET")
	r.HandleFunc("/api/v1/pools", r.createPool).Methods("POST")
	r.HandleFunc("/api/v1/pools/{username}", r.getPool).Methods("GET")
	r.HandleFunc("/api/v1/pools/{username}", r.deletePool).Methods("DELETE")
	r.HandleFunc("/api/v1/pools/{username}/config", r.updatePoolConfig).Methods("PUT", "PATCH")

	// PHP installation endpoints
	r.HandleFunc("/api/v1/php/install/{version}", r.installPHP).Methods("POST")
	r.HandleFunc("/api/v1/php/versions", r.listPHPVersions).Methods("GET")
	r.HandleFunc("/api/v1/php/available", r.listAvailablePHP).Methods("GET")
	
	// Provider endpoints
	r.HandleFunc("/api/v1/providers", r.listProviders).Methods("GET")
	r.HandleFunc("/api/v1/providers/{provider}/install/{version}", r.installPHPWithProvider).Methods("POST")
	r.HandleFunc("/api/v1/providers/{provider}/versions", r.listPHPVersionsByProvider).Methods("GET")
	r.HandleFunc("/api/v1/providers/{provider}/available", r.listAvailablePHPByProvider).Methods("GET")

	// Health check
	r.HandleFunc("/health", r.healthCheck).Methods("GET")
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) listPools(w http.ResponseWriter, req *http.Request) {
	pools, err := r.poolManager.ListPools()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, pools)
}

func (r *Router) createPool(w http.ResponseWriter, req *http.Request) {
	var reqBody struct {
		Username   string `json:"username"`
		PHPVersion string `json:"php_version"`
		Provider   string `json:"provider"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if reqBody.Username == "" {
		jsonError(w, http.StatusBadRequest, "username is required")
		return
	}

	if reqBody.PHPVersion == "" {
		reqBody.PHPVersion = "8.2"
	}

	if reqBody.Provider == "" {
		reqBody.Provider = "remi"
	}

	if err := r.poolManager.CreatePool(reqBody.Username, reqBody.PHPVersion, reqBody.Provider); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{
		"message":  "Pool created successfully",
		"username": reqBody.Username,
	})
}

func (r *Router) getPool(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	pools, err := r.poolManager.ListPools()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, pool := range pools {
		if pool.User == username {
			jsonResponse(w, http.StatusOK, pool)
			return
		}
	}

	jsonError(w, http.StatusNotFound, "Pool not found")
}

func (r *Router) deletePool(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	if err := r.poolManager.DeletePool(username); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"message":  "Pool deleted successfully",
		"username": username,
	})
}

func (r *Router) updatePoolConfig(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	username := vars["username"]

	var settings map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&settings); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(settings) == 0 {
		jsonError(w, http.StatusBadRequest, "No settings provided")
		return
	}

	if err := r.poolManager.UpdatePoolConfig(username, settings); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":  "Pool configuration updated successfully",
		"username": username,
		"settings": settings,
	})
}

func (r *Router) installPHP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	version := vars["version"]
	
	// Check for provider parameter in query string
	providerParam := req.URL.Query().Get("provider")
	
	var err error
	if providerParam != "" {
		// Use specific provider
		providerType := provider.ProviderType(providerParam)
		err = r.packageManager.InstallPHPWithProvider(version, providerType)
	} else {
		// Use default provider
		err = r.packageManager.InstallPHP(version)
	}

	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	providerUsed := providerParam
	if providerUsed == "" {
		providerUsed = "remi" // default
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":  "PHP installed successfully",
		"version":  version,
		"provider": providerUsed,
	})
}

func (r *Router) listPHPVersions(w http.ResponseWriter, req *http.Request) {
	versions, err := r.packageManager.ListInstalledPHP()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Ensure versions is never nil
	if versions == nil {
		versions = []string{}
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"versions": versions,
	})
}

func (r *Router) listAvailablePHP(w http.ResponseWriter, req *http.Request) {
	versions, err := r.packageManager.ListAvailablePHP()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"versions": versions,
	})
}

func (r *Router) listProviders(w http.ResponseWriter, req *http.Request) {
	providers := []map[string]string{
		{
			"type":        "remi",
			"name":        "Remi Repository",
			"description": "Remi repository for RHEL, ondrej PPA for Debian",
			"status":      "active",
		},
		{
			"type":        "lsphp",
			"name":        "LiteSpeed PHP",
			"description": "LiteSpeed Web Server PHP",
			"status":      "active",
		},
		{
			"type":        "alt-php",
			"name":        "Alternative PHP",
			"description": "Alternative PHP packages",
			"status":      "stub",
		},
		{
			"type":        "docker",
			"name":        "Docker PHP",
			"description": "Docker-hosted PHP containers",
			"status":      "stub",
		},
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
	})
}

func (r *Router) installPHPWithProvider(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	version := vars["version"]
	providerTypeStr := vars["provider"]
	
	providerType := provider.ProviderType(providerTypeStr)
	if err := r.packageManager.InstallPHPWithProvider(version, providerType); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":  "PHP installed successfully",
		"version":  version,
		"provider": providerTypeStr,
	})
}

func (r *Router) listPHPVersionsByProvider(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	providerTypeStr := vars["provider"]
	
	providerType := provider.ProviderType(providerTypeStr)
	phpProvider, err := r.packageManager.GetProviderByType(providerType)
	if err != nil {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("Invalid provider: %s", providerTypeStr))
		return
	}

	versions, err := phpProvider.ListInstalledPHP()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	if versions == nil {
		versions = []string{}
	}
	
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"provider": providerTypeStr,
		"versions": versions,
	})
}

func (r *Router) listAvailablePHPByProvider(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	providerTypeStr := vars["provider"]
	
	providerType := provider.ProviderType(providerTypeStr)
	phpProvider, err := r.packageManager.GetProviderByType(providerType)
	if err != nil {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("Invalid provider: %s", providerTypeStr))
		return
	}

	versions, err := phpProvider.ListAvailablePHP()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"provider": providerTypeStr,
		"versions": versions,
	})
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}
