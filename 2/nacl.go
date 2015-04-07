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
var ErrNonceWrite = errors.New("Could not send nonce value")
var ErrNonceRead = errors.New("Could not read nonce value")
var ErrDecryption = errors.New("Could not decrypt received message")

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
}

func (s secureReader) Read(p []byte) (int, error) {
	// Read nonce from underlying Reader
	var nonce [24]byte
	if _, err := io.ReadFull(s.backer, nonce[:]); err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, ErrNonceRead
	}

	// Read message from underlying Reader
	buffer := make([]byte, config.BufferSize)
	c, err := s.backer.Read(buffer)
	if c <= 0 {
		return c, err
	}

	// Decrypt new message
	decrypted, success := box.OpenAfterPrecomputation(nil, buffer[:c], &nonce, &s.sharedKey)
	if !success {
		return 0, ErrDecryption
	}

	return copy(p, decrypted), err
}

type secureWriter struct {
	backer    io.Writer
	sharedKey [32]byte
}

func (s secureWriter) Write(p []byte) (n int, err error) {
	// create random nonce and send
	var nonce [24]byte
	if _, err = io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return 0, ErrNonceWrite
	}
	if _, err = s.backer.Write(nonce[:]); err != nil {
		return 0, ErrNonceWrite
	}

	// encrypted and send message
	encrypted := box.SealAfterPrecomputation(nil, p, &nonce, &s.sharedKey)
	if _, err := s.backer.Write(encrypted); err != nil {
		return 0, err
	}

	return len(p), nil
}

func handshake(conn net.Conn) (*NaclReadWriteCloser, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, NaclEncryptionError
	}

	// Send public key
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, NaclKeyExchangeError
	}

	// Read othe side's public key
	var otherPub [32]byte
	if _, err := io.ReadFull(conn, otherPub[:]); err != nil {
		return nil, NaclKeyExchangeError
	}

	// Return created reader and writer
	return &NaclReadWriteCloser{
		backer: conn,
		Reader: NewSecureReader(conn, priv, &otherPub),
		Writer: NewSecureWriter(conn, priv, &otherPub),
	}, nil
}
