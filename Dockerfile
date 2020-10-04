FROM golang:1.14 AS build
WORKDIR /go/src/github.com/duyet/gaxy
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o gaxy .

FROM alpine:latest
WORKDIR /app
ENV GOOGLE_ORIGIN="https://www.google-analytics.com" \
    INJECT_PARAMS_FROM_REQ_HEADERS=""
COPY --from=build /go/src/github.com/duyet/gaxy/gaxy .
EXPOSE 3000
CMD ["./gaxy"]