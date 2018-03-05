# mikro
URL shortener with a [Redis](https://redis.io/) backend and [Prometheus](https://prometheus.io/) for metrics

## API specification


### POST /url

creates a new short url and returns status code depending on different states

sample Request: `curl -H 'Content-Type: application/json' -d '{"url":"http://iliasku.tech"}' http://localhost:3000/url`

sample Response: `{"url":"http://iliasku.tech"}, {"short":"http://mikro.me/3c"}`

returns HTTP status 201 on success

returns HTTP status 422 on errors (invalid url/parameters, shortening errors)
### GET /*url

Search in the storage for original url and redirects.

sample Request: `curl -H 'Content-Type: application/json' http://localhost:3000/3c`

sample Response: `{"redirect_url": "http://iliasku.tech"}`

returns HTTP status 302 and location to redirect if the url exists in redis

returns HTTP status 404 if there is no such url
### GET /version

returns current deployed version.

sample Response: `{"version": "v0.1"}`

returns HTTP status 200 if application is alive and ready to process requests
### GET /metrics

returns prometheus metrics (default golang metrics, http responses by code and method, http response latencies histograms)

returns HTTP status 200 if application is alive and ready to process requests



## deployment using helm

assuming you have a running kubernetes cluster with tiller:

`helm repo add helm-mikro https://iliasku.github.io/mikro`

`helm install helm-mikro/helm-mikro`
