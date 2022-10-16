package sse

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
)

type Client struct {
	url     string
	chs     map[string]chan *Event
	client  *http.Client
	closeCh chan struct{}
}

func NewClient(url string) *Client {
	return &Client{
		url:     url,
		chs:     make(map[string]chan *Event),
		client:  &http.Client{},
		closeCh: make(chan struct{}),
	}
}

func (cli *Client) Start() error {
	req, err := http.NewRequest("GET", cli.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	resp, err := cli.client.Get(cli.url)
	if err != nil {
		return err
	}

	scanner := NewEventStreamScanner(resp.Body)
	go func() {
	Loop:
		for {
			select {
			case <-cli.closeCh:
				resp.Body.Close()
				break Loop
			case <-req.Context().Done():
				resp.Body.Close()
				break Loop
			default:
				if scanner.Scan() {
					data := scanner.Bytes()
					e := NewEventFromString(string(data))
					if ch, ok := cli.chs[e.Name]; ok {
						ch <- e
					}
				}
				if err := scanner.Err(); err != nil {
					return
				}
			}
		}
	}()
	return nil
}

func (cli *Client) Subscribe(name string, ch chan *Event) {
	cli.chs[name] = ch
}

func (cli *Client) Close() {
	cli.closeCh <- struct{}{}
}

func NewEventStreamScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 1024)
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
			return i + 2, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
	scanner.Split(split)

	return scanner
}
