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
func NewProxyHandler(rawurl string, onProxyAction func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	targetURL, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	targetURL.Path = "/" // Proxy the host only here
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		funcURL := rawurl
		if subPath, ok := vars["after_gateway_api_sub_path"]; ok {
			funcURL = funcURL + subPath
		}
		if r.Method != http.MethodOptions {
			if onProxyAction != nil {
				onProxyAction(w, r)
			}
			q := r.URL.RawQuery
			// request will be copied
			r.Host = targetURL.Host
			r.URL, _ = url.Parse(funcURL) // TODO: process path wildcard
			r.URL.RawQuery = q
			reverseProxy.ServeHTTP(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
	}
}

// HandleProxyFunc let a router or subrouter reverse proxy a raw url to a path
func HandleProxyFunc(r *mux.Router, path string, rawurl string, onProxyAction func(http.ResponseWriter, *http.Request)) {
	// Process
	if strings.HasSuffix(path, "/*") {
		if !strings.HasSuffix(rawurl, "/") {
			err := fmt.Errorf("If the path has suffix /*, the function url must end with /")
			panic(err)
		}
		path = strings.TrimSuffix(path, "*") + "{after_gateway_api_sub_path:.+}"
	}
	proxyHandler := NewProxyHandler(rawurl, onProxyAction)
	r.HandleFunc(path, proxyHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost)
}
