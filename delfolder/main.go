package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// normalizeInputPath nimmt einen vom User eingegebenen Windows/UNC-Pfad
// und wandelt ihn in eine saubere, plattformunabhängige Form mit / um.
func normalizeInputPath(p string) string {
	cleaned := filepath.Clean(p)
	return filepath.ToSlash(cleaned)
}

func main() {
	// Flags definieren
	pathFlag := flag.String("path", ".", "Basis-Pfad zum Durchsuchen (z.B. C:\\ordner oder \\\\server\\share)")
	daysFlag := flag.Int("days", 30, "Ordner älter als diese Anzahl an Tagen werden berücksichtigt")
	dryRun := flag.Bool("dry-run", true, "Nur anzeigen, nicht löschen")
	logFileFlag := flag.String("log", "", "Pfad zur Logdatei (optional)")

	// Eigene Hilfe-Ausgabe
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Ordner-Cleanup Tool\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Verwendung:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s [Optionen]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Optionen:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), `
Hinweise:
  - Es werden nur Ordner berücksichtigt, die ein "_" im Namen enthalten.
  - Ordner, die mit "_" BEGINNEN, werden ignoriert.
  - Pfade können wie im Explorer eingegeben werden (C:\... oder \\server\share).

Beispiele:
  # Testlauf, nur loggen auf Konsole
  %[1]s -path "C:\daten\ordner" -days 30 -dry-run=true

  # UNC-Pfad Testlauf
  %[1]s -path "\\server\share\ordner" -days 30 -dry-run=true

  # Wirklich löschen + Logdatei
  %[1]s -path "C:\daten\ordner" -days 30 -dry-run=false -log C:\temp\cleanup.log
`, os.Args[0])
	}

	flag.Parse()

	// Logging konfigurieren
	if *logFileFlag != "" {
		f, err := os.OpenFile(*logFileFlag, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Konnte Logdatei nicht öffnen: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Eingabe normalisieren
	rootPath := normalizeInputPath(*pathFlag)
	days := *daysFlag
	cutoff := time.Now().AddDate(0, 0, -days)

	log.Printf("Starte Suche in: %s | Älter als %d Tage | DryRun=%v\n", rootPath, days, *dryRun)

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() != filepath.Base(rootPath) {
			name := d.Name()

			// Ordner die mit "_" beginnen -> ignorieren
			if strings.HasPrefix(name, "_") {
				return nil
			}

			// Nur Ordner mit "_" irgendwo im Namen
			if strings.Contains(name, "_") {
				info, err := d.Info()
				if err != nil {
					return err
				}

				modTime := info.ModTime()
				if modTime.Before(cutoff) {
					if *dryRun {
						log.Printf("[TEST] Würde löschen: %s (letzte Änderung: %s)\n", normalizeInputPath(path), modTime.Format(time.RFC3339))
					} else {
						log.Printf("Lösche Ordner: %s (letzte Änderung: %s)\n", normalizeInputPath(path), modTime.Format(time.RFC3339))
						if err := os.RemoveAll(path); err != nil {
							log.Printf("Fehler beim Löschen von %s: %v\n", normalizeInputPath(path), err)
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Fehler beim Durchsuchen: %v\n", err)
	}
}
