package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big" // Neu: Für die sichere Generierung von Zufallszahlen
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	saltLength = 16 // Length of the salt
	keyLength  = 32 // AES-256 key length

	// Konstanten für die Passwortgenerierung
	charsetNumbers = "0123456789"
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// DeriveKey uses scrypt to derive a key from the password and salt
func deriveKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, 1<<15, 8, 1, keyLength)
}

// Pad applies PKCS7 padding to data
func pad(data []byte) []byte {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// Unpad removes PKCS7 padding from data
func unpad(data []byte) ([]byte, error) {
	padding := data[len(data)-1]
	if int(padding) > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:len(data)-int(padding)], nil
}

// Encrypt encrypts plaintext using AES-256 with a password-derived key
func encrypt(plaintext []byte, password string) (string, string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", "", err
	}

	// Derive the key
	key, err := deriveKey(password, salt)
	if err != nil {
		return "", "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	// Pad the plaintext
	plaintext = pad(plaintext)

	// Create IV and ciphertext buffer
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", err
	}

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// Return salt and ciphertext as Base64-encoded strings
	return base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts Base64-encoded salt and ciphertext using AES-256 with a password-derived key
func decrypt(saltBase64, ciphertextBase64, password string) ([]byte, error) {
	// Decode the Base64-encoded salt
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding salt: %v", err)
	}

	// Decode the Base64-encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding ciphertext: %v", err)
	}

	// Derive the key
	key, err := deriveKey(password, salt)
	if err != nil {
		return nil, err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Check ciphertext length
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract IV from ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad the decrypted data and return it
	return unpad(ciphertext)
}

// generatePassword generates a cryptographically secure random password
func generatePassword(length int, includeNumbers, includeLower, includeUpper bool) (string, error) {
	var charset string
	if includeNumbers {
		charset += charsetNumbers
	}
	if includeLower {
		charset += charsetLower
	}
	if includeUpper {
		charset += charsetUpper
	}

	// Sicherstellen, dass mindestens ein Zeichensatz ausgewählt wurde
	if charset == "" {
		return "", fmt.Errorf("mindestens ein Zeichentyp (Zahlen, Kleinbuchstaben oder Großbuchstaben) muss enthalten sein")
	}

	// Sicherstellen, dass die Passwortlänge gültig ist
	if length <= 0 {
		return "", fmt.Errorf("Passwortlänge muss größer als 0 sein")
	}

	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	// crypto/rand verwenden, um Zeichen sicher auszuwählen
	for i := 0; i < length; i++ {
		// rand.Int gibt eine kryptographisch sichere Zufallszahl im Bereich [0, max) zurück.
		index, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = charset[index.Int64()]
	}

	return string(password), nil
}

func printUsage() {
	fmt.Println(`Verwendung:

Modi für Verschlüsselung/Entschlüsselung:
  --mode        Modus: 'crypt' für Verschlüsselung oder 'decrypt' für Entschlüsselung
  --text        Text zum Verschlüsseln/Entschlüsseln (Format 'salt:ciphertext' für Entschlüsselung)
  --password    Passwort für Verschlüsselung/Entschlüsselung

Modus für Passwortgenerierung:
  --createpass  Aktiviert den Modus zur Passwortgenerierung (ignoriert --mode und --text)
  --length      Erforderlich bei --createpass: Länge des zu generierenden Passworts (> 0)
  --numbers     Optional bei --createpass: true/false, ob Zahlen enthalten sein sollen (Standard: true)
  --lowerCases  Optional bei --createpass: true/false, ob Kleinbuchstaben enthalten sein sollen (Standard: true)
  --upperCases  Optional bei --createpass: true/false, ob Großbuchstaben enthalten sein sollen (Standard: true)

Beispiele:
  Verschlüsselung:        ./cryptdecrypt --mode crypt --text "mein Text" --password "meinPasswort"
  Entschlüsselung:        ./cryptdecrypt --mode decrypt --text "salt:ciphertext" --password "meinPasswort"
  Passwort-Generierung:   ./cryptdecrypt --createpass --length 12 --numbers=true --lowerCases=true`)
}

