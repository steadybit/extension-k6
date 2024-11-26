<img src="./logo.webp" height="130" align="right" alt="K6 logo">

# Steadybit extension-k6

A [Steadybit](https://www.steadybit.com/) action implementation to integrate k6 load tests into Steadybit experiments.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_k6).

## Configuration

| Environment Variable                  | Helm value             | Meaning                                                                                                | Reuired | Default |
|---------------------------------------|------------------------|--------------------------------------------------------------------------------------------------------|---------|---------|
| `STEADYBIT_EXTENSION_CLOUD_API_TOKEN` | `k6.cloudApiToken`     | K6 Cloud API Token. If provided, the extension will have the option to run load tests in the k6 cloud. | no      |         |
| `HTTPS_PROXY`                         | via extraEnv variables | Configure the proxy to be used for K6 Cloud communication.                                             | no      |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

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
    --set k6.cloudApiToken="111-222-333"
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
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-discovery) for more
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
1. The installed agent version needs to be >= X.XX, and - only for on-prem customers - the platform version needs to be >=X.X
2. Activate the location selection via environment or helm variable when deploying the latest extension version (see [configuration options](#configuration).
3. Configure every environment that should be able to run k6 load tests by including the execution location in the environment configuration.
	 One option is to add the statement `or target via the query language.type="com.steadybit.extension_k6.location"` to your existing query.
	 You can also filter the available execution locations down, e.g., via the clustername by using `(target.type="com.steadybit.extension_k6.location" and k8s.cluster-name="CLUSTER-NAME")`

