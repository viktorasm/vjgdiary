package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	r := require.New(t)
	s, err := BuildServer()
	r.NoError(err)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.ServeHTTP(resp, req)

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
