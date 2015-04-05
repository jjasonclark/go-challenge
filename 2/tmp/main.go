package main

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/nacl/box"
)

func blah() (int, error) {
	var err error
	var pub1, pub2, priv1, priv2 *[32]byte
	pub1, priv1, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return 0, err
	}

	pub2, priv2, err = box.GenerateKey(rand.Reader)
	if err != nil {
		return 0, err
	}

	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	var readKey [32]byte
	var writeKey [32]byte
	box.Precompute(&writeKey, pub2, priv1)
	box.Precompute(&readKey, pub1, priv2)

	var encrypted_back, decrypted_back [2048]byte
	decrypted := decrypted_back[:0]
	encrypted := encrypted_back[:0]

	message := []byte("hello, world!1234567890123456789012345678901234567890123456789012345")

	fmt.Printf("Message is %d bytes: %s\n", len(message), string(message))
	encrypted = box.SealAfterPrecomputation(encrypted, message[:], &nonce, &writeKey)
	fmt.Printf("encrypted is %d bytes\n", len(encrypted))

	decrypted, _ = box.OpenAfterPrecomputation(decrypted, encrypted, &nonce, &readKey)
	fmt.Printf("Decrypted is %d bytes: %s\n", len(decrypted), string(decrypted))

	return 0, nil
}

func main() {
	blah()
}
