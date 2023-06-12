# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.20-alpine AS build

ARG NAME
ARG VERSION
ARG REVISION
ARG ADDITIONAL_BUILD_PARAMS

WORKDIR /app

RUN apk add build-base
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags="\
    -X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
    -o ./extension \
    ${ADDITIONAL_BUILD_PARAMS}

##
## Runtime
##
FROM alpine:3.17

ARG TARGETARCH=amd64
ADD https://github.com/grafana/k6/releases/download/v0.44.0/k6-v0.44.0-linux-$TARGETARCH.tar.gz /

RUN tar -xzf k6-v0.44.0-linux-$TARGETARCH.tar.gz && \
    rm k6-v0.44.0-linux-$TARGETARCH.tar.gz && \
    mv k6-v0.44.0-linux-$TARGETARCH/k6 /usr/local/bin/k6 && \
    rm -rf k6-v0.44.0-linux-$TARGETARCH

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /app/extension /extension

EXPOSE 8087
EXPOSE 8088

ENTRYPOINT ["/extension"]
