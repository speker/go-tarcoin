package limiter_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdata/influxdb/pkg/limiter"
)

func TestWriter_Limited(t *testing.T) {
	r := bytes.NewReader(bytes.Repeat([]byte{0}, 1024*1024))

	limit := 512 * 1024
	w := limiter.NewWriter(nopWriteCloser{ioutil.Discard}, limit, 10*1024*1024)

	start := time.Now()
	n, err := io.Copy(w, r)
	elapsed := time.Since(start)
	if err != nil {
		t.Error("copy error: ", err)
	}

	rate := float64(n) / elapsed.Seconds()
	if rate > float64(limit) {
		t.Errorf("rate limit mismath: exp %f, got %f", float64(limit), rate)
	}
}

func TestWriter_Limiter_ExceedBurst(t *testing.T) {
	limit := 10
	burstLimit := 20

	twentyOneBytes := make([]byte, 21)

	b := nopWriteCloser{bytes.NewBuffer(nil)}

	w := limiter.NewWriter(b, limit, burstLimit)
	n, err := w.Write(twentyOneBytes)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(twentyOneBytes) {
		t.Errorf("exected %d bytes written, but got %d", len(twentyOneBytes), n)
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (d nopWriteCloser) Close() error { return nil }
