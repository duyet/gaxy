package main

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	expectedBody := "pong"

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req, -1)

	assert.Equalf(t, false, err != nil, "Index route")
	assert.Equalf(t, 200, resp.StatusCode, "statusCode should be 200")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nilf(t, err, "err should be nil")
	assert.Equalf(t, string(body), expectedBody, "body should ok")
}

func TestGAJS(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	req := httptest.NewRequest("GET", "/ga.js", nil)
	resp, err := app.Test(req, -1)

	assert.Equalf(t, false, err != nil, "Index route")
	assert.Equalf(t, 200, resp.StatusCode, "statusCode should be 200")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nilf(t, err, "err should be nil")
	assert.NotEmpty(t, string(body), "body should not empty")
	assert.Contains(t, string(body), "google", "body should contains some keywords")
	assert.Equal(t, resp.Header.Get("Content-Type"), "text/javascript", "content-type should be text/javascript")
}

func TestRoutePrefix(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"

	app := Setup(config)

	req1 := httptest.NewRequest("GET", "/ga.js", nil)
	req2 := httptest.NewRequest("GET", "/prefix/ga.js", nil)

	resp1, err1 := app.Test(req1, -1)
	assert.Equalf(t, false, err1 != nil, "err should not be nil")

	resp2, err2 := app.Test(req2, -1)
	assert.Equalf(t, false, err2 != nil, "err should be nil")

	assert.Equalf(t, 200, resp1.StatusCode, "statusCode should be 200")
	assert.Equalf(t, 200, resp2.StatusCode, "statusCode should be 200")

	os.Setenv("ROUTE_PREFIX", "")
}

func TestContentReplacement(t *testing.T) {
	config := LoadConfig()
	app := Setup(config)

	req := httptest.NewRequest("GET", "/analytics.js", nil)

	resp, err := app.Test(req, -1)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	assert.Contains(t, string(body), "example.com")
}

func TestContentReplacementWithCustomEnv(t *testing.T) {
	config := LoadConfig()
	config.GoogleOrigin = "https://www.googletagmanager.com"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/gtag.js", nil)

	resp, err := app.Test(req, -1)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	assert.Contains(t, string(body), "example.com")

	// googletagmanager.com should be replaced by example.com
	assert.NotContains(t, string(body), "googletagmanager.com")
}

func TestInjectHeader(t *testing.T) {
	config := LoadConfig()
	config.InjectParamsFromReqHeaders = "x-email__uip,user-agent__ua"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/collect", nil)
	req.Header.Add("X-Email", "me@duyet.net")
	req.Header.Add("user-agent", "Unitest")

	resp, err := app.Test(req, -1)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	assert.NotEmpty(t, string(body))
}

func TestContentReplacementWithPrefix(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/prefix/analytics.js", nil)

	resp, err := app.Test(req, -1)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	assert.Contains(t, string(body), "example.com/prefix")
}

func TestBehindReverseProxy(t *testing.T) {
	config := LoadConfig()
	config.RoutePrefix = "/prefix"
	app := Setup(config)

	req := httptest.NewRequest("GET", "/prefix/analytics.js", nil)
	req.Header.Add("X-Forwarded-Host", "hihihi.com")

	resp, err := app.Test(req, -1)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Equalf(t, false, err != nil, "err should not be nil")

	assert.Contains(t, string(body), "hihihi.com/prefix")
}
