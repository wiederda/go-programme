package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func printHelp() {
	fmt.Println("Extrahiert Text aus einem Bild mithilfe von Tesseract OCR.")
	fmt.Println("Argumente:")
	fmt.Println("  <Bildpfad>   Pfad zum Bild, aus dem der Text extrahiert werden soll.")
	fmt.Println("Optionen:")
	fmt.Println("  --help       Zeigt diese Hilfe an.")
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		printHelp()
		return
	}

	imagePath := os.Args[1]

	if _, err := os.Stat(imagePath); err != nil {
		log.Fatalf("Bilddatei nicht gefunden: %s", imagePath)
	}

	// Basisname ohne Erweiterung als Tesseract-Ausgabename verwenden
	base := filepath.Base(imagePath)
	name := base[:len(base)-len(filepath.Ext(base))]
	outFile := name + "_ocr"

	cmd := exec.Command("tesseract", imagePath, outFile)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Tesseract-Fehler: %v", err)
	}

	txtFile := outFile + ".txt"

	data, err := os.ReadFile(txtFile)
	if err != nil {
		log.Fatalf("Kann Ausgabedatei nicht lesen: %v", err)
	}

	fmt.Println("Erkannter Text:")
	fmt.Println(string(data))

	if err := os.Remove(txtFile); err != nil {
		log.Printf("Temporäre Datei konnte nicht gelöscht werden: %v", err)
	}
}
