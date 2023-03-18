package goners

import (
	"io"
	"net/http"
	"os"

	"github.com/cdfmlr/goners/wsforwarder"
	"golang.org/x/net/websocket"
)

// Outputer recv data from chan in & write them to somewhere.
type Outputer interface {
	Output(in chan []byte) // Output blocks.
}

// fileOutputer outputs to a file: one data one line
type fileOutputer struct {
	file io.WriteCloser
}

func NewFileOutputer(file string) (Outputer, error) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &fileOutputer{file: f}, nil
}

func (o fileOutputer) Output(in chan []byte) {
	for data := range in {
		o.file.Write(data)
		o.file.Write([]byte("\n"))
	}
	o.file.Close()
}

type webSocketOutputer struct {
	forwarder wsforwarder.Forwarder
}

func NewWebSocketOutputer(listenAddr string) (Outputer, error) {
	wso := &webSocketOutputer{
		forwarder: wsforwarder.NewMessageForwarder(),
	}

	handler := websocket.Handler(func(c *websocket.Conn) {
		wso.forwarder.ForwardMessageTo(c)
	})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", handler)
		http.ListenAndServe(listenAddr, mux)
	}()
	return wso, nil
}

func (o webSocketOutputer) Output(in chan []byte) {
	o.forwarder.ForwardMessageFrom(in)
}
