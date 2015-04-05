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

	var nonceRead [24]byte
	_, err = rand.Read(nonceRead[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	var nonceWrite [24]byte
	_, err = rand.Read(nonceWrite[:])
	if err != nil {
		// todo: better error
		return 0, err
	}

	// var sharedKey [32]byte
	// box.Precompute(&sharedKey, pub2, priv1)

	var decrypted []byte
	// var encrypted_back [2048]byte
	var encrypted []byte
	message := []byte("hello, world!12345678901234567890123456789012345678901234567890")
	fmt.Printf("pointers %v %v %v %v\n", pub1, priv1, pub2, priv2)

	fmt.Printf("Message is %d bytes\n", len(message))
	encrypted = box.Seal(nil, message[:], &nonceRead, pub2, priv1)
	fmt.Printf("encrypted is %d bytes\n", len(encrypted))

	decrypted, _ = box.Open(nil, encrypted, &nonceRead, pub1, priv2)
	fmt.Printf("Decrypted is %d bytes: %s\n", len(decrypted), string(decrypted))

	return 0, nil
}

func main() {
	blah()
}
