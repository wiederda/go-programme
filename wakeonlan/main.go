package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Funktion: MAC-Adresse validieren
func validateMACAddress(mac string) bool {
	regex := regexp.MustCompile(`^([0-9A-Fa-f]{2}([-:])?){5}[0-9A-Fa-f]{2}$`)
	return regex.MatchString(mac)
}

// Funktion: MAC-Adresse in Byte-Array konvertieren
func parseMACAddress(mac string) ([]byte, error) {
	// Entferne Delimiter wie ":" oder "-"
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")

	// Prüfe Länge der bereinigten MAC-Adresse
	if len(mac) != 12 {
		return nil, errors.New("ungültige MAC-Adresse")
	}

	// Hexadezimal-String in Byte-Array umwandeln
	macBytes := make([]byte, 6)
	for i := 0; i < 6; i++ {
		byteValue, err := strconv.ParseUint(mac[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, errors.New("fehler beim Parsen der MAC-Adresse")
		}
		macBytes[i] = byte(byteValue)
	}
	return macBytes, nil
}

// Funktion: Magic-Paket erstellen
func createMagicPacket(macBytes []byte) []byte {
	// Magic-Paket besteht aus 6x 0xFF und 16 Wiederholungen der MAC-Adresse
	magicPacket := make([]byte, 102)
	for i := 0; i < 6; i++ {
		magicPacket[i] = 0xFF
	}
	for i := 1; i <= 16; i++ {
		copy(magicPacket[i*6:(i+1)*6], macBytes)
	}
	return magicPacket
}

// Funktion: Wake-on-LAN-Paket senden
func sendWakeOnLan(macAddress, broadcastIP string) error {
	// MAC-Adresse in Bytes konvertieren
	macBytes, err := parseMACAddress(macAddress)
	if err != nil {
		return err
	}

	// Magic-Paket erstellen
	magicPacket := createMagicPacket(macBytes)

	// Verbindung über UDP erstellen
	conn, err := net.Dial("udp", broadcastIP+":9")
	if err != nil {
		return err
	}
	defer conn.Close()

	// Magic-Paket senden
	_, err = conn.Write(magicPacket)
	if err != nil {
		return err
	}

	fmt.Println("Wake-on-LAN-Paket gesendet!")
	return nil
}

func printHelp() {
	fmt.Println("Verwendung: wol <MAC-Adresse> [Broadcast-IP]")
	fmt.Println()
	fmt.Println("Argumente:")
	fmt.Println("  <MAC-Adresse>    Ziel-MAC-Adresse im Format XX:XX:XX:XX:XX:XX")
	fmt.Println("  [Broadcast-IP]   (Optional) Broadcast-IP-Adresse, Standard ist 255.255.255.255")
	fmt.Println()
	fmt.Println("Optionen:")
	fmt.Println("  -h, --help       Zeigt diese Hilfemeldung an")
	fmt.Println()
	fmt.Println("Beispiele:")
	fmt.Println("  wol 00:1A:2B:3C:4D:5E")
	fmt.Println("  wol 00-1A-2B-3C-4D-5E 192.168.0.255")
}

func main() {
	// Argumente prüfen
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		os.Exit(0)
	}

	// Argumente einlesen
	macAddress := os.Args[1]
	broadcastIP := "255.255.255.255" // Standard-Broadcast-IP
	if len(os.Args) > 2 {
		broadcastIP = os.Args[2]
	}

	// MAC-Adresse validieren
	if !validateMACAddress(macAddress) {
		fmt.Println("Fehler: Ungültige MAC-Adresse.")
		os.Exit(1)
	}

	// Wake-on-LAN-Paket senden
	err := sendWakeOnLan(macAddress, broadcastIP)
	if err != nil {
		fmt.Printf("Fehler: %v\n", err)
		os.Exit(1)
	}
}
