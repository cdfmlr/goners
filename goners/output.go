package goners

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/cdfmlr/goners/wsforwarder"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

// Outputer recv data from chan in & write them to somewhere.
type Outputer interface {
	Output(in <-chan []byte) // Output blocks.
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

func (o fileOutputer) Output(in <-chan []byte) {
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

func (o webSocketOutputer) Output(in <-chan []byte) {
	o.forwarder.ForwardMessageFrom(in)
}

// PacketsFormater helps converting CaptureLivePackets.out into Output.in
type PacketsFormater interface {
	FormatPackets(in <-chan *Packet) <-chan []byte
}

type PacketsFormaterFunc func(in <-chan *Packet) <-chan []byte

func (f PacketsFormaterFunc) FormatPackets(in <-chan *Packet) <-chan []byte {
	return f(in)
}

// StringPacketsFormater formats recved input packets into strings,
// and send them to the returned output chan.
var StringPacketsFormater = PacketsFormaterFunc(func(in <-chan *Packet) <-chan []byte {
	out := make(chan []byte, ChanBufSize)
	go func() {
		for p := range in {
			out <- []byte(p.String())
		}
	}()
	return out
})

// StringPacketsFormater formats recved input packets into JSON bytes,
// and send them to the returned output chan.
var JsonPacketsFormater = PacketsFormaterFunc(func(in <-chan *Packet) <-chan []byte {
	out := make(chan []byte, ChanBufSize)
	go func() {
		for p := range in {
			j, err := json.Marshal(p)
			if err != nil {
				slog.Error("JsonPacketsFormater: marshal packet failed.", "err", err)
				continue
			}
			out <- j
		}
	}()
	return out
})
