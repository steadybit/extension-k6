# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM goreleaser/goreleaser:v2.4.1 AS build

ARG TARGETOS TARGETARCH
ARG BUILD_WITH_COVERAGE
ARG BUILD_SNAPSHOT=true
ARG SKIP_LICENSES_REPORT=false

WORKDIR /app

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH goreleaser build --snapshot="${BUILD_SNAPSHOT}" --single-target -o extension
##
## Runtime
##
FROM alpine:3.20

LABEL "steadybit.com.discovery-disabled"="true"
ARG K6_VERSION=v0.51.0
ARG TARGETARCH=amd64

ADD https://github.com/grafana/k6/releases/download/$K6_VERSION/k6-$K6_VERSION-linux-$TARGETARCH.tar.gz /

RUN tar -xzf k6-$K6_VERSION-linux-$TARGETARCH.tar.gz && \
    rm k6-$K6_VERSION-linux-$TARGETARCH.tar.gz && \
    mv k6-$K6_VERSION-linux-$TARGETARCH/k6 /usr/local/bin/k6 && \
    rm -rf k6-$K6_VERSION-linux-$TARGETARCH

RUN apk add zip

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /app/extension /extension

EXPOSE 8087 8088

ENTRYPOINT ["/extension"]
