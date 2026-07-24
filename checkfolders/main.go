package main

import (
	"flag"
	"fmt"
	"io/fs" // Import für fs.DirEntry
	"os"
	"path/filepath"
	"strings"
)

var (
	rootDir     string
	excludeDirs []string
	outputFile  string
)

// normalizePath bereinigt und normalisiert den übergebenen Pfad.
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	// filepath.Clean verarbeitet unter Windows auch UNC-Pfade korrekt.
	return filepath.Clean(p)
}

// showHelp zeigt die Hilfeinformationen an und beendet das Programm.
func showHelp() {
	// UNC-Pfad-Beispiel im Hilfetext für den Endbenutzer korrigiert (von \\\\ zu \\)
	helpText := `
checkfolders - Durchsucht Verzeichnisse rekursiv nach fehlenden .jpg und/oder .nfo-Dateien

Verwendung:
  checkfolders --root=<Startverzeichnis> [--exclude=dir1,dir2,...] [--out=Pfad\zur\Datei.txt] [--jpg] [--nfo]

Parameter:
  --root        Pfad zum Wurzelverzeichnis der Suche (z. B. C:\Hörspiele oder \\Server\Share)
  --exclude   (Optional) Kommagetrennte Liste von Ordnernamen, die ignoriert werden sollen
              Standard: #recycle,.ds_store,thumbs.db
  --out       (Optional) Pfad zur Ausgabedatei (Standard: missing_files.txt)
  --jpg         Nur Verzeichnisse mit fehlenden .jpg-Dateien anzeigen
  --nfo         Nur Verzeichnisse mit fehlenden .nfo-Dateien anzeigen

Beispiele:
  checkfolders --root="C:\Share"
  checkfolders --root="\\Server\Share" --exclude=".git,tmp" --out="D:\result.txt" --nfo
`
	fmt.Println(helpText)
	os.Exit(0)
}

func main() {
	// 1. Flags definieren
	rootPtr := flag.String("root", "", "Pfad zum Startverzeichnis")
	excludeRaw := flag.String("exclude", "", "Kommagetrennte Liste von auszuschließenden Ordnern")
	outPtr := flag.String("out", "missing_files.txt", "Pfad zur Ausgabedatei")
	nfoOnly := flag.Bool("nfo", false, "Nur fehlende .nfo-Dateien melden")
	jpgOnly := flag.Bool("jpg", false, "Nur fehlende .jpg-Dateien melden")
	help := flag.Bool("help", false, "Zeigt Hilfeinformationen an")
	flag.Parse()

	// 2. Hilfe anzeigen, falls angefordert oder --root fehlt
	if *help || *rootPtr == "" {
		showHelp()
	}

	rootDir = normalizePath(*rootPtr)
	outputFile = normalizePath(*outPtr)

	// 3. Ausschlüsse initialisieren und ergänzen
	// Standard-Ausschlüsse
	excludeDirs = []string{"#recycle", ".ds_store", "thumbs.db"}

	// Benutzerdefinierte Ausschlüsse ergänzen
	if *excludeRaw != "" {
		for _, e := range strings.Split(*excludeRaw, ",") {
			// Trimmen und zur Liste hinzufügen
			excludeDirs = append(excludeDirs, strings.TrimSpace(e))
		}
	}

	missingDirs := []string{}

	// 4. Rekursive Verzeichnisdurchsuchung mit filepath.WalkDir
	// Die Funktion erhält nun (path string, d fs.DirEntry, err error)
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Fehler bei Zugriff/Lesen eines Pfades melden, aber weiterlaufen
			fmt.Printf("Fehler beim Zugriff auf %s: %v\n", path, err)
			return nil
		}

		// Überprüfen, ob es ein Verzeichnis ist.
		if !d.IsDir() {
			return nil // Nur Ordner prüfen
		}

		// Ausschluss prüfen (für den aktuellen Ordnernamen)
		for _, exclude := range excludeDirs {
			// Prüfen, ob der Pfad den Ausschluss enthält (case-insensitive)
			if strings.Contains(strings.ToLower(path), strings.ToLower(exclude)) {
				return filepath.SkipDir // Komplettes Unterverzeichnis überspringen
			}
		}

		hasJPG, hasNFO := false, false

		// Verzeichnisinhalt lesen (Dieser Teil bleibt gleich, da wir die Inhalte des Verzeichnisses prüfen müssen)
		entries, err := os.ReadDir(path)
		if err != nil {
			// Fehler beim Lesen des Verzeichnisses melden, aber weiterlaufen (kann an Berechtigungen liegen)
			fmt.Printf("Fehler beim Lesen des Verzeichnisses %s: %v\n", path, err)
			return nil
		}

		// Verzeichnisinhalte nach gesuchten Dateien durchsuchen
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := strings.ToLower(entry.Name())

			if !hasJPG && strings.HasSuffix(name, ".jpg") {
				hasJPG = true
			}
			if !hasNFO && strings.HasSuffix(name, ".nfo") {
				hasNFO = true
			}

			if hasJPG && hasNFO {
				break
			}
		}

		// 5. Prüflogik anwenden
		missing := false
		switch {
		case *nfoOnly && !*jpgOnly:
			// Nur NFO fehlt
			missing = !hasNFO
		case *jpgOnly && !*nfoOnly:
			// Nur JPG fehlt
			missing = !hasJPG
		default:
			// Standard: Eines oder beide fehlen
			missing = !hasJPG || !hasNFO
		}

		if missing {
			missingDirs = append(missingDirs, path)
		}
		return nil
	})

	// 6. Fehlerbehandlung nach dem Walk
	if err != nil {
		fmt.Println("Fehler beim Durchsuchen des Wurzelverzeichnisses:", err)
		return
	}

	// 7. Ergebnis auswerten und berichten
	if len(missingDirs) == 0 {
		fmt.Println("Alle Ordner enthalten die geforderten Dateien.")
		return
	}

	// Zielordner erstellen falls nötig (um Fehler beim Schreiben zu vermeiden)
	outDir := filepath.Dir(outputFile)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		err = os.MkdirAll(outDir, 0755) // Erstellt alle notwendigen Ordner
		if err != nil {
			fmt.Println("Fehler beim Erstellen des Ausgabeordners:", err)
			return
		}
	}

	// 8. Ergebnis in Datei schreiben
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Fehler beim Schreiben der Ausgabedatei:", err)
		return
	}
	defer f.Close() // Stellt sicher, dass die Datei geschlossen wird

	for _, dir := range missingDirs {
		// Pfade in die Datei schreiben, gefolgt von einem Zeilenumbruch
		_, err := f.WriteString(dir + "\n")
		if err != nil {
			// Fehlerbehandlung, falls das Schreiben fehlschlägt
			break
		}
	}
}
