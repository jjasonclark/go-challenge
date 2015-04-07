package main

import (
	"bytes"
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
	decrypted *bytes.Buffer
	sharedKey [32]byte
}

func (s secureReader) readBacker() ([]byte, error) {
	// Read nonce
	var nonce [24]byte
	if _, err := io.ReadFull(s.backer, nonce[:]); err != nil {
		return nil, err
	}

	// Read message from underlying Reader
	buffer := bytes.NewBuffer((make([]byte, config.BufferSize))[0:0])
	if _, err := buffer.ReadFrom(s.backer); err != nil {
		return nil, err
	}

	// Decrypt new message part
	decrypted, success := box.OpenAfterPrecomputation(nil, buffer.Bytes(), &nonce, &s.sharedKey)
	if !success {
		// what does this mean?
		return nil, nil
	}

	return decrypted, nil
}

func (s secureReader) Read(p []byte) (int, error) {
	for {
		i, err := s.decrypted.Write(p)
		decrypted, err := s.readBacker()

		c := copy(p, decrypted)
		s.decrypted.Write(decrypted[c:])
	}
	return s.decrypted.Read(p)
}

type secureWriter struct {
	backer    io.Writer
	sharedKey [32]byte
}

func (s secureWriter) Write(p []byte) (n int, err error) {
	// create random nonce
	var nonce [24]byte
	if _, err = io.ReadFull(rand.Reader, nonce[:]); err != nil {
		// todo: better error
		return 0, err
	}

	encrypted := box.SealAfterPrecomputation(nil, p, &nonce, &s.sharedKey)

	// write nonce
	if _, err = s.backer.Write(nonce[:]); err != nil {
		return 0, err
	}

	// write message
	return s.backer.Write(encrypted)
}

func serverHandshake(conn net.Conn) (*NaclReadWriteCloser, error) {
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
	_, err = io.ReadFull(conn, otherPub[:])
	if err != nil {
		return nil, NaclKeyExchangeError
	}

	// Return created reader and writer
	return &NaclReadWriteCloser{
		backer: conn,
		Reader: NewSecureReader(conn, priv, &otherPub),
		Writer: NewSecureWriter(conn, priv, &otherPub),
	}, nil
}
