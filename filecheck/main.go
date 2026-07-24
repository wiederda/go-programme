package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ------------------- Helper für UNC- und Pfadprobleme -------------------

func isUNCPath(p string) bool {
	return strings.HasPrefix(p, `\\`)
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	if isUNCPath(p) {
		return p
	}
	return filepath.Clean(p)
}

// ------------------- FileInfo -------------------

type FileInfo struct {
	Path string
	Size int64
}

// ------------------- Globaler Output -------------------

var outputWriter io.Writer

func initOutput(logPath string) (*os.File, error) {
	if logPath == "" {
		outputWriter = os.Stdout
		return nil, nil
	}

	logPath = normalizePath(logPath)

	// Prüfen, ob logPath ein Verzeichnis ist oder auf Slash endet
	info, err := os.Stat(logPath)
	if err == nil && info.IsDir() {
		logPath = filepath.Join(logPath, "duplicates.log")
	} else if strings.HasSuffix(logPath, `\`) || strings.HasSuffix(logPath, `/`) {
		logPath = filepath.Join(logPath, "duplicates.log")
	}
	// ansonsten wurde eine Datei angegeben → genau so verwenden

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen/Öffnen der Log-Datei %s: %w", logPath, err)
	}

	outputWriter = file
	fmt.Printf("✍️ Ausgabe wird in die Datei '%s' geschrieben (Anhängen).\n", logPath)
	return file, nil
}

func logf(format string, a ...interface{}) {
	fmt.Fprintf(outputWriter, format, a...)
}

// ------------------- Hashing & Suche -------------------

func calculateHash(filePath string) ([32]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [32]byte{}, err
	}

	var hashBytes [32]byte
	copy(hashBytes[:], hash.Sum(nil))
	return hashBytes, nil
}

func findFilesWithInfo(rootDir string) ([]*FileInfo, error) {
	var files []*FileInfo
	logf("Starte Suche in Ordner: %s\n", rootDir)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logf("[WARNUNG] Fehler beim Zugriff auf %s: %v\n", path, err)
			return nil
		}
		if !info.IsDir() && info.Size() > 0 {
			files = append(files, &FileInfo{Path: path, Size: info.Size()})
		}
		return nil
	})

	return files, err
}

func identifyDuplicates(allFiles []*FileInfo) map[[32]byte][]string {
	duplicates := make(map[[32]byte][]string)
	sizeGroups := make(map[int64][]*FileInfo)

	for _, f := range allFiles {
		sizeGroups[f.Size] = append(sizeGroups[f.Size], f)
	}

	fileCount := len(allFiles)
	processedCount := 0

	logf("\n--- Verarbeite %d Dateien ---\n", fileCount)

	for _, group := range sizeGroups {
		if len(group) < 2 {
			processedCount += len(group)
			continue
		}
		for _, f := range group {
			hash, err := calculateHash(f.Path)
			processedCount++
			fmt.Fprintf(os.Stderr, "\rBerechne Hashes... %d/%d", processedCount, fileCount)
			if err != nil {
				logf("\n[WARNUNG] Hash-Fehler für %s: %v\n", f.Path, err)
				continue
			}
			duplicates[hash] = append(duplicates[hash], f.Path)
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	logf("Hash-Berechnung abgeschlossen.\n")

	return duplicates
}

// ------------------- Backup -------------------

func moveDuplicatesToBackup(hashToFilePaths map[[32]byte][]string, backupDir string) (int, error) {
	if backupDir == "" {
		return 0, nil
	}

	backupDir = normalizePath(backupDir)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return 0, fmt.Errorf("konnte Backup-Ordner nicht erstellen: %w", err)
	}

	totalMoved := 0
	logf("\n--- Verschiebe Duplikate in Backup-Ordner: %s ---\n", backupDir)

	for hash, paths := range hashToFilePaths {
		if len(paths) > 1 {
			hashDir := filepath.Join(backupDir, fmt.Sprintf("%x", hash))
			if err := os.Mkdir(hashDir, 0755); err != nil {
				logf("[FEHLER] Konnte Unterordner %s nicht erstellen: %v\n", hashDir, err)
				continue
			}
			logf("   [BEHALTEN] Originalkopie: %s\n", paths[0])

			for _, oldPath := range paths[1:] {
				fileName := filepath.Base(oldPath)
				ext := filepath.Ext(fileName)
				nameWithoutExt := fileName[:len(fileName)-len(ext)]
				timestamp := time.Now().UnixNano()
				newFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
				newPath := filepath.Join(hashDir, newFileName)

				err := os.Rename(oldPath, newPath)
				if err != nil {
					logf("[FEHLER] Konnte Duplikat %s nicht nach %s verschieben: %v\n", oldPath, newPath, err)
				} else {
					logf("   [VERSCHOBEN] Duplikat: %s -> %s\n", oldPath, newPath)
					totalMoved++
				}
			}
		}
	}

	return totalMoved, nil
}

// ------------------- Main -------------------

func main() {
	dirPtr := flag.String("dir", "", "Der zu durchsuchende Startordner (erforderlich)")
	logPtr := flag.String("log", "", "Pfad zur Log-Datei. Konsole, wenn leer.")
	backupDirPtr := flag.String("backupdir", "", "OPTIONAL: Duplikate in Backup verschieben")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Verwendung: %s -dir <Ordnerpfad> [-log <Log-Datei>] [-backupdir <Backup-Pfad>]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Sucht rekursiv nach doppelten Dateien (gleiche Größe & SHA-256 Hash).")
		flag.PrintDefaults()
	}

	flag.Parse()

	rootDir := normalizePath(*dirPtr)
	backupDir := normalizePath(*backupDirPtr)
	logPath := normalizePath(*logPtr)

	if rootDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	info, err := os.Stat(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kritischer Fehler: Der Ordnerpfad '%s' ist ungültig: %v\n", rootDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Kritischer Fehler: Der Pfad '%s' ist kein Ordner.\n", rootDir)
		os.Exit(1)
	}

	logFile, err := initOutput(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kritischer Fehler: %v\n", err)
		os.Exit(1)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	start := time.Now()
	logf("========================================================\n")
	logf("Duplikatsuche gestartet am: %s\n", start.Format(time.RFC850))

	allFiles, err := findFilesWithInfo(rootDir)
	if err != nil {
		logf("Fehler beim Durchsuchen des Ordners: %v\n", err)
		return
	}

	if len(allFiles) == 0 {
		logf("Keine Dateien im Ordner gefunden.\n")
		return
	}

	hashToFilePaths := identifyDuplicates(allFiles)

	movedCount := 0
	if backupDir != "" {
		movedCount, err = moveDuplicatesToBackup(hashToFilePaths, backupDir)
		if err != nil {
			logf("[KRITISCHER FEHLER] Fehler beim Verschieben: %v\n", err)
		}
	} else {
		logf("\nOption -backupdir nicht angegeben. Keine Duplikate verschoben.\n")
		duplicateGroupsFound := false
		groupCounter := 0
		logf("\n--- Details der gefundenen Duplikatgruppen ---\n")
		for _, paths := range hashToFilePaths {
			if len(paths) > 1 {
				duplicateGroupsFound = true
				groupCounter++
				logf("Duplikatgruppe %d (gleicher Inhalt):\n", groupCounter)
				logf("  %d Kopien gefunden:\n", len(paths))
				for i, path := range paths {
					logf("    [%d] %s\n", i+1, path)
				}
				logf("\n")
			}
		}
		if !duplicateGroupsFound {
			logf("Keine Duplikate mit gleichem Inhalt gefunden.\n")
		}
		logf("---------------------------------------------------\n")
	}

	duration := time.Since(start)
	duplicateGroupsCount := 0
	totalDuplicateFiles := 0
	for _, paths := range hashToFilePaths {
		if len(paths) > 1 {
			duplicateGroupsCount++
			totalDuplicateFiles += len(paths) - 1
		}
	}

	logf("\n--- Abschlussbericht ---\n")
	logf("Gesamtzeit: %s\n", duration)

	if duplicateGroupsCount > 0 {
		logf("⚠️ **%d Gruppen von Duplikaten gefunden.**\n", duplicateGroupsCount)
		logf("   Das entspricht %d Duplikatdateien (ohne Originalkopie).\n", totalDuplicateFiles)
		if backupDir != "" {
			logf("📦 **%d Duplikate wurden in Backup verschoben.**\n", movedCount)
			logf("   Originalpfade sind im Log protokolliert.\n")
			logf("   Backup-Ordner: %s\n", backupDir)
		} else {
			logf("ℹ️ Duplikatgruppen im detaillierten Log-Abschnitt. Keine Dateien verschoben.\n")
		}
	} else {
		logf("👍 **Keine Duplikate gefunden.**\n")
	}
	logf("Duplikatsuche beendet.\n")
	logf("========================================================\n")
}
