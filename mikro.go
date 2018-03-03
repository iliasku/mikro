package mikro

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"crypto/md5"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/garyburd/redigo/redis"
	"github.com/asaskevich/govalidator"
	"github.com/prometheus/client_golang/prometheus"
)

type WrapHTTPHandler struct {
	handler http.Handler
}

type LoggedResponse struct {
	http.ResponseWriter
	status int
}

type Site struct {
	Host string
	RedisURL string
}

type shortRequest struct {
	URL string
}

var (
	httpResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "http_responses_total",
			Help:      "The count of http responses issued, classified by code and method.",
		},
		[]string{"code", "method"},
	)

	httpResponseLatencies = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "http_response_latencies",
			Help:      "Distribution of http response latencies (ms), classified by code and method.",
		},
		[]string{"code", "method"},
	)
)

func (loggedResponse *LoggedResponse) WriteHeader(status int) {
	loggedResponse.status = status
	loggedResponse.ResponseWriter.WriteHeader(status)
}

func (wrappedHandler *WrapHTTPHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	loggedWriter := &LoggedResponse{ResponseWriter: writer, status: 200}

	start := time.Now()
	wrappedHandler.handler.ServeHTTP(loggedWriter, request)
	elapsed := time.Since(start)
	msElapsed := elapsed / time.Millisecond

	status := strconv.Itoa(loggedWriter.status)
	httpResponsesTotal.WithLabelValues(status, request.Method).Inc()
	httpResponseLatencies.WithLabelValues(status, request.Method).Observe(float64(msElapsed))

	log.SetPrefix("[Info]")
	log.Printf("[%s] %s - %d, Method: %s, time elapsed was: %d(ms).\n",
		request.RemoteAddr, request.URL, loggedWriter.status, request.Method, msElapsed)
}

func (site Site) redisURL() string {
	if site.RedisURL != "" {
		return site.RedisURL
	}
	return ""
}

func (site Site) redisdb() redis.Conn {
	redisdb, err := redis.DialURL(site.redisURL())
	if err != nil {
		panic(err)
	}
	return redisdb
}

func (site Site) saveShort(url string) (shortest string, err error) {
	if !govalidator.IsURL(url) {
		return "", errors.New("invalid url")
	}

	redisdb := site.redisdb()
	defer redisdb.Close()

	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))

	similar, _ := redis.String(redisdb.Do("GET", "i:"+hash))
	if similar != "" {
		return site.Host + similar, nil
	}

	for hashShortestLen := 1; hashShortestLen <= 32; hashShortestLen++ {
		s, _ := redisdb.Do("GET", hash[0:hashShortestLen])
		if s == nil {
			shortest = hash[0:hashShortestLen]
			break
		}
	}
	if shortest == "" {
		return "", errors.New("url shortening failed")
	}

	redisdb.Do("SET", shortest, url)
	redisdb.Do("SET", "i:"+hash, shortest)
	return site.Host + shortest, nil
}

func (site Site) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	decoder := json.NewDecoder(r.Body)
	var t shortRequest
	decoder.Decode(&t)

	shortURL, err := site.saveShort(t.URL)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "{\"url\":%q}, {\"short\":%q}", t.URL, shortURL)
		return
	}
}

func (site Site) Redirect(w http.ResponseWriter, r *http.Request) {
	redisdb := site.redisdb()
	defer redisdb.Close()

	t, _ := redis.String(redisdb.Do("GET", r.URL.Path[1:]))
	u, _ := url.Parse(t)

	if u.String() == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusFound)
		fmt.Fprintf(w, "{\"redirect_url\":%q}", u.String())
		return
}

func init() {
	prometheus.MustRegister(httpResponsesTotal)
	prometheus.MustRegister(httpResponseLatencies)
}

func version(w http.ResponseWriter, r *http.Request) {
	version := "v0.1"
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"version\":%q}", version)
}

func main() {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	http.HandleFunc("/", site.Redirect)
	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/url", site.Post)
	http.HandleFunc("/version", version)
	log.Fatalln(http.ListenAndServe(":3000", &WrapHTTPHandler{http.DefaultServeMux}))
}