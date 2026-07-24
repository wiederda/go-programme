package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

type Entry struct {
	Name       string `json:"name"`
	IV         string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
}

type Container struct {
	Version int     `json:"version"`
	Salt    string  `json:"salt"`
	Entries []Entry `json:"entries"`
}

func main() {
	var filePath, masterPass, cmd, entryName, entryValue string

	flag.Usage = printHelp // Überschreibt die Standard-Hilfe
	flag.StringVar(&filePath, "file", "vault.json", "Pfad zur Passwortdatei")
	flag.StringVar(&masterPass, "master", "", "Master-Passwort")
	flag.StringVar(&cmd, "cmd", "", "Befehl: add, get, del")
	flag.StringVar(&entryName, "name", "", "Name des Eintrags")
	flag.StringVar(&entryValue, "value", "", "Wert des Eintrags (nur add)")
	flag.Parse()

	if masterPass == "" || cmd == "" {
		printHelp()
		return
	}

	container, err := loadContainer(filePath)
	if err != nil {
		// Neue Datei erstellen, wenn nicht existiert
		container = &Container{
			Version: 1,
			Salt:    randomBase64(16),
			Entries: []Entry{},
		}
	}

	masterKey := deriveKey(masterPass, container.Salt)

	switch cmd {
	case "add":
		if entryName == "" || entryValue == "" {
			fmt.Println("Für add -name und -value angeben")
			return
		}
		addEntry(container, masterKey, entryName, entryValue)
	case "get":
		if entryName == "" {
			fmt.Println("Für get -name angeben")
			return
		}
		val, err := getEntry(container, masterKey, entryName)
		if err != nil {
			fmt.Println("Fehler:", err)
		} else {
			fmt.Println("Wert:", val)
		}
	case "del":
		if entryName == "" {
			fmt.Println("Für del -name angeben")
			return
		}
		deleteEntry(container, entryName)
	case "list": // <-- NEUER FALL
		listEntries(container)
	default:
		fmt.Println("Ungültiger Befehl")
		printHelp()
		return
	}

	// Datei sichern, außer bei get
	if cmd != "get" {
		err = saveContainer(filePath, container)
		if err != nil {
			fmt.Println("Fehler beim Speichern:", err)
		}
	}
}

// ----------------- Hilfsfunktionen -----------------

func printHelp() {
	fmt.Print(`Mini-Passwort-Keystore (Vault) CLI
Usage:
  -file string
        Pfad zur Passwortdatei (default "vault.json")
  -master string
        Master-Passwort
  -cmd string
        Befehl: add, get, del, list
  -name string
        Name des Eintrags
  -value string
        Wert des Eintrags (nur für add)

Beispiele:
  # Neuen Eintrag hinzufügen
  vault -file vault.json -master meinMaster -cmd add -name db -value geheimesPasswort

  # Eintrag abrufen
  vault -file vault.json -master meinMaster -cmd get -name db

  # Eintrag löschen
  vault -file vault.json -master meinMaster -cmd del -name db

  # Alle gespeicherten Namen auflisten
  vault -file vault.json -master meinMaster -cmd list
`)
}

func deriveKey(password, salt string) []byte {
	return pbkdf2.Key([]byte(password), []byte(salt), 100000, 32, sha256.New)
}

func randomBase64(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func encrypt(masterKey []byte, plaintext string) (string, string) {
	block, _ := aes.NewCipher(masterKey)
	gcm, _ := cipher.NewGCM(block)
	iv := make([]byte, gcm.NonceSize())
	_, _ = io.ReadFull(rand.Reader, iv)
	ct := gcm.Seal(nil, iv, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(iv), base64.StdEncoding.EncodeToString(ct)
}

func decrypt(masterKey []byte, ivB64, ctB64 string) (string, error) {
	iv, _ := base64.StdEncoding.DecodeString(ivB64)
	ct, _ := base64.StdEncoding.DecodeString(ctB64)
	block, _ := aes.NewCipher(masterKey)
	gcm, _ := cipher.NewGCM(block)
	pt, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func loadContainer(path string) (*Container, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Container
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func saveContainer(path string, c *Container) error {
	data, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(path, data, 0600)
}

func addEntry(c *Container, masterKey []byte, name, value string) {
	iv, ct := encrypt(masterKey, value)
	for i, e := range c.Entries {
		if e.Name == name {
			c.Entries[i].IV = iv
			c.Entries[i].Ciphertext = ct
			fmt.Println("Eintrag aktualisiert")
			return
		}
	}
	c.Entries = append(c.Entries, Entry{
		Name:       name,
		IV:         iv,
		Ciphertext: ct,
	})
	fmt.Println("Eintrag hinzugefügt")
}

func getEntry(c *Container, masterKey []byte, name string) (string, error) {
	for _, e := range c.Entries {
		if e.Name == name {
			return decrypt(masterKey, e.IV, e.Ciphertext)
		}
	}
	return "", errors.New("Eintrag nicht gefunden")
}

func deleteEntry(c *Container, name string) {
	newEntries := []Entry{}
	found := false
	for _, e := range c.Entries {
		if e.Name != name {
			newEntries = append(newEntries, e)
		} else {
			found = true
		}
	}
	if !found {
		fmt.Println("Eintrag nicht gefunden")
		return
	}
	c.Entries = newEntries
	fmt.Println("Eintrag gelöscht")
}

// listEntries gibt alle Namen der gespeicherten Einträge aus.
func listEntries(c *Container) {
	if len(c.Entries) == 0 {
		fmt.Println("Der Vault ist leer. Es sind keine Einträge gespeichert.")
		return
	}

	fmt.Println("--- Gespeicherte Einträge ---")
	for _, e := range c.Entries {
		fmt.Printf("- %s\n", e.Name)
	}
	fmt.Println("----------------------------")
}
