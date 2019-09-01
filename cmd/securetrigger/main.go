package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"
)

func main() {

	signature := os.Getenv("X-Hub-Signature")
	if len(signature) == 0 {
		log.Println("Err securing endpoint: no signature")
		os.Exit(1)
	}

	secret := os.Getenv("Secret")
	if len(secret) == 0 {
		log.Println("Err securing endpoint: no secret")
		os.Exit(1)
	}

	payload := os.Getenv("Payload")
	if len(payload) == 0 {
		log.Println("Err securing endpoint: no payload")
		os.Exit(1)
	}

	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature[5:]), []byte(expectedMAC)) {
		log.Println("Err securing endpoint: signature doesn't match")
		os.Exit(1)
	}

}
