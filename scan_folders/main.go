package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	help := flag.Bool("h", false, "Hilfe anzeigen")
	rootPath := flag.String("path", ".", "Startverzeichnis")
	logFile := flag.String("log", "scan_output.txt", "Log-Datei")
	dryRun := flag.Bool("dry-run", true, "Nur anzeigen, nichts löschen")
	days := flag.Int("days", 0, "Nur Ordner berücksichtigen, die älter sind als X Tage (0 = ignorieren)")
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	*rootPath = filepath.Clean(*rootPath)
	*logFile = filepath.Clean(*logFile)

	// Root prüfen
	if _, err := os.Stat(*rootPath); os.IsNotExist(err) {
		fmt.Printf("Root-Verzeichnis existiert nicht: %s\n", *rootPath)
		return
	}

	// Logverzeichnis prüfen
	logDir := filepath.Dir(*logFile)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		fmt.Printf("Ordner für Log-Datei existiert nicht: %s\n", logDir)
		return
	}

	ignoreRegex := regexp.MustCompile(`_\d{8}_\d{4}$`)

	out, err := os.Create(*logFile)
	if err != nil {
		fmt.Println("Konnte Log-Datei nicht erstellen:", err)
		return
	}
	defer out.Close()

	fmt.Printf(
		"== Ordnerscan gestartet ==\nPfad: %s\nDry-Run: %v\nTage: %d\nLog: %s\n\n",
		*rootPath, *dryRun, *days, *logFile,
	)

	cutoff := time.Now().AddDate(0, 0, -*days)

	count := 0
	deleted := 0

	err = filepath.WalkDir(*rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(out, "Fehler bei %s: %v\n", path, err)
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if ignoreRegex.MatchString(d.Name()) {
			return filepath.SkipDir
		}

		if *days > 0 {
			info, err := d.Info()
			if err != nil {
				fmt.Fprintf(out, "Fehler bei Info %s: %v\n", path, err)
				return nil
			}

			if info.ModTime().After(cutoff) {
				return filepath.SkipDir
			}
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			fmt.Fprintf(out, "Fehler beim Lesen von %s: %v\n", path, err)
			return nil
		}

		hasTxt := false
		hasCerts := false

		for _, f := range entries {
			if f.IsDir() {
				continue
			}

			name := strings.ToLower(f.Name())

			switch {
			case strings.HasSuffix(name, ".txt"):
				hasTxt = true
			case strings.HasSuffix(name, ".cer"),
				strings.HasSuffix(name, ".pem"),
				strings.HasSuffix(name, ".p7b"),
				strings.HasSuffix(name, ".pfx"),
				strings.HasSuffix(name, ".crt"),
				strings.HasSuffix(name, ".der"):
				hasCerts = true
			}
		}

		if hasTxt && !hasCerts {
			count++
			fmt.Printf("[+] %s\n", path)
			fmt.Fprintf(out, "[MATCH] %s\n", path)

			if !*dryRun {
				if err := os.RemoveAll(path); err != nil {
					fmt.Fprintf(out, "Fehler beim Löschen von %s: %v\n", path, err)
				} else {
					fmt.Fprintf(out, "[DELETED] %s\n", path)
					deleted++
				}
			}

			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(out, "Scan-Fehler: %v\n", err)
		fmt.Println("Scan abgebrochen:", err)
		return
	}

	fmt.Printf(
		"\n== Scan abgeschlossen ==\nGefundene Ordner: %d\nGelöscht: %d\nDetails siehe: %s\n",
		count, deleted, *logFile,
	)
}

func printHelp() {
	fmt.Println("scan_folders – erkennt von OpenSSL erzeugte Ordner, die nur eine .txt-Datei enthalten und keine Zertifikatsdateien (.cer/.pem/.p7b/.pfx), \nund entfernt diese ungenutzten Verzeichnisse anhand der unten genannten Parameter.")
	fmt.Println()
	fmt.Println("Parameter:")
	fmt.Println("  -path       Startverzeichnis")
	fmt.Println("  -log        Log-Datei")
	fmt.Println("  -dry-run    Nur anzeigen, nichts löschen (true/false)")
	fmt.Println("  -days       Ordner nur berücksichtigen, wenn sie älter sind als X Tage")
	fmt.Println()
	fmt.Println("Beispiel:")
	fmt.Println("  scan_folders -path=C:\\Daten -days=30 -dry-run=false -log=result.txt")
}
