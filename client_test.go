package sse

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHandlerFunc(eventStrings []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, eventString := range eventStrings {
			if _, err := w.Write([]byte(eventString)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		}
	}
}

func testServer(eventStrings []string) *httptest.Server {
	mux := http.NewServeMux()
	handler := testHandlerFunc(eventStrings)
	mux.HandleFunc("/", handler)
	srv := httptest.NewServer(mux)
	return srv
}

func TestEventStreamScanner(t *testing.T) {
	b := strings.Builder{}
	b.WriteString("event: ping\ndata: ping1\n\n")
	b.WriteString("event: ping\ndata: ping2\n\n")
	b.WriteString("event: ping\ndata: ping3\n\n")
	buf := bytes.NewBufferString(b.String())
	scanner := NewEventStreamScanner(buf)
	expected := []string{
		"event: ping\ndata: ping1",
		"event: ping\ndata: ping2",
		"event: ping\ndata: ping3",
	}
	for _, e := range expected {
		if scanner.Scan() {
			assert.Equal(t, e, string(scanner.Bytes()))
		} else {
			assert.Fail(t, "failed to scan")
		}
	}
}

func TestSubscibe(t *testing.T) {
	srv := testServer([]string{
		"event: ping1\ndata: hello\n\n",
		"event: ping2\ndata: hello\n\n",
	})
	defer srv.Close()

	cli := NewClient(srv.URL)
	ch1 := make(chan *Event)
	cli.Subscribe("ping1", ch1)
	ch2 := make(chan *Event)
	cli.Subscribe("ping2", ch2)
	cli.Start()

	evt1 := <-ch1
	assert.Equal(t, "ping1", evt1.Name)

	evt2 := <-ch2
	assert.Equal(t, "ping2", evt2.Name)
}
