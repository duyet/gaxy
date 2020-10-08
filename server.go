package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/valyala/fasthttp"
)

var config = LoadConfig()
var proxyClient = &fasthttp.Client{}

func main() {
	app := Setup()
	// Start server
	fmt.Printf("Listen on port %s", config.Port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", config.Port)))
}

// Setup Setup a fiber app with all of its routes
func Setup() *fiber.App {
	app := fiber.New()

	// CORS
	app.Use(cors.New())

	// Logger
	app.Use(logger.New())

	// Handler
	if config.RoutePrefix != "" {
		subRoute := app.Group(config.RoutePrefix)
		subRoute.Get("/ping", pingHandler)
		subRoute.All("/*", handleRequestAndRedirect)
	}
	app.Get("/ping", pingHandler)
	app.All("/*", handleRequestAndRedirect)

	return app
}

// Ping handler
func pingHandler(c *fiber.Ctx) error {
	return c.Send([]byte("pong"))
}

// Given a request send it to the appropriate url
func handleRequestAndRedirect(c *fiber.Ctx) error {
	upstreamReq := fasthttp.AcquireRequest()
	upstreamResp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(upstreamReq)
	defer fasthttp.ReleaseResponse(upstreamResp)

	c.Request().CopyTo(upstreamReq)

	// Overwrite
	url, _ := url.Parse(config.GoogleOrigin)
	upstreamReq.SetHost(url.Host)
	upstreamReq.URI().SetScheme(url.Scheme)
	// Trim prefix
	reqURI := string(c.Request().RequestURI())
	if config.RoutePrefix != "" && strings.HasPrefix(reqURI, config.RoutePrefix+"/") {
		reqURI := string(c.Request().RequestURI())
		reqURI = strings.TrimPrefix(reqURI, config.RoutePrefix)
		upstreamReq.SetRequestURI(reqURI)
	}

	// Prepare request
	prepareRequest(upstreamReq, c)
	fmt.Printf("GET %s -> making request to %s", c.Params("*"), upstreamReq.String())

	// Start request to dest URL
	if err := proxyClient.Do(upstreamReq, upstreamResp); err != nil {
		return err
	}

	// Post process the response
	if err := postprocessResponse(upstreamResp, c); err != nil {
		return err
	}

	return nil
}

// Prepare request
func prepareRequest(upstreamResp *fasthttp.Request, c *fiber.Ctx) {
	for _, name := range strings.Split(config.InjectParamsFromReqHeaders, ",") {
		// Convert header fields to request params
		// e.g. INJECT_PARAMS_FROM_REQ_HEADERS=uip,user-agent
		//   will be add this to the URI: ?uip=[VALUE]&user-agent=[VALUE]
		// To rename the key, use [HEADER_NAME]__[NEW_NAME]
		// e.g. INJECT_PARAMS_FROM_REQ_HEADERS=x-email__uip,user-agent__ua
		if name != "" {
			if strings.Contains(name, "__") {
				ss := strings.Split(name, "__")
				val := c.Get(ss[0])
				upstreamResp.URI().QueryArgs().Add(ss[1], val)
			} else {
				val := c.Get(name)
				upstreamResp.URI().QueryArgs().Add(name, val)
			}
		}
	}

	// Overwrite IP, UA
	upstreamResp.URI().QueryArgs().Add("uip", c.IP())
	upstreamResp.URI().QueryArgs().Add("ua", c.Get("User-Agent"))
}

// Post process response
func postprocessResponse(upstreamResp *fasthttp.Response, c *fiber.Ctx) error {
	// Inject
	upstreamResp.Header.Add("x-proxy-by", "gaxy")

	if strings.Contains(c.Params("*"), "ga.js") {
		contentEncoding := string(upstreamResp.Header.Peek("Content-Encoding"))
		var body []byte
		var err error
		switch contentEncoding {
		case "gzip":
			body, err = upstreamResp.BodyGunzip()
		case "br":
			body, err = upstreamResp.BodyUnbrotli()
		case "deflate":
			body, err = upstreamResp.BodyInflate()
		default:
			body = upstreamResp.Body()
		}
		if err != nil {
			return err
		}

		bodyString := string(body)
		url, err := url.Parse(c.BaseURL())
		if err != nil {
			return err
		}
		currentHost := url.Host
		find := []string{
			"ssl.google-analytics.com",
			"www.google-analytics.com",
			"google-analytics.com",
		}

		for _, toReplace := range find {
			bodyString = strings.ReplaceAll(bodyString, toReplace, currentHost)
		}

		c.Response().SetBodyString(bodyString)
		c.Response().Header.SetContentType(string(upstreamResp.Header.ContentType()))
	}

	return nil
}
