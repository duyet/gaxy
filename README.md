# gaxy

Google Analytics / Google Tag Manager Proxy by Go

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

From Docker

```sh
docker run -it -p 3000:3000 ghcr.io/duyet/gaxy:latest
```

Registry: https://github.com/users/duyet/packages/container/package/gaxy
