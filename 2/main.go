package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/nacl/box"
)

// NewSecureReader instantiates a new SecureReader
func NewSecureReader(r io.Reader, priv, pub *[32]byte) io.Reader {
	return r
}

// NewSecureWriter instantiates a new SecureWriter
func NewSecureWriter(w io.Writer, priv, pub *[32]byte) io.Writer {
	return w
}

// Dial generates a private/public key pair,
// connects to the server, perform the handshake
// and return a reader/writer.
func Dial(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", addr)
}

// Serve starts a secure echo server on the given listener.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go echoConnection(conn)
	}
	return nil
}

func serverHandshake(conn net.Conn) (net.Conn, error) {
	pk, _, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return conn, err
	}
	_, err = conn.Write(pk[:])
	return conn, err
}

var config = struct {
	BufferSize uint64
}{
	BufferSize: 1024 * 32, // 32kb
}

func echoConnection(conn net.Conn) {
	defer conn.Close()
	if _, err := serverHandshake(conn); err != nil {
		return
	}
	buf := make([]byte, config.BufferSize)
	for {
		r, err := conn.Read(buf)
		if err != nil {
			return
		}
		// os.Stdout.Write(buf)
		fmt.Println(string(buf))
		for w := 0; r > w; {
			c, err := conn.Write(buf[w:])
			if err != nil {
				break
			}
			w += c
		}
	}
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
