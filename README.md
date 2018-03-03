# mikro
ulr shortener using redis as backend and prometheus for metrics

## API specification

### POST /url

creates a new short url and returns status code depending on different states

* sample Request: `{"url": "http://google.com"}`
* sample Response: `{"short": "http://localhost:3000/gIld"}`
* returns HTTP status 201 on success
* returns HTTP status 422 on errors (invalid url, shortening errors)

### GET /*url

Search in the storage for original url and redirects.

* sample Request: `{"url": "http://localhost:3000/gIld"}`
* sample Response: `{"redirect_url": "http://google.com"}`
* returns HTTP status 302 and location to redirect
* returns HTTP status 404 if there is no such url

### GET /version

returns current deployed version.

* sample Response: `{"version": "v0.1"}`
* returns HTTP status 200 if application is alive and ready to process requests

### GET /metrics

* returns prometheus metrics (default golang metrics, http responses by code and method, http response latencies histograms)
* returns HTTP status 200 if application is alive and ready to process requests
