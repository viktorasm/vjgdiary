package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	s := BuildServer()

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.ServeHTTP(resp, req)
	r := require.New(t)

	r.Equal(http.StatusOK, resp.Code)

	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/login", nil)
	s.ServeHTTP(resp, req)

	r.Equal(http.StatusFound, resp.Code)
	r.Equal("/", resp.Header().Get("Location"))

	resp = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_app/env.js", nil)
	s.ServeHTTP(resp, req)
	r.Equal(http.StatusOK, resp.Code)

}
