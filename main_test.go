package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prologic/bitcask"
	"github.com/stretchr/testify/assert"
)

var path = "db_test"

func prepareDB(path string) *bitcask.Bitcask {
	db, err := bitcask.Open(path)
	if err != nil {
		panic("unable to create/open DB")
	}

	return db
}

func cleanupDB(path string) {
	os.RemoveAll(path)
}

func TestNonExisting(t *testing.T) {
	db := prepareDB(path)
	defer cleanupDB(path)

	handler := func(w http.ResponseWriter, r *http.Request) {
		Bin(w, r, db)
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	val := req.URL.Query()
	val.Add("id", "nonexisting")
	req.URL.RawQuery = val.Encode()

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestExisting(t *testing.T) {
	db := prepareDB(path)
	defer cleanupDB(path)

	handler := func(w http.ResponseWriter, r *http.Request) {
		Bin(w, r, db)
	}

	// post
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// add id to form
	id, err := writer.CreateFormField("id")
	if err != nil {
		panic(err)
	}
	io.WriteString(id, "existing")

	// add file to form
	file, err := writer.CreateFormFile("file", "f")
	if err != nil {
		panic(err)
	}
	io.WriteString(file, "existing")

	writer.Close()

	req := httptest.NewRequest("POST", "http://example.com/", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// get
	w, req, body = nil, nil, nil
	req = httptest.NewRequest("GET", "http://example.com/", nil)
	val := req.URL.Query()
	val.Set("id", "existing")
	req.URL.RawQuery = val.Encode()

	w = httptest.NewRecorder()
	handler(w, req)

	resp = w.Result()
	body = &bytes.Buffer{}
	body.ReadFrom(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "existing", body.String())
}
