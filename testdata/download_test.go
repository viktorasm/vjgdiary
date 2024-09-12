//go:build gentestdata
// +build gentestdata

package testdata

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vjgdienynas/collector"
)

func TestDownloadData(t *testing.T) {
	t.Skip("unfinished. need to think about storing returned cookies etc.")

	c := collector.NewCollector()
	c.WithTransport(&loggingTransport{http.DefaultTransport})

	r := require.New(t)
	_ = os.RemoveAll("data")
	r.NoError(os.MkdirAll("data", 0755))
	r.NoError(c.Login(os.Getenv("E2E_USER"), os.Getenv((os.Getenv("E2E_PASSWORD")))))
}

type loggingTransport struct {
	rt http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Call the original RoundTripper (which performs the request)
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset response body after reading

	if err := os.WriteFile(convertURLToFilename(req.URL), body, 0644); err != nil {
		return nil, fmt.Errorf("writing file: %w", err)
	}

	return resp, nil
}

var urlToFileName = regexp.MustCompile(`[^a-zA-Z0-9-_]`)

func convertURLToFilename(url *url.URL) string {
	return filepath.Join("data", strings.TrimSpace(urlToFileName.ReplaceAllString(url.String(), "_")))
}
