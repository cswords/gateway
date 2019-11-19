package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// NewProxyHandler TODO
func NewProxyHandler(config map[string]string) http.HandlerFunc {
	target := config["target"]
	targetURL, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	pathBase := config["pathBase"]

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {

		if !strings.HasPrefix(r.URL.RawPath, pathBase) {
			err := fmt.Errorf("Request URL %q does not match path base %q", r.URL.String(), pathBase)
			panic(err)
		}

		if r.Method != http.MethodOptions {
			r.Host = targetURL.Host

			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
			r.URL.RawPath = targetURL.RawPath + r.URL.RawPath[len(pathBase):len(r.URL.RawPath)]
			r.URL.RawQuery = strings.Join([]string{targetURL.RawQuery, r.URL.RawQuery}, "&")

			log.Println("Request url rewriten to %q", r.URL.String())

			// request will be copied
			reverseProxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
