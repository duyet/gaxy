# gaxy

Deployment: https://project-gaxy.appspot.com/ga.js

![Docker](https://github.com/duyet/gaxy/workflows/Docker/badge.svg)
![Go test](https://github.com/duyet/gaxy/workflows/Go/badge.svg)

Google Analytics / Google Tag Manager Proxy by Go.

![How it works?](.github/screenshot/how-gaxy-works.png)
<!-- https://sketchviz.com/@duyet/d4c36c277140a24111a723c439291303/9b91c5b780ff792c7dc08f70d22442a3ac523096 -->

## Development

Start server in local:

```sh
go run *.go
```

Build binary:

```sh
go build -o gaxy .
./gaxy
```

Testing:

```sh
go test
```

## Installation

### Using Docker

https://github.com/users/duyet/packages/container/package/gaxy

```sh
docker run -it -p 3000:3000 \
    -e ROUTE_PREFIX=/analytics \
    -e GOOGLE_ORIGIN=https://www.google-analytics.com \
    ghcr.io/duyet/gaxy:latest
```

### Using Helm

https://github.com/duyet/charts/tree/master/gaxy

```sh
helm repo add duyet https://duyet.github.io/charts
helm install google-analytics-proxy duyet/gaxy
```

### Using Google App Engine

https://cloud.google.com/appengine/docs/standard/go/quickstart

```sh
# 1. install gcloud
# 2. install app-engine-go component
gcloud components install app-engine-go
# 3. deploy
gcloud app deploy
```

### Environment variables

The following environment values are provided to customize Gaxy:

- `ROUTE_PREFIX`: Gaxy proxy prefix (e.g. `/analytics`). Default **""**
- `GOOGLE_ORIGIN`: Hostname to Google Analytics. Default **https://www.google-analytics.com**
- `INJECT_PARAMS_FROM_REQ_HEADERS`: Convert header fields (if gaxy is behind reverse proxy) to request parameters.
  - e.g. `INJECT_PARAMS_FROM_REQ_HEADERS=uip,user-agent` will be add this to the collector URI: `?uip=[VALUE]&user-agent=[VALUE]`
  - To rename the key, use `[HEADER_NAME]__[NEW_NAME]` e.g. `INJECT_PARAMS_FROM_REQ_HEADERS=x-email__uip,user-agent__ua`
  - List all the parameters of Google Analytics:

        - https://developers.google.com/analytics/devguides/collection/protocol/v1/parameters
        - https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference

- `PORT`: Gaxy webserver port. Default: **8080**

## Usage

```html
<!-- Google Analytics -->
<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-XXXXX-Y', 'auto');
ga('send', 'pageview');
</script>
<script async src='https://project-gaxy.appspot.com/analytics.js'></script>
<!-- End Google Analytics -->
```

## License

MIT
