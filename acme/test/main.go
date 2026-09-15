package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starte RSA-4096 Generierung...")
	start := time.Now()
	_, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Println("FEHLER:", err)
		return
	}
	fmt.Println("Fertig nach:", time.Since(start))
}
