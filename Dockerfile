# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /prometheus-gmail-exporter-go

##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=build /prometheus-gmail-exporter-go /prometheus-gmail-exporter-go

EXPOSE 2112

USER nonroot:nonroot

ENTRYPOINT ["/prometheus-gmail-exporter-go"]
