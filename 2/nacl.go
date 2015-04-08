package main

import (
	"crypto/rand"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

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
	r     io.Reader
	key   *[32]byte
	nonce *[24]byte
}

func (r *SecureReader) readNonce() error {
	if r.nonce == nil {
		// Read nonce from underlying Reader
		var nonce [24]byte
		if _, err := io.ReadFull(r.r, nonce[:]); err != nil {
			return ErrNonceRead
		}
		r.nonce = &nonce
	}
	return nil
}

func (r *SecureReader) Read(p []byte) (int, error) {
	if err := r.readNonce(); err != nil {
		return 0, err
	}

	// Read message from underlying Reader
	buffer := make([]byte, len(p)+secretbox.Overhead)
	c, err := r.r.Read(buffer)
	if c <= 0 {
		return c, err
	}

	// Decrypt new message
	decrypted, success := box.OpenAfterPrecomputation(nil, buffer[:c], r.nonce, r.key)
	if !success {
		return 0, ErrDecryption
	}

	return copy(p, decrypted), err
}

type SecureWriter struct {
	w     io.Writer
	key   *[32]byte
	nonce *[24]byte
}

func (w *SecureWriter) writeNonce() error {
	if w.nonce == nil {
		// create random nonce and send
		var nonce [24]byte
		if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
			return ErrNonceWrite
		}
		if _, err := w.w.Write(nonce[:]); err != nil {
			return ErrNonceWrite
		}
		w.nonce = &nonce
	}
	return nil
}

func (w *SecureWriter) Write(p []byte) (int, error) {
	if err := w.writeNonce(); err != nil {
		return 0, err
	}

	// encrypted and send message
	encrypted := box.SealAfterPrecomputation(nil, p, w.nonce, w.key)
	if _, err := w.w.Write(encrypted); err != nil {
		return 0, err
	}

	return len(p), nil
}

func handshake(conn net.Conn) (*SecureReader, *SecureWriter, error) {
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
