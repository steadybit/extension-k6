# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM goreleaser/goreleaser:v2.15.4 AS build

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_WITH_COVERAGE
ARG BUILD_SNAPSHOT=true
ARG SKIP_LICENSES_REPORT=false
ARG VERSION=unknown
ARG REVISION=unknown

WORKDIR /app

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOTOOLCHAIN=auto goreleaser build --snapshot="${BUILD_SNAPSHOT}" --single-target -o extension
##
## K6 with extensions
##
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS k6-builder

ARG TARGETOS
ARG TARGETARCH
ARG K6_VERSION=v1.7.1

RUN apk add --no-cache git
RUN go install go.k6.io/xk6/cmd/xk6@latest

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH xk6 build --k6-version $K6_VERSION \
  --with github.com/grafana/xk6-dns@latest \
  --with github.com/grafana/xk6-faker@latest \
  --with github.com/grafana/xk6-icmp@latest \
  --with github.com/grafana/xk6-mqtt@latest \
  --with github.com/grafana/xk6-redis@latest \
  --with github.com/grafana/xk6-sql@latest \
  --with github.com/grafana/xk6-sql-driver-mysql@latest \
  --with github.com/grafana/xk6-sql-driver-postgres@latest \
  --with github.com/grafana/xk6-ssh@latest \
  --with github.com/grafana/xk6-subcommand-explore@latest \
  --with github.com/grafana/xk6-disruptor@latest \
  --with github.com/grafana/xk6-kubernetes@latest \
  --with github.com/grafana/xk6-loki@latest \
  --with github.com/grafana/xk6-client-prometheus-remote@latest \
  --with github.com/grafana/xk6-sql-driver-azuresql@latest \
  --with github.com/grafana/xk6-sql-driver-clickhouse@latest \
  --with github.com/grafana/xk6-sql-driver-sqlserver@latest \
  --with github.com/grafana/xk6-tls@latest \
  --with github.com/grafana/xk6-client-tracing@latest \
  --with github.com/grafana/xk6-subcommand-httpbin@latest \
  --with github.com/mostafa/xk6-kafka@latest \
  --with github.com/tango-tango/xk6-msgpack@latest \
  --with github.com/phymbert/xk6-sse@latest \
  --output /k6
##
## Runtime
##
FROM alpine:3.23

ARG VERSION=unknown
ARG REVISION=unknown

LABEL "steadybit.com.discovery-disabled"="true"
LABEL "version"="${VERSION}"
LABEL "revision"="${REVISION}"
RUN echo "$VERSION" > /version.txt && echo "$REVISION" > /revision.txt

ARG K6_VERSION=v1.6.1

RUN echo "$K6_VERSION" > /k6-version.txt

COPY --from=k6-builder /k6 /usr/local/bin/k6

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN apk update && apk upgrade --no-cache && apk add --no-cache zip && rm -rf /var/cache/apk/* && \
    adduser -u $USER_UID -D $USERNAME

USER $USER_UID

WORKDIR /

COPY --from=build /app/extension /extension
COPY --from=build /app/licenses /licenses

EXPOSE 8087 8088

ENTRYPOINT ["/extension"]
