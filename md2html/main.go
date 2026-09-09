package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
)

// convertMarkdownToHTML konvertiert eine Markdown-Datei in HTML.
// Die Verzeichnisstruktur unterhalb des Quellverzeichnisses bleibt erhalten.
func convertMarkdownToHTML(mdPath, sourceDir, outputDir string) error {
	// Markdown-Datei lesen
	markdownData, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("Datei %q konnte nicht gelesen werden: %w", mdPath, err)
	}

	// Markdown in HTML konvertieren
	htmlData := markdown.ToHTML(markdownData, nil, nil)

	// Relativen Pfad zum Quellverzeichnis ermitteln
	relativePath, err := filepath.Rel(sourceDir, mdPath)
	if err != nil {
		return fmt.Errorf(
			"relativer Pfad für %q konnte nicht ermittelt werden: %w",
			mdPath,
			err,
		)
	}

	// Dateiendung .md durch .html ersetzen
	relativePath = strings.TrimSuffix(
		relativePath,
		filepath.Ext(relativePath),
	) + ".html"

	// Zielpfad zusammensetzen
	outputPath := filepath.Join(outputDir, relativePath)

	// Zielverzeichnis anlegen
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf(
			"Zielverzeichnis %q konnte nicht erstellt werden: %w",
			filepath.Dir(outputPath),
			err,
		)
	}

	// HTML-Datei schreiben
	if err := os.WriteFile(outputPath, htmlData, 0644); err != nil {
		return fmt.Errorf(
			"HTML-Datei %q konnte nicht geschrieben werden: %w",
			outputPath,
			err,
		)
	}

	fmt.Printf("  %s -> %s\n", mdPath, outputPath)

	return nil
}

// printHelp zeigt die Hilfe an.
func printHelp() {
	fmt.Println(`md2html - Markdown-Dateien in HTML konvertieren

Verwendung:
  md2html <quelle> <ziel>

Argumente:
  <quelle>    Quellverzeichnis mit Markdown-Dateien
  <ziel>      Zielverzeichnis für die erzeugten HTML-Dateien

Optionen:
  -h, --help  Diese Hilfe anzeigen

Beispiele:
  md2html ./content ./output

  md2html "/home/user/docs" "/var/www/docs"

  md2html "C:\Docs\content" "D:\Web\docs"

  md2html "\\server\share\docs" "\\server\share\html"

Hinweise:
  - Markdown-Dateien werden rekursiv verarbeitet.
  - Die Verzeichnisstruktur unterhalb des Quellverzeichnisses bleibt erhalten.
  - Dateien mit der Endung .md werden in .html umgewandelt.
  - Die Dateiendung .MD wird ebenfalls erkannt.
  - Windows-Laufwerkspfade werden unterstützt.
  - Windows UNC-/SMB-Pfade werden unterstützt.
  - Leerzeichen in Pfaden sind erlaubt; bei Bedarf den Pfad in Anführungszeichen setzen.`)
}

func main() {
	args := os.Args[1:]

	// Hilfe
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printHelp()
		return
	}

	// Anzahl der Argumente prüfen
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "Fehler: Quell- und Zielverzeichnis müssen angegeben werden.")
		fmt.Fprintln(os.Stderr, "Verwende 'md2html -h' für weitere Informationen.")
		os.Exit(1)
	}

	sourceDir := filepath.Clean(args[0])
	outputDir := filepath.Clean(args[1])

	// Quellverzeichnis prüfen
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		log.Fatalf(
			"Fehler: Quellverzeichnis %q konnte nicht geöffnet werden: %v",
			sourceDir,
			err,
		)
	}

	if !sourceInfo.IsDir() {
		log.Fatalf(
			"Fehler: Der Quellpfad %q ist kein Verzeichnis.",
			sourceDir,
		)
	}

	// Zielverzeichnis erstellen
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf(
			"Fehler: Zielverzeichnis %q konnte nicht erstellt werden: %v",
			outputDir,
			err,
		)
	}

	fmt.Printf("Quelle: %s\n", sourceDir)
	fmt.Printf("Ziel:   %s\n\n", outputDir)

	var processed int
	var failed int

	// Quellverzeichnis rekursiv durchsuchen
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			failed++

			log.Printf(
				"Fehler beim Zugriff auf %q: %v",
				path,
				walkErr,
			)

			return nil
		}

		// Verzeichnisse überspringen
		if info.IsDir() {
			return nil
		}

		// Nur Markdown-Dateien verarbeiten
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}

		// Datei konvertieren
		if err := convertMarkdownToHTML(path, sourceDir, outputDir); err != nil {
			failed++

			log.Printf(
				"Fehler beim Verarbeiten von %q: %v",
				path,
				err,
			)

			return nil
		}

		processed++

		return nil
	})

	if err != nil {
		log.Fatalf(
			"Fehler beim Durchsuchen des Quellverzeichnisses: %v",
			err,
		)
	}

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Printf("Verarbeitet: %d\n", processed)
	fmt.Printf("Fehler:      %d\n", failed)

	if failed > 0 {
		fmt.Println("Konvertierung mit Fehlern abgeschlossen.")
		os.Exit(1)
	}

	fmt.Println("Konvertierung erfolgreich abgeschlossen.")
}
