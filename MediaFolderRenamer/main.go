package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define command-line flags
	rootDir := flag.String("rootDir", "./", "Stammverzeichnis, von dem aus gestartet werden soll")
	extensions := flag.String("extensions", "mp4,mkv", "Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen")
	excludeDir := flag.String("excludeDir", "#recycle", "Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen")
	renameByFile := flag.Bool("renameByFile", false, "Umbenennen der Ordner basierend auf den Dateinamen im Ordner")
	renameByFolder := flag.Bool("renameByFolder", false, "Umbenennen von Dateien im Ordner basierend auf dem Ordnernamen")
	help := flag.Bool("help", false, "Zeigt die Hilfe an")
	flag.Parse()

	if *help {
		fmt.Println("Verwendung: MediaFolderRenamer [Optionen]")
		fmt.Println("\nOptionen:")
		fmt.Println("  --rootDir        Stammverzeichnis, von dem aus gestartet werden soll (Standard: ./). Unterstützt UNC-Pfade wie \\\\\\\\servername\\\\freigabe")
		fmt.Println("                   Hinweis: In der Kommandozeile müssen Backslashes (\\) durch doppelte Backslashes (\\\\) maskiert werden.")
		fmt.Println("  --extensions     Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen (Standard: mp4,mkv)")
		fmt.Println("  --excludeDir     Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen (Standard: #recycle)")
		fmt.Println("  --renameByFile   Umbenennen der Ordner basierend auf den Dateinamen im Ordner")
		fmt.Println("  --renameByFolder Umbenennen von Dateien im Ordner basierend auf dem Ordnernamen")
		fmt.Println("  --help           Zeigt diese Hilfe an")
		fmt.Println("\nBeispiel:")
		fmt.Println("  MediaFolderRenamer --rootDir=\\\\\\\\servername\\\\freigabe --extensions=mkv,mp4 --excludeDir=#recycle --excludeDir oder --renameByFolder")
		return
	}

	// Check if the root directory exists
	if _, err := os.Stat(*rootDir); os.IsNotExist(err) {
		fmt.Printf("Fehler: Das angegebene Stammverzeichnis %s existiert nicht oder ist nicht erreichbar.\n", *rootDir)
		return
	} else if err != nil {
		fmt.Printf("Fehler beim Zugriff auf das Stammverzeichnis %s: %v\n", *rootDir, err)
		return
	}

	fmt.Printf("Starte Verarbeitung im Verzeichnis: %s\n", *rootDir)

	extList := strings.Split(*extensions, ",")
	excludeList := strings.Split(*excludeDir, ",")

	// Sammle alle Unterverzeichnisse vor der Verarbeitung
	dirsToProcess := []string{}
	err := filepath.Walk(*rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Fehler beim Durchlaufen des Pfads %s: %v\n", path, err)
			return nil // Fehler ignorieren und weitermachen
		}

		// Ausschließen von Ordnern anhand des exakten Namens
		for _, excludeName := range excludeList {
			if strings.ToLower(filepath.Base(path)) == strings.ToLower(strings.TrimSpace(excludeName)) {
				// Den Ordner überspringen und nicht weiterverarbeiten
				return filepath.SkipDir
			}
		}

		if info.IsDir() && path != *rootDir {
			dirsToProcess = append(dirsToProcess, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Fehler beim Sammeln der Verzeichnisse: %v\n", err)
		return
	}

	// Verzeichnisse verarbeiten
	for _, dir := range dirsToProcess {
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Fehler beim Lesen des Verzeichnisses %s: %v\n", dir, err)
			continue
		}

		// Wenn renameByFile gesetzt ist, den Ordner basierend auf der ersten passenden Datei umbenennen
		if *renameByFile {
			foundFile := false
			for _, file := range files {
				if !file.IsDir() {
					ext := strings.ToLower(filepath.Ext(file.Name()))
					for _, validExt := range extList {
						if ext == "."+strings.TrimSpace(validExt) {
							foundFile = true
							newFolderName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
							parentDir := filepath.Dir(dir)
							newPath := filepath.Join(parentDir, newFolderName)

							// Rename the folder
							if err := os.Rename(dir, newPath); err != nil {
								fmt.Printf("Fehler beim Umbenennen des Ordners %s in %s: %v\n", dir, newPath, err)
							} else {
								fmt.Printf("Ordner %s in %s umbenannt\n", dir, newPath)
							}
							break // Nur die erste passende Datei verwenden
						}
					}
				}
			}
			if !foundFile {
				fmt.Printf("Keine passenden Dateien in %s gefunden, Überspringen...\n", dir)
			}
		}

		// Wenn renameByFolder gesetzt ist, Dateien im Ordner basierend auf dem Ordnernamen umbenennen
		if *renameByFolder {
			// Zähle die MP4- und MKV-Dateien im Ordner
			validFileCount := 0
			for _, file := range files {
				if !file.IsDir() {
					ext := strings.ToLower(filepath.Ext(file.Name()))
					for _, validExt := range extList {
						if ext == "."+strings.TrimSpace(validExt) {
							validFileCount++
							// Wenn mehr als eine gültige Datei gefunden wurde, überspringen
							if validFileCount > 1 {
								fmt.Printf("Ordner '%s' wird übersprungen, da er mehr als eine MP4 oder MKV-Datei enthält.\n", dir)
								break
							}
						}
					}
				}
			}

			// Wenn mehr als eine gültige Datei gefunden wurde, dann abbrechen
			if validFileCount > 1 {
				continue // Überspringe diesen Ordner und gehe zum nächsten
			}

			// Umbenennen der Datei, wenn nur eine gültige Datei vorhanden ist
			for _, file := range files {
				if !file.IsDir() {
					ext := strings.ToLower(filepath.Ext(file.Name()))
					for _, validExt := range extList {
						if ext == "."+strings.TrimSpace(validExt) {
							newFileName := filepath.Base(dir) + filepath.Ext(file.Name())
							newFilePath := filepath.Join(dir, newFileName)

							// Datei umbenennen
							if err := os.Rename(filepath.Join(dir, file.Name()), newFilePath); err != nil {
								fmt.Printf("Fehler beim Umbenennen der Datei %s in %s: %v\n", file.Name(), newFileName, err)
							} else {
								fmt.Printf("Datei %s in %s umbenannt\n", file.Name(), newFileName)
							}
							break // Nur die erste passende Datei verwenden
						}
					}
				}
			}
		}
	}
}
