# Test data

`archive.tar` is the [k6 archive](https://grafana.com/docs/k6/latest/misc/archive/) of `archive/tests/script.js`,
which imports `archive/lib/requests.js` via a relative path. Built inside the k6 image, so the archive only records
`/scripts/...` paths and no details of the machine it was built on. Use the `K6_VERSION` from the `Dockerfile`:

```sh
docker run --rm -v "$PWD/archive:/scripts" -w /scripts grafana/k6:<version> archive tests/script.js -O /scripts/archive.tar
mv archive/archive.tar archive.tar
```
