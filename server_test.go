package main

import (
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Equal(t, "pong", string(body), "body should be 'pong'")
}

func TestGAJS(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	req := httptest.NewRequest("GET", "/ga.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.NotEmpty(t, string(body), "body should not be empty")
	assert.Contains(t, string(body), "google", "body should contain 'google' keyword")
	assert.Equal(t, "text/javascript", resp.Header.Get("Content-Type"), "content type should be text/javascript")
}

func TestRoutePrefix(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"

	app := Setup(config)

	req1 := httptest.NewRequest("GET", "/ga.js", nil)
	req2 := httptest.NewRequest("GET", "/prefix/ga.js", nil)

	resp1, err1 := app.Test(req1, -1)
	assert.NoError(t, err1, "request without prefix should not fail")

	resp2, err2 := app.Test(req2, -1)
	assert.NoError(t, err2, "request with prefix should not fail")

	assert.Equal(t, 200, resp1.StatusCode, "status code should be 200")
	assert.Equal(t, 200, resp2.StatusCode, "status code should be 200")
}

func TestContentReplacement(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	req := httptest.NewRequest("GET", "/analytics.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "example.com", "body should contain replaced domain")
}

func TestContentReplacementWithCustomEnv(t *testing.T) {
	config := LoadConfig()
	config.GoogleOrigin = "https://www.googletagmanager.com"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/gtag.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "example.com", "body should contain replaced domain")
	assert.NotContains(t, string(body), "googletagmanager.com", "googletagmanager.com should be replaced")
}

func TestInjectHeader(t *testing.T) {
	config := LoadConfig()
	config.InjectParamsFromReqHeaders = "x-email__uip,user-agent__ua"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/collect", nil)
	req.Header.Add("X-Email", "me@duyet.net")
	req.Header.Add("user-agent", "Unitest")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err, "request should not fail")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.NotEmpty(t, string(body), "body should not be empty")
}

func TestContentReplacementWithPrefix(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/prefix/analytics.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "example.com/prefix", "body should contain replaced domain with prefix")
}

func TestBehindReverseProxy(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/prefix/analytics.js", nil)
	req.Header.Add("X-Forwarded-Host", "hihihi.com")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err, "request should not fail")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "hihihi.com/prefix", "body should contain forwarded host with prefix")
}
