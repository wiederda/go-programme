package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define flags
	rootDir := flag.String("rootDir", "./", "Stammverzeichnis, von dem aus gestartet werden soll (Standard: ./). Unterstützt UNC-Pfade wie \\\\servername\\freigabe.")
	extensions := flag.String("extensions", "mp4,mkv", "Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen (Standard: mp4,mkv).")
	excludeDir := flag.String("excludeDir", "#recycle", "Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen (Standard: #recycle).")
	outputFile := flag.String("outputFile", "output.txt", "Optionaler Name der Datei, in der die Ergebnisse gespeichert werden (Standard: output.txt).")
	help := flag.Bool("help", false, "Zeigt diese Hilfe an.")

	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Println("Verwendung: FolderFileMatcher [Optionen]")
		fmt.Println("\nOptionen:")
		fmt.Println("  --rootDir        Stammverzeichnis, von dem aus gestartet werden soll (Standard: ./). Unterstützt UNC-Pfade wie \\\\servername\\freigabe")
		fmt.Println("                   Hinweis: In der Kommandozeile müssen Backslashes (\\) durch doppelte Backslashes (\\\\) maskiert werden.")
		fmt.Println("  --extensions     Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen (Standard: mp4,mkv)")
		fmt.Println("  --excludeDir     Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen (Standard: #recycle)")
		fmt.Println("  --outputFile     Optionaler Name der Datei, in der die Ergebnisse gespeichert werden (Standard: output.txt)")
		fmt.Println("  --help           Zeigt diese Hilfe an")
		fmt.Println("\nBeispiel:")
		fmt.Println("  FolderFileMatcher --rootDir=\\\\\\\\servername\\\\freigabe --extensions=mkv,mp4 --excludeDir=#recycle --outputFile=results.txt")
		return
	}

	// Validate rootDir
	if *rootDir == "" {
		fmt.Println("Bitte geben Sie ein gültiges Stammverzeichnis an. Nutzen Sie --help für weitere Informationen.")
		os.Exit(1)
	}

	var mismatchedPaths []string

	extensionList := strings.Split(*extensions, ",")
	excludeDirs := strings.Split(*excludeDir, ",")

	err := filepath.Walk(*rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip files directly in rootDir
		relPath, err := filepath.Rel(*rootDir, path)
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Dir(relPath) == "." {
			// Ignoriere Dateien direkt im rootDir
			return nil
		}

		// Skip excluded directories
		for _, exclude := range excludeDirs {
			if info.IsDir() && strings.EqualFold(info.Name(), exclude) {
				return filepath.SkipDir
			}
		}

		// Process files in subdirectories
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			for _, validExt := range extensionList {
				if ext == "."+strings.ToLower(validExt) {
					folderName := filepath.Base(filepath.Dir(path))
					fileName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

					if folderName != fileName {
						mismatchedPaths = append(mismatchedPaths, path)
					}
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Fehler beim Durchsuchen des Verzeichnisses: %v\n", err)
		os.Exit(1)
	}

	if len(mismatchedPaths) > 0 {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Fehler beim Erstellen der Datei %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
		defer file.Close()

		for _, path := range mismatchedPaths {
			file.WriteString(path + "\n")
		}

		fmt.Printf("Mismatch-Pfade wurden in %s gespeichert.\n", *outputFile)
	} else {
		fmt.Println("Keine Mismatches gefunden.")
	}
}
