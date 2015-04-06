package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
	"log"
	"net"
	"os"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	reader := secureReader{
		backer:    r,
		decrypted: make([]byte, config.BufferSize)[:0],
	}
	box.Precompute(&reader.sharedKey, pub, priv)
	return reader
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	writer := secureWriter{
		backer: w,
	}
	box.Precompute(&writer.sharedKey, pub, priv)
	return writer
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return serverHandshake(conn)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	sc, err := serverHandshake(conn)
	if err != nil {
		return err
	}
	buf := make([]byte, config.BufferSize)
	r, err := sc.Read(buf)
	fmt.Printf("Jason: read server for %d bytes\n", r)
	if err != nil {
		return err
	}
	fmt.Println(string(buf[:r]))
	_, err = sc.Write(buf[:r])
	if err != nil {
		return err
	}
	return nil
}

var config = struct {
	BufferSize uint64
}{
	BufferSize: 1024 * 32, // 32kb
}

func main() {
	port := flag.Int("l", 0, "Listen mode. Specify port")
	flag.Parse()

	// Server mode
	if *port != 0 {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		log.Fatal(Serve(l))
	}

	// Client mode
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <port> <message>", os.Args[0])
	}
	conn, err := Dial("localhost:" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if _, err := conn.Write([]byte(os.Args[2])); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, len(os.Args[2]))
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf[:n])
}
