package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"

	"github.com/liyue201/goqr"
	qrw "github.com/skip2/go-qrcode"
)

func writeQRCode(text, file string, size int) error {
	return qrw.WriteFile(text, qrw.Medium, size, file)
}

func readQRCode(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, qr := range qrCodes {
		results = append(results, string(qr.Payload))
	}

	return results, nil
}

func main() {
	mode := flag.String("mode", "write", "write|read")
	text := flag.String("text", "", "Text für QR-Code (nur write)")
	input := flag.String("in", "", "Eingabedatei (nur read)")
	output := flag.String("out", "qr.png", "Ausgabedatei (write)")
	size := flag.Int("size", 256, "QR-Bildgröße in Pixeln")

	flag.Parse()

	switch *mode {
	case "write":
		if *text == "" {
			fmt.Println("Fehler: -text fehlt")
			return
		}
		if err := writeQRCode(*text, *output, *size); err != nil {
			fmt.Println("Fehler beim Schreiben:", err)
			return
		}
		fmt.Printf("✔ QR-Code gespeichert: %s\n", *output)

	case "read":
		if *input == "" {
			fmt.Println("Fehler: -in fehlt")
			return
		}
		results, err := readQRCode(*input)
		if err != nil {
			fmt.Println("Fehler beim Lesen:", err)
			return
		}
		if len(results) == 0 {
			fmt.Println("Keine QR-Codes gefunden.")
			return
		}
		for i, r := range results {
			fmt.Printf("QR %d: %s\n", i+1, r)
		}

	default:
		fmt.Println("Unbekannter Modus:", *mode)
	}
}
