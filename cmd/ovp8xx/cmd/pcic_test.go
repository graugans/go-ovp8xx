package cmd

import (
	"net"
	"strconv"
	"testing"
)

func TestPCICConnectionRefused(t *testing.T) {
	// Reserve a free port and close it again, so the connection is refused
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()

	rootCmd.SetArgs([]string{"--ip", "127.0.0.1", "pcic", "--port", port})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected an error when the connection is refused")
	}
}
