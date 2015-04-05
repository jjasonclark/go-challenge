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
	sharedKey [32]byte
	decrypted []byte
}

func (s *secureReader) Read(p []byte) (n int, err error) {
	read := copy(p, s.decrypted)
	for ; read < len(p); read += copy(p[read:], s.decrypted) {
		// Read message from underlying Reader
		message := make([]byte, config.BufferSize)
		c, err := s.backer.Read(message[:])
		if err != nil {
			return read, err
		}

		// create random nonce
		var nonce [24]byte
		_, err = rand.Read(nonce[:])
		if err != nil {
			//what does this mean?
			return read, err
		}

		// Decrypt new message part
		var decrypted []byte
		var success bool
		decrypted, success = box.OpenAfterPrecomputation(s.decrypted, message[:c], &nonce, &s.sharedKey)
		if !success {
			// what does this mean?
			return read, nil
		}
		s.decrypted = decrypted
	}
	return read, nil
}

type secureWriter struct {
	backer    io.Writer
	encrypted []byte
	sharedKey [32]byte
}

func (s *secureWriter) Write(p []byte) (n int, err error) {
	// seal message
	// write to p all that I have
	// save remaining for later

	// create random nonce
	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	encrypted := box.SealAfterPrecomputation(s.encrypted, p, &nonce, &s.sharedKey)

	var wrote int
	for wrote < len(encrypted) {
		c, err := s.backer.Write(encrypted)
		wrote += c
		s.encrypted = s.encrypted[c:]
		if err != nil {
			return wrote, err
		}
	}
	s.encrypted = s.encrypted[0:0]
	return wrote, nil
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
