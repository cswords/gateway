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
	targetURL.Path = "/" // Proxy only needs the origin
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		funcURL := rawurl
		if subPath, ok := vars["after_gateway_api_sub_path"]; ok {
			funcURL = funcURL + subPath
		}
		// Temporarily add this before fixing cors middleware
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		sw := ResponseWriter{ResponseWriter: w}
		if r.Method != http.MethodOptions {
			if onProxyAction != nil {
				onProxyAction(w, r)
			}
			rawQuery := r.URL.RawQuery
			r.Host = targetURL.Host
			r.URL, _ = url.Parse(funcURL)
			r.URL.RawQuery = rawQuery
			// request will be copied
			reverseProxy.ServeHTTP(sw, r)
		} else {
			sw.WriteHeader(http.StatusNoContent)
		}
	}
}

// ResponseWriter TODO
type ResponseWriter struct {
	http.ResponseWriter
}

// Header TODO
func (w ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write TODO
func (w ResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

// WriteHeader TODO
func (w ResponseWriter) WriteHeader(statusCode int) {
	deduplicate(w.Header(), "Access-Control-Allow-Origin")
	deduplicate(w.Header(), "Access-Control-Allow-Headers")
	w.ResponseWriter.WriteHeader(statusCode)
}

func deduplicate(h http.Header, key string) {
	if v := h[key]; len(v) > 1 {
		h[key] = []string{v[0]}
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
	r.HandleFunc(path, proxyHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodOptions)
}

type singleHeaderReponseWriter struct {
	http.ResponseWriter
	header http.Header
}

// Header TODO
func (w singleHeaderReponseWriter) Header() http.Header {
	if w.header == nil {

	}
	return w.header
}

// Write TODO
func (w singleHeaderReponseWriter) Write(b []byte) (i int, e error) {
	i, e = w.ResponseWriter.Write(b)
	return
}

// WriteHeader TODO
func (w singleHeaderReponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

type nonDuplicateHeader struct {
	http.Header
}

func (h nonDuplicateHeader) Set(key, value string) {
	if _, ok := h.Header[key]; !ok {
		h.Header.Set(key, value)
	}
}
