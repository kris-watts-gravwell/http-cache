package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestResponseWriter_Write(t *testing.T) {
	tests := []struct {
		name string
		data [][]byte
	}{
		{
			name: "empty",
			data: [][]byte{},
		},
		{
			name: "short",
			data: [][]byte{[]byte("foo")},
		},
		{
			name: "more",
			data: [][]byte{[]byte("foo bar baz")},
		},
		// there is no buffer currently so until then this will fail
		//{
		//	name: "multi",
		//	data: [][]byte{[]byte("foo bar baz"), []byte("foo bar bar baz")},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := func(w http.ResponseWriter, r *http.Request) {
				for _, b := range tt.data {
					w.Write(b)
				}
			}
			recorder := httptest.NewRecorder()
			rw := &responseWriter{ResponseWriter: recorder}
			next(rw, &http.Request{})

			if !bytes.Equal(recorder.Body.Bytes(), bytes.Join(tt.data, nil)) {
				t.Errorf("recorder body does not match, expected: '%s', got: '%s'", rw.body, tt.data)
			}
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}

	recorder := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: recorder}
	next(rw, &http.Request{})
	response := recorder.Result()
	if response.StatusCode != http.StatusAccepted {
		t.Errorf("response did not get expected status code, expected: 202, got: %d", recorder.Code)
	}
}

func TestResponseWriter_Header(t *testing.T) {
	tests := []struct {
		name   string
		next   http.HandlerFunc
		header http.Header
	}{
		{
			name: "add",
			next: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Foo", "bar")
				w.WriteHeader(http.StatusOK)
			},
			header: http.Header{"Foo": {"bar"}},
		},
		{
			name: "add-set",
			next: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("foo", "bar")
				w.Header().Add("foo", "baz")
				w.Header().Add("foo-bar", "foo")
				w.Header().Add("foo-bar", "bar")
				w.Header().Set("foo-bar", "baz")
				w.Header().Set("bar", "baz")
				w.Header().Del("bar")
				w.WriteHeader(http.StatusOK)
			},
			header: http.Header{"Foo": {"bar", "baz"}, "Foo-Bar": {"baz"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			rw := &responseWriter{ResponseWriter: recorder}
			tt.next(rw, &http.Request{})
			response := recorder.Result()

			if !reflect.DeepEqual(response.Header, tt.header) {
				t.Errorf("response header does not match, expected: '%v', got: '%v'", tt.header, response.Header)
			}
		})
	}
}

func TestResponseWriter_HeaderDel(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("bar")
		w.WriteHeader(http.StatusOK)
	}
	recorder := httptest.NewRecorder()
	// simulate other middleware setting values
	recorder.Header().Set("bar", "foo")
	recorder.Header().Set("untouched", "foo")
	rw := &responseWriter{ResponseWriter: recorder}
	next(rw, &http.Request{})

	response := recorder.Result()

	if response.Header.Get("bar") != "" {
		t.Errorf("response did not get empty header bar, got: %s", response.Header.Get("bar"))
	}

	if response.Header.Get("untouched") != "foo" {
		t.Errorf("response did not get expected header untouched, got: %s", response.Header.Get("untouched"))
	}
}
