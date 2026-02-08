package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"lightweight-php/manager"

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

	// PHP installation endpoints
	r.HandleFunc("/api/v1/php/install/{version}", r.installPHP).Methods("POST")
	r.HandleFunc("/api/v1/php/versions", r.listPHPVersions).Methods("GET")
	r.HandleFunc("/api/v1/php/available", r.listAvailablePHP).Methods("GET")

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

	if err := r.poolManager.CreatePool(reqBody.Username, reqBody.PHPVersion); err != nil {
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

func (r *Router) installPHP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	version := vars["version"]

	if err := r.packageManager.InstallPHP(version); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"message": "PHP installed successfully",
		"version": version,
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

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}
