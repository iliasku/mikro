package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPost(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	site.redisdb().Do("FLUSHALL")
	var jsonStr = []byte(`{"url":"http://iliasku.tech"}`)
	req, _ := http.NewRequest("POST", "http://localhost:3000/url", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	site.Post(w, req)
	if w.Code != 201 {
		t.Error("expected code was 201 but we got ", w.Code)
	}
	if w.Body.String() != "{\"url\":\"http://iliasku.tech\"}, {\"short\":\"http://mikro.me/3\"}" {
		t.Error("produced value is not correct", w.Body.String())
	}

}

func TestPostError(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	site.redisdb().Do("FLUSHALL")
	var jsonStr = []byte(`{"url":"httttp://invalid"}`)
	req, _ := http.NewRequest("POST", "http://localhost:3000/url", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	site.Post(w, req)
	if w.Code != 422 {
		t.Error("expected code was 422 but we got ", w.Code)
	}

}

func TestSaveShort(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	site.redisdb().Do("FLUSHALL")
	u, err := site.saveShort("http://iliasku.tech")
	if u != "http://mikro.me/3" || err != nil {
		t.Error("produced value is not correct", u)
	}
	u, err = site.saveShort("http://iliasku.tech")
	if u != "http://mikro.me/3" || err != nil {
		t.Error("produced value is not the same", u)
	}
	u, err = site.saveShort("http://ilias ku.tech")
	if u != "" || err == nil {
		t.Error("wrong url did not cause error", u)
	}
}

func TestRedirect(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	site.redisdb().Do("FLUSHALL")
	site.saveShort("http://iliasku.tech")
	req, _ := http.NewRequest("GET", "http://localhost:3000/3", nil)
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Code != 302 {
		t.Error("expected code was 302 but we got ", w.Code)
	}
}

func TestRedirectNotFound(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	site.redisdb().Do("FLUSHALL")
	req, _ := http.NewRequest("GET", "http://localhost:3000/333", nil)
	w := httptest.NewRecorder()
	site.Redirect(w, req)
	if w.Code != 404 {
		t.Error("expected code was 404 but we got ", w.Code)
	}
}

func TestRedis(t *testing.T) {
	site := Site{Host: "http://mikro.me/", RedisURL: "redis://localhost:6379/0"}
	if site.redisURL() != "redis://localhost:6379/0" {
		t.Error("wrong REDISURL")
	}
	site = Site{RedisURL: ""}
	defer func() {
		recover()
	}()
	site.redisdb()

	t.Error("wrong REDISURL didn't cause any error")
}
