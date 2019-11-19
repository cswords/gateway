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

		if !strings.HasPrefix(r.URL.Path, pathBase) {
			err := fmt.Errorf("Request URL %q does not match path base %q", r.URL.String(), pathBase)
			panic(err)
		}

		if r.Method != http.MethodOptions {

			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
			r.URL.Path = targetURL.Path + r.URL.Path[len(pathBase):len(r.URL.Path)]
			r.URL.RawQuery = strings.Join([]string{targetURL.RawQuery, r.URL.RawQuery}, "&")

			r.Host = r.URL.Host
			r.RequestURI = r.URL.RequestURI()

			log.Println("Request url rewriten to %q", r.URL.String())

			// request will be copied
			reverseProxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
