package main

import (
	"crypto/rand"
	"errors"
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
	decrypted []byte
	sharedKey [32]byte
}

func (s *secureReader) readBacker(p []byte) (int, error) {
	// Read nonce
	var nonce [24]byte
	var err error
	_, err = io.ReadFull(s.backer, nonce[:])
	if err != nil {
		return 0, err
	}

	// Read message from underlying Reader
	message := make([]byte, config.BufferSize)
	c, err := s.backer.Read(message[:])
	if err != nil {
		return 0, err
	}

	// Decrypt new message part
	var decrypted []byte
	var success bool
	before := len(s.decrypted)
	decrypted, success = box.OpenAfterPrecomputation(s.decrypted, message[:c], &nonce, &s.sharedKey)
	after := len(decrypted)
	read := after - before
	if !success {
		// what does this mean?
		return 0, nil
	}
	s.decrypted = decrypted
	return read, nil
}

func (s secureReader) Read(p []byte) (int, error) {
	s.readBacker(p)

	read2 := copy(p, s.decrypted)     // TODO: need a writer or something to do this for me
	s.decrypted = s.decrypted[read2:] // TODO: maybe a circular buffer?

	return read2, nil
}

type secureWriter struct {
	backer    io.Writer
	sharedKey [32]byte
}

func (s secureWriter) Write(p []byte) (n int, err error) {
	// create random nonce
	var nonce [24]byte
	_, err = io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	encrypted := make([]byte, len(p)+1024)[:0]
	encrypted = box.SealAfterPrecomputation(encrypted, p, &nonce, &s.sharedKey)

	// write nonce
	_, err = s.backer.Write(nonce[:])
	if err != nil {
		return 0, err
	}

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
		Reader: NewSecureReader(conn, priv, &otherPub),
		Writer: NewSecureWriter(conn, priv, &otherPub),
	}, nil
}
