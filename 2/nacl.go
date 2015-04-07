package main

import (
	"crypto/rand"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
)

var config = struct {
	BufferSize int
}{
	BufferSize: 1024 * 32, // 32kb
}

// Error exchanging the public keys
var ErrKeyExchange = errors.New("Could not exhange public keys")

// Error generating public and private key pairs
var ErrKeyGen = errors.New("Could not generate encryption keys")

// Error decrypting recieved message
var ErrDecryption = errors.New("Could not decrypt received message")

// Error sending nonce value for message
var ErrNonceWrite = errors.New("Could not send nonce value")

// Error reading nonce value for message
var ErrNonceRead = errors.New("Could not read nonce value")

type SecureReader struct {
	r   io.Reader
	key [32]byte
}

func (r SecureReader) Read(p []byte) (int, error) {
	// Read nonce from underlying Reader
	var nonce [24]byte
	if _, err := io.ReadFull(r.r, nonce[:]); err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, ErrNonceRead
	}

	// Read message from underlying Reader
	buffer := make([]byte, config.BufferSize)
	c, err := r.r.Read(buffer)
	if c <= 0 {
		return c, err
	}

	// Decrypt new message
	decrypted, success := box.OpenAfterPrecomputation(nil, buffer[:c], &nonce, &r.key)
	if !success {
		return 0, ErrDecryption
	}

	return copy(p, decrypted), err
}

type SecureWriter struct {
	w   io.Writer
	key [32]byte
}

func (w SecureWriter) Write(p []byte) (n int, err error) {
	// create random nonce and send
	var nonce [24]byte
	if _, err = io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return 0, ErrNonceWrite
	}
	if _, err = w.w.Write(nonce[:]); err != nil {
		return 0, ErrNonceWrite
	}

	// encrypted and send message
	encrypted := box.SealAfterPrecomputation(nil, p, &nonce, &w.key)
	if _, err := w.w.Write(encrypted); err != nil {
		return 0, err
	}

	return len(p), nil
}

func handshake(conn net.Conn) (io.Reader, io.Writer, error) {
	// Generate random key pair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, ErrKeyGen
	}

	// Send public key
	if _, err := conn.Write(pub[:]); err != nil {
		return nil, nil, ErrKeyExchange
	}

	// Read othe side's public key
	var otherPub [32]byte
	if _, err := io.ReadFull(conn, otherPub[:]); err != nil {
		return nil, nil, ErrKeyExchange
	}

	// Return created reader and writer
	return NewSecureReader(conn, priv, &otherPub),
		NewSecureWriter(conn, priv, &otherPub),
		nil
}
