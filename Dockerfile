# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM goreleaser/goreleaser:v1.22.1 AS build

ARG TARGETOS TARGETARCH
ARG BUILD_WITH_COVERAGE
ARG BUILD_SNAPSHOT=true

WORKDIR /app

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH goreleaser build --snapshot="${BUILD_SNAPSHOT}" --single-target -o extension
##
## Runtime
##
FROM alpine:3.18

LABEL "steadybit.com.discovery-disabled"="true"

ARG TARGETARCH=amd64
ADD https://github.com/grafana/k6/releases/download/v0.44.0/k6-v0.44.0-linux-$TARGETARCH.tar.gz /

RUN tar -xzf k6-v0.44.0-linux-$TARGETARCH.tar.gz && \
    rm k6-v0.44.0-linux-$TARGETARCH.tar.gz && \
    mv k6-v0.44.0-linux-$TARGETARCH/k6 /usr/local/bin/k6 && \
    rm -rf k6-v0.44.0-linux-$TARGETARCH

RUN apk add zip

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /app/extension /extension

EXPOSE 8087 8088

ENTRYPOINT ["/extension"]
