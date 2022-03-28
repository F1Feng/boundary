package worker

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type testConn struct {
	net.Conn
	t *testing.T
}

func (c testConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.t.Logf("Read %d bytes", n)
	return n, err
}

func (c testConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.t.Logf("Write %d bytes", n)
	return n, err
}

type testResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	t *testing.T
}

func (trw testResponseWriter) Write(b []byte) (int, error) {
	trw.t.Logf("Got %d bytes to write %q", len(b), b)
	return trw.ResponseWriter.Write(b)
}

func (trw testResponseWriter) WriteHeader(statusCode int) {
	trw.t.Logf("Writing header %d", statusCode)
	trw.ResponseWriter.WriteHeader(statusCode)
}

func (trw testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c, rw, err := trw.Hijacker.Hijack()
	if c != nil {
		c = testConn{Conn: c, t: trw.t}
	}
	return c, rw, err
}

func setupServer(t *testing.T, wg *sync.WaitGroup) string {
	t.Helper()

	wsHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, err := websocket.Accept(rw, req, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		require.NoError(t, err)
		wsjson.Write(context.Background(), conn, map[string]interface{}{"msg": "hello world"})
	})

	wrapingHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tw := testResponseWriter{ResponseWriter: rw, t: t}
		if rwh, ok := rw.(http.Hijacker); ok {
			tw.Hijacker = rwh
		}
		wsHandler.ServeHTTP(tw, req)
	})
	server := &http.Server{
		Handler:           wrapingHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
	}

	sl, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	go func() {
		defer func() {
			sl.Close()
			wg.Done()
		}()
		require.NoError(t, server.Serve(sl))
	}()
	return sl.Addr().String()
}

func TestWebSocketHijacking(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	addr := setupServer(t, wg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s", addr), nil)
	require.NoError(t, err)
	t.Logf("Got http response on client: %v", resp)
	defer func() {
		require.NoError(t, c.Close(websocket.StatusInternalError, "the sky is falling"))
	}()

	mt, b, err := c.Read(ctx)
	require.NoError(t, err)
	t.Logf("Received data %q of type %v", b, mt)

	require.NoError(t, wsjson.Write(ctx, c, "hi"))
	require.NoError(t, c.Close(websocket.StatusNormalClosure, ""))
	wg.Wait()
	t.Logf("Closed everything down cleanly.")
}
