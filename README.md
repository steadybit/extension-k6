<img src="./logo.webp" height="130" align="right" alt="K6 logo">

# Steadybit extension-k6

A [Steadybit](https://www.steadybit.com/) action implementation to integrate k6 load tests into Steadybit experiments.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_k6).

## Configuration

| Environment Variable                  | Helm value             | Meaning                                                                                                | Reuired | Default |
|---------------------------------------|------------------------|--------------------------------------------------------------------------------------------------------|---------|---------|
| `STEADYBIT_EXTENSION_CLOUD_API_TOKEN` | `k6.cloudApiToken`     | K6 Cloud API Token. If provided, the extension will have the option to run load tests in the k6 cloud. | no      |         |
| `HTTPS_PROXY`                         | via extraEnv variables | Configure the proxy to be used for Datadog communication.                                              | no      |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Using Helm in Kubernetes

```sh
helm repo add steadybit-extension-k6 https://steadybit.github.io/extension-k6
helm repo update
helm upgrade steadybit-extension-k6 \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    steadybit-extension-k6/steadybit-extension-k6
```

### Using Docker

```sh
docker run \
  --rm \
  -p 8087 \
  --name steadybit-extension-k6 \
  ghcr.io/steadybit/extension-k6:latest
```

If you want to use K6 Cloud you need to provide an k6 cloud api token. You can add it for example with `--set k6.cloudApiToken="111-222-333"`

### Linux Package

Please use our [outpost-linux.sh script](https://docs.steadybit.com/install-and-configure/install-outpost-agent-preview/install-on-linux-hosts) to install the extension on your Linux machine.
The script will download the latest version of the extension and install it using the package manager.

After installing configure the extension by editing `/etc/steadybit/extension-k6` and then restart the service.

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.

## Proxy
To communicate to Datadog via a proxy, we need the environment variable `https_proxy` to be set.
This can be set via helm using the extraEnv variable

```bash
--set "extraEnv[0].name=HTTPS_PROXY" \
--set "extraEnv[0].value=https:\\user:pwd@CompanyProxy.com:8888"
```
