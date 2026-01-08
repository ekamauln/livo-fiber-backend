package main

import (
	"crypto/rand"
	"fmt"
	"log"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

func main() {
	// Generate exactly 32 random characters for PASETO v4 symmetric key
	key := make([]byte, 32)
	randomBytes := make([]byte, 32)

	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Failed to generate random key:", err)
	}

	// Convert random bytes to charset characters
	for i := 0; i < 32; i++ {
		key[i] = charset[int(randomBytes[i])%len(charset)]
	}

	keyString := string(key)

	fmt.Println("Generated PASETO V4 Symmetric Key (32 bytes):")
	fmt.Println("================================================")
	fmt.Println("\nAdd this to your .env file:")
	fmt.Printf("PASETO_SYMMETRIC_KEY=%s\n", keyString)
	fmt.Println("\nKey length:", len(keyString), "bytes")
	fmt.Println("\nNote: This is a raw 32-byte string, no encoding/decoding needed.")
}
