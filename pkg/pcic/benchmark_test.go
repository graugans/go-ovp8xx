package pcic_test

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"errors"
	"io"
	"testing"

	"github.com/graugans/go-ovp8xx/v2/pkg/pcic"
)

type discardHandler struct{ chunks int }

func (d *discardHandler) Result(f pcic.Frame)                   { d.chunks += len(f.Chunks) }
func (d *discardHandler) Error(pcic.ErrorMessage)               {}
func (d *discardHandler) Notification(pcic.NotificationMessage) {}

// BenchmarkProcessIncoming measures the receive path with the recorded
// device data (26 frames with 6 chunks each, ~340 KiB per frame).
func BenchmarkProcessIncoming(b *testing.B) {
	file, err := tfs.Open("testdata/pcic-test-data.blob.bz2")
	if err != nil {
		b.Fatal(err)
	}
	raw, err := io.ReadAll(bzip2.NewReader(file))
	if err != nil {
		b.Fatal(err)
	}
	_ = file.Close()

	h := &discardHandler{}
	src := bytes.NewReader(raw)
	reader := bufio.NewReader(src)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(bufio.NewReadWriter(reader, nil)))
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src.Reset(raw)
		reader.Reset(src)
		for {
			err := p.ProcessIncomming(h)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
