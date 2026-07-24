package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func parseConfig(path string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Kommentare und leere Zeilen überspringen
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		// Zeile in Schlüssel und Wert trennen
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	// Pfad zur OpenSSL-Konfigurationsdatei
	opensslConfPath := "openssl.conf"

	// Konfigurationsdatei einlesen
	config, err := parseConfig(opensslConfPath)
	if err != nil {
		fmt.Printf("Fehler beim Einlesen der Konfigurationsdatei: %v\n", err)
		return
	}

	// Werte aus der Konfiguration abrufen oder Standardwerte setzen
	keySize, err := strconv.Atoi(config["default_bits"])
	if err != nil {
		keySize = 4096 // Standardwert, falls default_bits fehlt oder ungültig ist
	}
	fmt.Printf("Schlüssellänge: %d\n", keySize)

	// CSR-Parameter aus der Konfiguration extrahieren
	country := config["C"]
	state := config["ST"]
	locality := config["L"]
	organization := config["O"]
	organizationalUnit := config["OU"]
	commonName := config["CN"]
	email := config["emailAddress"]

	if commonName == "" {
		fmt.Println("Fehler: CN (Common Name) ist nicht in der Konfiguration angegeben.")
		return
	}

	// Dateinamen basierend auf dem CN festlegen
	privateKeyFileName := fmt.Sprintf("%s.key", commonName)
	csrFileName := fmt.Sprintf("%s.csr", commonName)

	// DNS-Namen und IP-Adressen einlesen
	dnsNames := []string{}
	ipAddresses := []net.IP{}
	for key, value := range config {
		if strings.HasPrefix(key, "DNS.") {
			dnsNames = append(dnsNames, value)
		} else if strings.HasPrefix(key, "IP.") {
			ip := net.ParseIP(value)
			if ip != nil {
				ipAddresses = append(ipAddresses, ip)
			}
		}
	}

	// Schritt 1: Privaten RSA-Schlüssel generieren
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		fmt.Printf("Fehler beim Generieren des privaten Schlüssels: %v\n", err)
		return
	}

	// Privaten Schlüssel im PEM-Format speichern
	privateKeyFile, err := os.Create(privateKeyFileName)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen der Schlüsseldatei: %v\n", err)
		return
	}
	defer privateKeyFile.Close()

	pem.Encode(privateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	fmt.Printf("Privater Schlüssel gespeichert in %s\n", privateKeyFileName)

	// Schritt 2: CSR-Template erstellen und dynamische Werte einsetzen
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{country},
			Province:           []string{state},
			Locality:           []string{locality},
			Organization:       []string{organization},
			OrganizationalUnit: []string{organizationalUnit},
			CommonName:         commonName,
		},
		EmailAddresses: []string{email},
		DNSNames:       dnsNames,
		IPAddresses:    ipAddresses,
	}

	// Schritt 3: CSR erstellen
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen des CSR: %v\n", err)
		return
	}

	// CSR im PEM-Format speichern
	csrFile, err := os.Create(csrFileName)
	if err != nil {
		fmt.Printf("Fehler beim Erstellen der CSR-Datei: %v\n", err)
		return
	}
	defer csrFile.Close()

	pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	fmt.Printf("CSR gespeichert in %s\n", csrFileName)
}
