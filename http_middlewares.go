package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ------------------------------------------------------------------
// mwMultiMethod - produce handler for several http methods
func mwMultiMethod(in map[string]http.HandlerFunc) (http.HandlerFunc, error) {
	switch len(in) {
	case 0:
		return nil, fmt.Errorf("requires at least one handler")
	case 1:
		for method, handler := range in {
			return mwMethodOnly(handler, method), nil
		}
	}

	for method := range in {
		if method == "" {
			return nil, fmt.Errorf("mixing predetermined HTTP method with empty is not allowed")
		}
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		for method, handler := range in {
			if req.Method == method {
				handler.ServeHTTP(rw, req)
				return
			}
		}

		// not matched http method
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}, nil
}

// ------------------------------------------------------------------
// mwMethodOnly - allow one HTTP method only
func mwMethodOnly(handler http.HandlerFunc, method string) http.HandlerFunc {
	if method == "" {
		return handler
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == method {
			handler.ServeHTTP(rw, req)
		} else {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

// ------------------------------------------------------------------
// mwBasicAuth - add HTTP Basic Authentication
func mwBasicAuth(handler http.HandlerFunc, user, pass string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		reqUser, reqPass, ok := req.BasicAuth()
		if !ok || reqUser != user || reqPass != pass {
			rw.Header().Set("WWW-Authenticate", `Basic realm="Please enter user and password"`)
			http.Error(rw, "name/password is required", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(rw, req)
	}
}

// ------------------------------------------------------------------
// mwLogging - add logging for handler
func mwLogging(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		remoteAddr := req.RemoteAddr
		if realIP, ok := req.Header["X-Real-Ip"]; ok && len(realIP) > 0 {
			remoteAddr = realIP[0] + ", " + remoteAddr
		}
		start := time.Now()
		handler.ServeHTTP(rw, req)
		log.Printf("%s %s %s %s \"%s\" %s", req.Host, remoteAddr, req.Method, req.RequestURI, req.UserAgent(), time.Since(start).Round(time.Millisecond))
	}
}

// ------------------------------------------------------------------
// mwCommonHeaders - set common headers
func mwCommonHeaders(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Server", fmt.Sprintf("shell2http %s", VERSION))
		handler.ServeHTTP(rw, req)
	}
}