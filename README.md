# Steadybit extension-scaffold

TODO describe what your extension is doing here from a user perspective.

## Configuration

| Environment Variable                  | Meaning                                                                                                                                                                | Default                 |
|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|
| `STEADYBIT_EXTENSION_CLOUD_API_TOKEN` | Comma-separated list of discoverable robots                                                                                                                            | Bender,Terminator,R2-D2 |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Running the Extension

### Using Docker

```sh
$ docker run \
  --rm \
  -p 8080 \
  --name steadybit-extension-scaffold \
  ghcr.io/steadybit/extension-scaffold:latest
```

### Using Helm in Kubernetes

```sh
$ helm repo add steadybit-extension-scaffold https://steadybit.github.io/extension-scaffold
$ helm repo update
$ helm upgrade steadybit-extension-scaffold \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    steadybit-extension-scaffold/steadybit-extension-scaffold
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.