// splitSaltCiphertext splits input in the format 'salt:ciphertext'
func splitSaltCiphertext(input string) []string {
	parts := strings.SplitN(input, ":", 2) // Split into two parts
	if len(parts) != 2 {
		log.Fatalf("Text must be in 'salt:ciphertext' format.")
	}
	return parts
}

func main() {
	// Flags for Cryptography
	mode := flag.String("mode", "", "Mode: 'crypt' for encryption or 'decrypt' for decryption")
	text := flag.String("text", "", "Text to encrypt/decrypt (format 'salt:ciphertext' for decryption)")
	password := flag.String("password", "", "Password for encryption/decryption")

	// Flags for Password Generation (NEU)
	createPass := flag.Bool("createpass", false, "Enable password generation mode (ignores --mode and --text)")
	length := flag.Int("length", 0, "Required when --createpass: Length of the password to generate")
	includeNumbers := flag.Bool("numbers", true, "Optional when --createpass: Include numbers (default: true)")
	includeLower := flag.Bool("lowerCases", true, "Optional when --createpass: Include lowercase letters (default: true)")
	includeUpper := flag.Bool("upperCases", true, "Optional when --createpass: Include uppercase letters (default: true)")

	flag.Usage = printUsage
	flag.Parse()

	if *createPass {
		// --- Passwort-Generierungsmodus ---
		if *length <= 0 {
			fmt.Println("Error: Password length (--length) must be greater than 0 for password generation.")
			flag.Usage()
			os.Exit(1)
		}

		// Prüfe, ob mindestens eine Zeichenart ausgewählt wurde
		if !*includeNumbers && !*includeLower && !*includeUpper {
			fmt.Println("Error: At least one character type (--numbers, --lowerCases, or --upperCases) must be set to true.")
			flag.Usage()
			os.Exit(1)
		}

		genPassword, err := generatePassword(*length, *includeNumbers, *includeLower, *includeUpper)
		if err != nil {
			log.Fatalf("Password generation error: %v", err)
		}

		fmt.Printf("Generated Password: %s\n", genPassword)
		return // Beende das Programm nach der Generierung
	}

	// --- Kryptografie-Modus (Original-Logik) ---

	// Überprüfen, ob erforderliche Parameter für die Kryptografie-Modi bereitgestellt wurden
	if *mode == "" || *password == "" {
		fmt.Println("Error: All parameters (mode, password) are required for crypt/decrypt modes.")
		flag.Usage()
		os.Exit(1)
	}

	if *mode == "crypt" {
		// Sicherstellen, dass Text für die Verschlüsselung bereitgestellt wird
		if *text == "" {
			fmt.Println("Error: Text to encrypt is required.")
			flag.Usage()
			os.Exit(1)
		}

		// Verschlüsselung durchführen
		salt, ciphertext, err := encrypt([]byte(*text), *password)
		if err != nil {
			log.Fatalf("Encryption error: %v", err)
		}

		// Ausgabe von Salt und Chiffretext
		fmt.Printf("Salt: %s\n", salt)
		fmt.Printf("Ciphertext: %s\n", ciphertext)
	} else if *mode == "decrypt" {
		// Sicherstellen, dass Text im Format 'salt:ciphertext' bereitgestellt wird
		if *text == "" {
			fmt.Println("Error: Text in 'salt:ciphertext' format is required.")
			flag.Usage()
			os.Exit(1)
		}

		// Salt und Chiffretext aufteilen
		parts := splitSaltCiphertext(*text)
		// Hinweis: splitSaltCiphertext beendet das Programm bereits bei einem Fehler.

		// Entschlüsselung durchführen
		salt, ciphertext := parts[0], parts[1]
		decryptedText, err := decrypt(salt, ciphertext, *password)
		if err != nil {
			log.Fatalf("Decryption error: %v", err)
		}

		// Ausgabe des entschlüsselten Textes
		fmt.Print(string(decryptedText))
	} else {
		fmt.Println("Invalid mode. Use 'crypt' or 'decrypt'.")
		flag.Usage()
		os.Exit(1)
	}
}
