# Test data

`archive.tar` is the [k6 archive](https://grafana.com/docs/k6/latest/misc/archive/) of `archive/tests/script.js`,
which imports `archive/lib/requests.js` via a relative path. That module records the counter `archive_module_calls`,
which the e2e test looks for in k6's metrics to tell that the relative import ran. Built inside the k6 image, so the
archive only records `/scripts/...` paths and no details of the machine it was built on. Use the `K6_VERSION` from the
`Dockerfile`:

```sh
docker run --rm -v "$PWD/archive:/scripts" -w /scripts grafana/k6:<version> archive tests/script.js -O /scripts/archive.tar
mv archive/archive.tar archive.tar
```

`scenarios/script.js` defines its own k6 scenario (`smoke`). The e2e tests run it as is, which works, and with the
step's Duration or Virtual users parameter set, which the extension refuses in prepare because those can't be
combined with k6 scenarios.
