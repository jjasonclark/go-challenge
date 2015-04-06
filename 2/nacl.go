package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

type NaclReadWriteCloser struct {
	Reader io.Reader
	Writer io.Writer
	backer io.ReadWriteCloser
}

var NaclKeyExchangeError = errors.New("Could not exhange keys")
var NaclEncryptionError = errors.New("Could not generate encryption keys")

func (s *NaclReadWriteCloser) Read(p []byte) (n int, err error) {
	return s.Reader.Read(p)
}

func (s *NaclReadWriteCloser) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

func (s *NaclReadWriteCloser) Close() error {
	return s.backer.Close()
}

type secureReader struct {
	backer    io.Reader
	sharedKey [32]byte
	decrypted []byte
}

func (s secureReader) readBacker(p []byte) (int, error) {
	// Read nonce
	var nonce [24]byte
	var err error
	_, err = io.ReadFull(s.backer, nonce[:])
	if err != nil {
		return 0, err
	}

	fmt.Println("Jason: Read nonce 24 bytes")

	// Read message from underlying Reader
	message := make([]byte, config.BufferSize)
	c, err := s.backer.Read(message[:])
	if err != nil {
		return 0, err
	}

	fmt.Printf("Jason: read message %d bytes\n", len(message))

	// Decrypt new message part
	var decrypted []byte
	var success bool
	before := len(s.decrypted)
	decrypted, success = box.OpenAfterPrecomputation(s.decrypted, message[:c], &nonce, &s.sharedKey)
	after := len(s.decrypted)
	read := after - before
	if !success {
		fmt.Printf("Jason: failed to read %d bytes\n", read)
		// what does this mean?
		return 0, nil
	}
	fmt.Printf("Jason: read %d bytes\n", read)
	s.decrypted = decrypted
	return read, nil
}

func (s secureReader) Read(p []byte) (int, error) {
	fmt.Printf("Jason: Trying to read %d bytes\n", len(p))
	wanted := len(p)
	read := copy(p, s.decrypted)
	fmt.Printf("Jason: already had %d bytes\n", read)
	s.decrypted = s.decrypted[read:]
	if read >= wanted {
		fmt.Println("Jason: read request via already decrypted")
		return read, nil
	}

	s.readBacker(p)

	read2 := copy(p, s.decrypted)     // TODO: need a writer or something to do this for me
	s.decrypted = s.decrypted[read2:] // TODO: maybe a circular buffer?

	fmt.Printf("Jason: read %d more bytes\n", read2)
	return read2, nil
}

type secureWriter struct {
	backer    io.Writer
	sharedKey [32]byte
}

func (s secureWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("Jason: Trying to write %d bytes\n", len(p))

	// create random nonce
	var nonce [24]byte
	_, err = io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	encrypted := make([]byte, len(p)+1024)[:0]
	encrypted = box.SealAfterPrecomputation(encrypted, p, &nonce, &s.sharedKey)

	fmt.Printf("Jason: writing nonce\n")
	// write nonce
	_, err = s.backer.Write(nonce[:])
	if err != nil {
		return 0, err
	}

	fmt.Printf("Jason: writing message %d bytes\n", len(encrypted))
	// write message
	return s.backer.Write(encrypted)
}

func serverHandshake(conn net.Conn) (io.ReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, NaclEncryptionError
	}

	// Send public key
	if _, err = conn.Write(pub[:]); err != nil {
		return nil, NaclKeyExchangeError
	}

	// Read othe side's public key
	var otherPub [32]byte
	var c int
	c, err = conn.Read(otherPub[:])
	if c < 32 || err != nil {
		return nil, NaclKeyExchangeError
	}

	// Return created reader and writer
	return &NaclReadWriteCloser{
		backer: conn,
		Reader: NewSecureReader(conn, priv, pub),
		Writer: NewSecureWriter(conn, priv, &otherPub),
	}, nil
}
