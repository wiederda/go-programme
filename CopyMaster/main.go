package main // CopyMaster

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var sourceDir, destDir, extension string
	var excludeDirs string
	var help bool

	// Kommandozeilen-Flags definieren
	// Die Beschreibung wird angepasst, um die vereinfachte Eingabe zu erklären
	flag.StringVar(&sourceDir, "source", "", "Quellverzeichnis (Unter Windows können normale Pfade wie C:\\Pfad oder \\\\Server\\Pfad verwendet werden)")
	flag.StringVar(&destDir, "dest", "", "Zielverzeichnis")
	flag.StringVar(&extension, "ext", "", "Dateiendung (z.B. .txt)")
	flag.BoolVar(&help, "help", false, "Zeigt diese Hilfe an")
	flag.StringVar(&excludeDirs, "exclude", "", "Komma-separierte Liste von Ordnern, die ausgeschlossen werden sollen")
	flag.Parse()

	// Überprüfen der Argumente und des Hilfebanners
	if help || sourceDir == "" || destDir == "" || extension == "" {
		fmt.Println("Verwendung: ./CopyMaster --source=<source_dir> --dest=<destination_dir> --ext=<file_extension> [--exclude=<exclude1,exclude2,...>]")
		fmt.Println("Quellverzeichnis und Zielverzeichnis müssen gültige Pfade sein.")
		fmt.Println("Unter Windows ist nun die normale Pfadeingabe möglich, z.B. C:\\Daten\\Projekte oder \\\\Server\\Freigabe.")
		fmt.Println("Beipiel:")
		fmt.Println("./CopyMaster --source=\"C:\\Users\\Daten\" --dest=\"D:\\Backup\" --ext=\".log\"")
		os.Exit(1)
	}

	//  Anpassung der Windows/UNC-Pfade:
	// Alle einzelnen Backslashes (\) werden durch den entsprechenden Pfadtrenner des Betriebssystems ersetzt
	// oder, falls die Standard-Go-Bibliotheken es erfordern (wie bei UNC-Pfaden, die mit \\ beginnen),
	// werden sie durch doppelte Backslashes ersetzt, WENN das OS nicht Windows ist.
	// Das `filepath` Paket auf Windows kann jedoch die Slashes oft selbst korrigieren.

	// Der einfachste Weg, die manuelle Maskierung für den Nutzer zu vermeiden,
	// ist, alle Backslashes durch Forward Slashes zu ersetzen. Go's Standardbibliotheken
	// (insbesondere `filepath`) können Forward Slashes auf Windows intern in Backslashes umwandeln.

	sourceDir = strings.ReplaceAll(sourceDir, "\\", string(filepath.Separator)) // Ersetzt '\' durch '/'(Linux) oder '\'(Windows)
	destDir = strings.ReplaceAll(destDir, "\\", string(filepath.Separator))

	// Hinzufügen einer speziellen Behandlung für UNC-Pfade auf Windows,
	// die mit einem doppelten Backslash beginnen, um sicherzustellen, dass sie korrekt behandelt werden.
	// In Go sollte dies in der Regel durch die Verwendung von `filepath.FromSlash` oder die korrekte
	// Handhabung von `filepath.Separator` abgedeckt sein, aber wir stellen sicher,
	// dass der Anfang eines UNC-Pfades, der nun als `//` vorliegt, beibehalten wird.

	// Da `filepath.Walk` und `os.Open` auf Windows in der Regel auch mit '/' umgehen können,
	// wird die einfache Ersetzung auf `filepath.Separator` beibehalten.
	// Wir nutzen die Tatsache, dass Go in der Regel Pfade mit '/' auf allen OS akzeptiert.

	// Test, um sicherzustellen, dass UNC-Pfade korrekt initialisiert werden, falls sie nach der Ersetzung
	// mit einem einzelnen Slash beginnen würden:
	if os.Getenv("GOOS") == "windows" {
		// Bei UNC-Pfaden wie "\\server\share" wird der Pfad nach der Ersetzung zu "//server/share".
		// Dies ist in Go auf Windows gültig.

		// Korrigiere eventuelle doppelte Slashes am Anfang, falls das Ursprungspfad mit
		// maskiertem Backslash eingegeben wurde und nun "//" entsteht.
		sourceDir = strings.ReplaceAll(sourceDir, "//", string(filepath.Separator)+string(filepath.Separator))
		destDir = strings.ReplaceAll(destDir, "//", string(filepath.Separator)+string(filepath.Separator))

		// Dies ist der Schlüssel: Wir müssen nur sicherstellen, dass die Eingabe die "normalen"
		// Backslashes enthält, und Go die Pfade korrekt interpretiert.
		// Die Ersetzung durch `filepath.Separator` macht es cross-platform-kompatibel.
		sourceDir = filepath.Clean(sourceDir)
		destDir = filepath.Clean(destDir)
	}

	// Die ausgeschlossenen Ordner in eine Liste umwandeln
	excludeList := strings.Split(excludeDirs, ",")
	for i := range excludeList {
		excludeList[i] = strings.TrimSpace(excludeList[i]) // Entfernt unnötige Leerzeichen
	}

	// Alle Dateien im Quellverzeichnis durchsuchen
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Überprüfen, ob der aktuelle Pfad in einem der auszuschließenden Ordner ist
		for _, exclude := range excludeList {
			if exclude != "" && strings.Contains(path, exclude) {
				return nil // Ordner überspringen
			}
		}

		// Überprüfen, ob die Datei die richtige Erweiterung hat
		if !info.IsDir() && strings.HasSuffix(info.Name(), extension) {
			// Nur den Dateinamen extrahieren und in das Zielverzeichnis kopieren
			destPath := filepath.Join(destDir, info.Name())

			// Datei kopieren
			if err := copyFile(path, destPath); err != nil {
				return err
			}

			fmt.Println("Copied:", path, "->", destPath)
		}
		return nil
	})

	// Fehlerbehandlung nach dem Durchlaufen des Verzeichnisses
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Funktion zum Kopieren von Dateien (unverändert)
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Dateiinhalt kopieren
	_, err = io.Copy(destFile, sourceFile)
	return err
}
