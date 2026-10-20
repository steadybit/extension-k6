<img src="./logo.webp" height="130" align="right" alt="K6 logo">

# Steadybit extension-k6

A [Steadybit](https://www.steadybit.com/) action implementation to integrate k6 load tests into Steadybit experiments.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_k6).

## Included k6 Extensions

The Docker image ships a custom k6 binary built with [xk6](https://github.com/grafana/xk6), including the following extensions.

### Official Grafana Extensions

| Extension                                                                        | Description                                    |
|----------------------------------------------------------------------------------|------------------------------------------------|
| [xk6-dns](https://github.com/grafana/xk6-dns)                                  | DNS resolution                                 |
| [xk6-faker](https://github.com/grafana/xk6-faker)                              | Random fake data generation                    |
| [xk6-icmp](https://github.com/grafana/xk6-icmp)                                | ICMP ping                                      |
| [xk6-mqtt](https://github.com/grafana/xk6-mqtt)                                | MQTT protocol                                  |
| [xk6-redis](https://github.com/grafana/xk6-redis)                              | Redis client                                   |
| [xk6-sql](https://github.com/grafana/xk6-sql)                                  | SQL database access                            |
| [xk6-sql-driver-mysql](https://github.com/grafana/xk6-sql-driver-mysql)        | MySQL driver for xk6-sql                       |
| [xk6-sql-driver-postgres](https://github.com/grafana/xk6-sql-driver-postgres)  | PostgreSQL driver for xk6-sql                  |
| [xk6-ssh](https://github.com/grafana/xk6-ssh)                                  | SSH commands                                   |
| [xk6-subcommand-explore](https://github.com/grafana/xk6-subcommand-explore)    | Explore k6 extensions for automatic resolution |

### Community Extensions

| Extension                                                                                    | Description                    |
|----------------------------------------------------------------------------------------------|--------------------------------|
| [xk6-disruptor](https://github.com/grafana/xk6-disruptor)                                  | Fault injection for k6         |
| [xk6-kafka](https://github.com/mostafa/xk6-kafka)                                          | Apache Kafka load testing      |
| [xk6-kubernetes](https://github.com/grafana/xk6-kubernetes)                                | Kubernetes API interaction     |
| [xk6-loki](https://github.com/grafana/xk6-loki)                                            | Loki log load testing          |
| [xk6-msgpack](https://github.com/Tango-Tango/xk6-msgpack)                                  | MessagePack serialization      |
| [xk6-client-prometheus-remote](https://github.com/grafana/xk6-client-prometheus-remote)    | Prometheus remote write client |
| [xk6-sql-driver-azuresql](https://github.com/grafana/xk6-sql-driver-azuresql)              | Azure SQL driver for xk6-sql   |
| [xk6-sql-driver-clickhouse](https://github.com/grafana/xk6-sql-driver-clickhouse)          | ClickHouse driver for xk6-sql  |
| [xk6-sql-driver-sqlserver](https://github.com/grafana/xk6-sql-driver-sqlserver)            | SQL Server driver for xk6-sql  |
| [xk6-sse](https://github.com/phymbert/xk6-sse)                                              | Server-Sent Events (SSE)       |
| [xk6-tls](https://github.com/grafana/xk6-tls)                                              | TLS metrics and configuration  |
| [xk6-client-tracing](https://github.com/grafana/xk6-client-tracing)                        | Distributed tracing client     |
| [xk6-subcommand-httpbin](https://github.com/grafana/xk6-subcommand-httpbin)                | Built-in httpbin test server   |

## Duration and Virtual Users

Both K6 actions have two optional parameters to override the script's options without touching the script:

- **Duration** – passed to k6 as `K6_DURATION` (e.g. `30s`). It replaces the script's `duration`, `stages` and
  `iterations`, and the experiment editor uses it to show how long the step runs.
- **Virtual users** – passed to k6 as `K6_VUS`.

Leave them empty to use the script's own options. They work the same for scripts and
archives, and take precedence over `K6_DURATION` / `K6_VUS` set in the step's environment variables.

They can't be combined with [k6 scenarios](https://grafana.com/docs/k6/latest/using-k6/scenarios/): a duration would
replace all scenarios with a single default one, and the virtual users alone would be ignored. If the script (or
archive) defines scenarios and either parameter is set, the step fails right at the start with an error, before any
load is generated. The extension checks this with `k6 inspect`, using the step's environment variables.

## Scripts with Multiple Files

The K6 actions accept either a single script (`.js`) or a [k6 archive](https://grafana.com/docs/k6/latest/misc/archive/)
(`.tar`). If your script imports other local modules (e.g. `import { login } from '../lib/auth.js'`) or reads files with
`open()`, bundle everything into an archive and upload that instead of the script:

```sh
k6 archive tests/main.js -O archive.tar
```

The archive contains the entry script, every local file it references, and the script's options, so k6 runs it without
further configuration. Environment variables configured in the step are still passed to k6 and take precedence over
values stored in the archive.

Keep in mind that the archive also stores the environment variables passed to `k6 archive` via `-e`/`--env` in plain
text, so pass secrets as environment variables in the step instead.

By default, the platform limits uploads to 5 MB per file. Scripts rarely come close, but data opened with `open()` (e.g. large
CSV or JSON datasets) is bundled into the archive too. If your archive gets too large, fetch such data at runtime
instead of bundling it.

## Environment of k6

k6 doesn't inherit the extension's environment. Scripts can read everything k6 gets via `__ENV`, so the extension only
passes:

- all `K6_*` variables, to configure k6 itself (e.g. outputs)
- `PATH`, `HOME`, `TMPDIR`, `TZ`
- `SSL_CERT_FILE`, `SSL_CERT_DIR`
- `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` (upper and lower case)
- `KUBERNETES_SERVICE_HOST`, `KUBERNETES_SERVICE_PORT`, `KUBECONFIG`, for the bundled xk6-kubernetes and xk6-disruptor
- the variables listed in `STEADYBIT_EXTENSION_K6_ENV_ALLOW_LIST` (Helm value `k6.envAllowList`)

The extension's own configuration and secrets (`STEADYBIT_EXTENSION_*`, e.g. `STEADYBIT_EXTENSION_CLOUD_API_TOKEN`) are
never passed. If the allow list contains such a variable, the extension ignores it and logs a warning at startup. The
allow list takes exact names only, no wildcards, so it's always clear which variables scripts can read.

Keep in mind that `K6_*` variables, including a `K6_CLOUD_TOKEN` set on the extension, are readable by scripts as well.
To pass values to a single step, use the step's *Environment variables* parameter instead.

```bash
--set "k6.envAllowList={MY_APP_URL,MY_APP_REGION}"
```

## Configuration

| Environment Variable                            | Helm value                | Meaning                                                                                                                                                                                              | Required | Default |
|-------------------------------------------------|---------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|---------|
| `STEADYBIT_EXTENSION_CLOUD_API_TOKEN`           | `k6.cloudApiToken`        | K6 Cloud API Token. If provided, the extension will have the option to run load tests in the k6 cloud.                                                                                               | no      |         |
| `STEADYBIT_EXTENSION_ENABLE_LOCATION_SELECTION` | `enableLocationSelection` | By default, the platform will select a random instance when executing actions from this extension. If you enable location selection, users can optionally specify the location via target selection. | no      | false   |
| `STEADYBIT_EXTENSION_K6_ENV_ALLOW_LIST`         | `k6.envAllowList`         | Comma-separated names of additional environment variables passed to k6, see [Environment of k6](#environment-of-k6). Exact names only, `STEADYBIT_EXTENSION_*` variables are ignored.                 | no      |         |
| `HTTPS_PROXY`                                   | via extraEnv variables    | Configure the proxy to be used for K6 Cloud communication.                                                                                                                                           | no      |         |

Beyond the settings above, this extension supports the configuration common to all Steadybit
extensions:

- [extension-kit](https://github.com/steadybit/extension-kit#environment-variables) — HTTP and
  health ports, TLS and mutual TLS, unix domain socket, logging, and pprof.
- [Target Filtering](https://github.com/steadybit/discovery-kit/blob/main/docs/target-filtering.md) —
  stop the extension reporting targets you do not want.
- [Group Matching](https://github.com/steadybit/discovery-kit/blob/main/docs/target-enrichment.md#group-matching) —
  tag discovered targets with a group, so enrichment rules only match within it.

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-k6.enabled=true \
```

If you want to use K6 Cloud you need to provide a k6 cloud api token. You can add it for example with `--set extension-k6.k6.cloudApiToken="111-222-333"`

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-k6/blob/main/charts/steadybit-extension-k6/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-k6/blob/main/charts/steadybit-extension-k6).

```bash
helm repo add steadybit-extension-k6 https://steadybit.github.io/extension-k6
helm repo update
helm upgrade steadybit-extension-k6 \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-agent \
    --set k6.cloudApiToken="111-222-333" \
    steadybit-extension-k6/steadybit-extension-k6
```

### Linux Package

Please use
our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-k6` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-registration) for more
information about extension registration and how to verify.

## Proxy
To communicate to K6 Cloud via a proxy, we need the environment variable `https_proxy` to be set.
This can be set via helm using the extraEnv variable

```bash
--set "extraEnv[0].name=HTTPS_PROXY" \
--set "extraEnv[0].value=https:\\user:pwd@CompanyProxy.com:8888"
```

## Location Selection
When multiple k6 extensions are deployed in different subsystems (e.g., multiple Kubernetes clusters), it can be tricky to ensure that the load test is performed from the right location when testing cluster-internal URLs or having different load testing hardware sizings.
To solve this, you can activate the location selection feature.
Once you do that, the k6 extension discovers itself as a k6 location.
When configuring the experiment, you can optionally define which extension's deployment should execute the loadtest.
Also, the execution locations are part of Steadybit's environment concept, so you can assign permissions for execution locations.

### Migration Guideline
Before activating the location selection feature, be sure to follow these steps:
1. The installed agent version needs to be >= 2.0.47, and - only for on-prem customers - the platform version needs to be >=2.2.2
2. Activate the location selection via environment or helm variable when deploying the latest extension version (see [configuration options](#configuration).
3. Configure every environment/scope that should be able to run k6 load tests by including the execution location in the environment/service scope.
   Simply add via query language `OR target.type ="com.steadybit.extension_k6.location"` or better, specify a Kubernetes cluster like `OR (target.type ="com.steadybit.extension_k6.location" AND k8s.cluster-name="<your-cluster-name>")` to filter the available execution locations.

## Version and Revision

The version and revision of the extension:
- are printed during the startup of the extension
- are added as a Docker label to the image
- are available via the `version.txt`/`revision.txt` files in the root of the image
