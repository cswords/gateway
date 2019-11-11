package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

// NewProxyHandler create a new proxy handler
func NewProxyHandler(rawurl string) func(http.ResponseWriter, *http.Request) {
	targetURL, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	targetURL.Path = "/" // Proxy only needs the origin
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		funcURL := rawurl
		if subPath, ok := vars["after_gateway_api_sub_path"]; ok {
			funcURL = funcURL + subPath
		}
		if r.Method != http.MethodOptions {
			rawQuery := r.URL.RawQuery
			r.Host = targetURL.Host
			r.URL, _ = url.Parse(funcURL)
			r.URL.RawQuery = strings.Join([]string{r.URL.RawQuery, rawQuery}, "&")
			// request will be copied
			reverseProxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// HandleProxyFunc let a router or subrouter reverse proxy a raw url to a path
func HandleProxyFunc(r *mux.Router, path string, config map[string]string) {
	rawurl := config["url"]
	// Process
	if strings.HasSuffix(path, "/*") {
		if !strings.HasSuffix(rawurl, "/") {
			err := fmt.Errorf("If the path has suffix /*, the function url must end with /")
			panic(err)
		}
		path = strings.TrimSuffix(path, "*") + "{after_gateway_api_sub_path:.+}"
	}
	proxyHandler := NewProxyHandler(rawurl)
	r.HandleFunc(path, proxyHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodOptions)
}
