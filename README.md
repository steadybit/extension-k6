<img src="./logo.webp" height="130" align="right" alt="Kong logo">

# Steadybit extension-k6

A [Steadybit](https://www.steadybit.com/) action implementation to integrate k6 load tests into Steadybit experiments.

## Configuration

| Environment Variable                  | Meaning                                                                                                | Default |
|---------------------------------------|--------------------------------------------------------------------------------------------------------|---------|
| `STEADYBIT_EXTENSION_CLOUD_API_TOKEN` | K6 Cloud API Token. If provided, the extension will have the option to run load tests in the k6 cloud. |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Running the Extension

### Using Docker

```sh
$ docker run \
  --rm \
  -p 8087 \
  --name steadybit-extension-k6 \
  ghcr.io/steadybit/extension-k6:latest
```

### Using Helm in Kubernetes

```sh
$ helm repo add steadybit-extension-scaffold https://steadybit.github.io/extension-k6
$ helm repo update
$ helm upgrade steadybit-extension-k6 \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    steadybit-extension-k6/steadybit-extension-k6
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.
