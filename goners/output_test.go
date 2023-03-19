package goners

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func TestFileOutputer_Output(t *testing.T) {
	tmpdir, err := os.MkdirTemp(".", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	type fields struct {
		file string
	}
	type args struct {
		in chan []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"fileOutputer",
			fields{file: path.Join(tmpdir, "out.txt")},
			args{in: make(chan []byte, 4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := NewFileOutputer(tt.fields.file)
			if err != nil {
				t.Fatal(err)
			}
			go o.Output(tt.args.in)

			t1 := time.Now()
			time.Sleep(time.Second)
			t2 := time.Now()

			tt.args.in <- []byte(fmt.Sprint(t1))
			tt.args.in <- []byte(fmt.Sprint(t2))

			time.Sleep(time.Second)

			readback, err := os.ReadFile(tt.fields.file)
			if err != nil {
				t.Fatal(err)
			}
			expected := fmt.Sprintf("%s\n%s\n", t1, t2)
			if string(readback) != expected {
				t.Errorf("❌ readback=%q ( expected %q)", readback, expected)
			}
		})
	}
}

func TestWebSocketOutput_Output(t *testing.T) {
	type fields struct {
		listenAddr string
	}
	type args struct {
		in chan []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"webSocketOutputer",
			fields{listenAddr: "localhost:9876"},
			args{in: make(chan []byte, 4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, handler := NewWebSocketOutputer()
			go func() {
				mux := http.NewServeMux()
				mux.Handle("/", handler)
				http.ListenAndServe(tt.fields.listenAddr, mux)
			}()

			go o.Output(tt.args.in)

			t1 := time.Now()
			time.Sleep(time.Second)
			t2 := time.Now()
			data := []string{fmt.Sprint(t1), fmt.Sprint(t2)}

			// ws 要同步发收

			client, err := websocket.Dial(fmt.Sprintf("ws://%s/", tt.fields.listenAddr), "", "http://localhost/")
			if err != nil {
				t.Fatal(err)
			}

			// send
			go func() {
				tt.args.in <- []byte(data[0])
				tt.args.in <- []byte(data[1])
			}()

			//recv
			for recvCount := 0; recvCount < len(data); recvCount++ {
				var recvMsg string
				if err := websocket.Message.Receive(client, &recvMsg); err != nil {
					t.Errorf("websocket.Message.Receive: %s", err.Error())
				}
				if recvMsg != data[recvCount] {
					t.Errorf("recvMsg != expected: %s != %s", recvMsg, data[recvCount])
				}
				t.Logf("recv: %s", recvMsg)
			}
		})
	}
}
