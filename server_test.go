package main

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	app := Setup()

	expectedBody := "pong"

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req, -1)

	assert.Equalf(t, false, err != nil, "Index route")
	assert.Equalf(t, 200, resp.StatusCode, "statusCode should be 200")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nilf(t, err, "err should be nil")
	assert.Equalf(t, string(body), expectedBody, "body should ok")
}

func TestGoogleAnalyticsJS(t *testing.T) {
	app := Setup()

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
