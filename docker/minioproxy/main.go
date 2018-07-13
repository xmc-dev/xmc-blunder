package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	port    = ":9000"
	newPort = ":9001"
	minio   = "minio" + port
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
func newProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("Host", minio)
		req.Host = minio
		req.Header.Set("Referer", strings.Replace(req.Referer(), newPort, port, 1))
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {
	u, err := url.Parse("http://" + minio + "/")
	if err != nil {
		panic(err)
	}
	p := newProxy(u)
	http.Handle("/", p)
	log.Fatal(http.ListenAndServe("0.0.0.0"+newPort, nil))
}
