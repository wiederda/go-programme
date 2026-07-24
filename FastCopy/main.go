package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// --- 1. GLOBALE KONFIGURATION ---

var (
	srcDir    string
	destDir   string
	mirror    bool
	emptyDirs bool
	noHash    bool
	verbose   bool
)

// Schwellenwert ab dem die gestufte Hash-Strategie greift (4 MB)
const partialHashThreshold = 4 * 1024 * 1024

// Größe der Partial-Hash-Blöcke (64 KB vom Anfang und Ende)
const partialHashBlockSize = 64 * 1024

// Maximale Pfadlänge unter Windows ohne Long Path Support
const windowsMaxPath = 260

// Struktur für alle Kopiervorgänge
type CopyTask struct {
	SourcePath string
	DestPath   string
	Reason     string
	Copy       bool
	Hash       bool
}

// HashResult (nur intern für Phase 2)
type HashResult struct {
	Task  CopyTask
	Equal bool
	Error error
}

// CopyResult (Finaler Output aller Worker)
type CopyResult struct {
	SourcePath string
	Reason     string
	Copied     bool
	Error      error
}

// --- 2. INITIALISIERUNG & FLAG-PARSING ---

func init() {
	flag.StringVar(&srcDir, "src", "", "Quellverzeichnis (Source Directory)")
	flag.StringVar(&destDir, "dest", "", "Zielverzeichnis (Destination Directory)")
	flag.BoolVar(&mirror, "MIR", false, "Spiegeln: Löscht Dateien/Ordner im Ziel, die nicht in der Quelle existieren.")
	flag.BoolVar(&emptyDirs, "E", false, "Kopiert leere Verzeichnisse.")
	flag.BoolVar(&noHash, "NO-HASH", false, "Deaktiviert den SHA-256-Inhaltsvergleich.")
	flag.BoolVar(&verbose, "v", false, "Aktiviert die detaillierte Protokollierung.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "FastCopy - Schnelles Dateisynchronisations-Tool (Go-basiert)\n")
		fmt.Fprintf(os.Stderr, "Verwendung: %s -src <QUELLE> -dest <ZIEL> [OPTIONEN]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptionen:\n")
		flag.PrintDefaults()
	}
}

// --- 3. HAUPTPROGRAMM ---

func main() {
	flag.Parse()

	if srcDir == "" || destDir == "" {
		fmt.Println("Fehler: Quell- und Zielverzeichnis (-src und -dest) müssen angegeben werden.")
		flag.Usage()
		os.Exit(1)
	}

	// FIX #3: Pfade normalisieren: filepath.Clean vereinheitlicht Separatoren,
	// entfernt trailing Slashes und behandelt Windows-Pfade wie C:\ korrekt –
	// ohne dass der Nutzer Backslashes escapen muss.
	srcDir = filepath.Clean(srcDir)
	destDir = filepath.Clean(destDir)

	// Verhindern, dass Quelle und Ziel identisch sind
	if srcDir == destDir {
		fmt.Fprintln(os.Stderr, "Fehler: Quell- und Zielverzeichnis sind identisch.")
		os.Exit(1)
	}

	// Verhindern, dass Ziel ein Unterverzeichnis der Quelle ist
	relCheck, err := filepath.Rel(srcDir, destDir)
	if err == nil && !strings.HasPrefix(relCheck, "..") {
		fmt.Fprintln(os.Stderr, "Fehler: Das Zielverzeichnis darf kein Unterverzeichnis des Quellverzeichnisses sein.")
		os.Exit(1)
	}

	startTime := time.Now()
	fmt.Printf("Starte FastCopy Synchronisation:\n  Quelle: %s\n  Ziel:   %s\n", srcDir, destDir)
	fmt.Printf("  Hash-Check: %s\n", map[bool]string{true: "Deaktiviert (/NO-HASH)", false: "Aktiviert (SHA-256, gestuft)"}[noHash])

	if err := syncDirectories(srcDir, destDir); err != nil {
		fmt.Fprintf(os.Stderr, "Synchronisation fehlgeschlagen: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nSynchronisation erfolgreich abgeschlossen.")
	fmt.Printf("Gesamtdauer: %s\n", time.Since(startTime))
}

// --- 4. SYNCHRONISATIONS-PHASEN (Pipeline Koordination) ---

func syncDirectories(source, destination string) error {
	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return fmt.Errorf("zielverzeichnis konnte nicht erstellt werden: %w", err)
	}

	// FIX #8: Verwaiste .tmp.fastcopy-Dateien aus früherem abgebrochenen Lauf aufräumen
	cleanupStaleTempFiles(destination)

	// PHASE 1: Scan & Filter
	tasksToHash, tasksToCopy, targetPaths, err := phase1_scanAndFilter(source, destination)
	if err != nil {
		return err
	}

	totalExpectedResults := len(tasksToCopy) + len(tasksToHash)
	if totalExpectedResults == 0 {
		fmt.Println("Keine Dateien zu synchronisieren.")
		if mirror {
			mirrorCleanup(destination, targetPaths)
		}
		return nil
	}

	// FIX #2: Deadlock-sichere Pipeline-Architektur.
	// Problem: copyTaskChan wurde von Phase 2 UND syncDirectories befüllt, aber
	// nur von Phase 2 geschlossen – wenn der gepufferte Channel voll läuft, blockiert
	// Phase 2 beim Senden und schließt den Channel nie → Deadlock.
	// Lösung: Zwei getrennte Input-Channels (directCopyChan, hashCopyChan),
	// die ein dedizierter Merger in den einen copyTaskChan für Phase 3 zusammenführt.
	// Der Merger schließt copyTaskChan erst, wenn BEIDE Quellen erschöpft sind.
	directCopyChan := make(chan CopyTask, len(tasksToCopy))
	hashCopyChan := make(chan CopyTask, len(tasksToHash)+1)
	copyTaskChan := make(chan CopyTask, totalExpectedResults)
	finalResultChan := make(chan CopyResult, totalExpectedResults)
	hashTaskChan := make(chan CopyTask, len(tasksToHash))

	var wg sync.WaitGroup

	// Merger: Liest aus directCopyChan und hashCopyChan, schreibt in copyTaskChan.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(copyTaskChan)
		var mergeWg sync.WaitGroup
		mergeWg.Add(2)
		go func() {
			defer mergeWg.Done()
			for t := range directCopyChan {
				copyTaskChan <- t
			}
		}()
		go func() {
			defer mergeWg.Done()
			for t := range hashCopyChan {
				copyTaskChan <- t
			}
		}()
		mergeWg.Wait()
	}()

	// Result Aggregator
	doneAggregating := make(chan struct{})
	var copied, skipped int
	go func() {
		copied, skipped = resultAggregator(finalResultChan)
		close(doneAggregating)
	}()

	wg.Add(2)
	go phase2_calculateHashes(tasksToHash, hashTaskChan, hashCopyChan, finalResultChan, &wg)
	go phase3_copyWorkers(copyTaskChan, finalResultChan, &wg)

	for _, task := range tasksToCopy {
		directCopyChan <- task
	}
	close(directCopyChan)

	for _, task := range tasksToHash {
		hashTaskChan <- task
	}
	close(hashTaskChan)

	wg.Wait()
	close(finalResultChan)
	<-doneAggregating

	fmt.Printf("\n--- Zusammenfassung ---\nKopiert: %d Dateien\nÜbersprungen: %d Dateien\n", copied, skipped)

	if mirror {
		mirrorCleanup(destination, targetPaths)
	}

	return nil
}

// resultAggregator zählt die finalen Ergebnisse.
func resultAggregator(results <-chan CopyResult) (int, int) {
	copiedCount := 0
	skippedCount := 0

	for res := range results {
		// FIX #3: filepath.Rel statt strings.TrimPrefix für korrekte Pfaddarstellung
		// auf allen Plattformen, inkl. Windows mit gemischten Separatoren.
		relPath, err := filepath.Rel(srcDir, res.SourcePath)
		if err != nil {
			relPath = res.SourcePath
		}

		if res.Error != nil {
			fmt.Printf("[FEHLER] Konnte %s nicht verarbeiten: %v\n", relPath, res.Error)
			skippedCount++
		} else if res.Copied {
			copiedCount++
		} else {
			if verbose {
				fmt.Printf("[ÜBERSPRINGE] %s (Grund: %s)\n", relPath, res.Reason)
			}
			skippedCount++
		}
	}

	return copiedCount, skippedCount
}

// --- HILFSFUNKTIONEN UND PHASEN-IMPLEMENTIERUNGEN ---

// PHASE 1: Scan & Filter
func phase1_scanAndFilter(source, destination string) ([]CopyTask, []CopyTask, map[string]bool, error) {
	fmt.Println("\nPhase 1: Starte Verzeichnis-Scan und Filterung...")

	var tasksToHash []CopyTask
	var tasksToCopy []CopyTask
	targetPaths := make(map[string]bool)

	err := filepath.Walk(source, func(quellPfad string, info os.FileInfo, err error) error {
		if err != nil {
			// Zugriffsfehler: loggen und weitermachen statt den gesamten Walk abzubrechen
			fmt.Printf("[WARNUNG] Zugriffsfehler bei %s: %v – wird übersprungen\n", quellPfad, err)
			return nil
		}

		relPfad, err := filepath.Rel(source, quellPfad)
		if err != nil {
			return fmt.Errorf("relativer pfad konnte nicht berechnet werden: %w", err)
		}

		zielPfad := filepath.Join(destination, relPfad)

		// FIX #5: Lange Pfade unter Windows (> 260 Zeichen) warnen und überspringen.
		// Ohne aktivierten Long Path Support in der Registry schlagen Dateioperationen
		// ab dieser Länge still fehl.
		if runtime.GOOS == "windows" && len(zielPfad) > windowsMaxPath {
			fmt.Printf("[WARNUNG] Pfad überschreitet %d Zeichen (Windows-Limit ohne Long Path Support): %s\n", windowsMaxPath, zielPfad)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Für Mirror-Map: einheitliche Slashes, plattformunabhängig vergleichbar
		normalizedRelPfad := filepath.ToSlash(relPfad)
		targetPaths[normalizedRelPfad] = true

		if info.IsDir() {
			if emptyDirs {
				if mkErr := os.MkdirAll(zielPfad, info.Mode()); mkErr != nil {
					fmt.Printf("[WARNUNG] Verzeichnis konnte nicht erstellt werden %s: %v\n", zielPfad, mkErr)
				}
			}
			return nil
		}

		// FIX #1: os.Stat-Fehler sauber aufdröseln.
		// Früher: `if os.IsNotExist(err) || info.Size() != zielInfo.Size()` –
		// bei einem anderen Stat-Fehler (z.B. Berechtigungsproblem) war zielInfo nil
		// und der Zugriff auf zielInfo.Size() führte zu einem Nil-Pointer-Panic.
		zielInfo, statErr := os.Stat(zielPfad)

		switch {
		case statErr != nil && !os.IsNotExist(statErr):
			// Unbekannter Fehler (Berechtigungen, kaputtes Dateisystem etc.) –
			// sicherheitshalber als "zu kopieren" behandeln
			fmt.Printf("[WARNUNG] Zieldatei nicht prüfbar %s: %v – wird als neu behandelt\n", zielPfad, statErr)
			tasksToCopy = append(tasksToCopy, CopyTask{SourcePath: quellPfad, DestPath: zielPfad, Reason: "Ziel-Stat fehlgeschlagen", Copy: true})

		case os.IsNotExist(statErr):
			tasksToCopy = append(tasksToCopy, CopyTask{SourcePath: quellPfad, DestPath: zielPfad, Reason: "Neu im Ziel", Copy: true})

		case info.Size() != zielInfo.Size():
			tasksToCopy = append(tasksToCopy, CopyTask{SourcePath: quellPfad, DestPath: zielPfad, Reason: "Größe unterschiedlich", Copy: true})

		case !noHash:
			tasksToHash = append(tasksToHash, CopyTask{SourcePath: quellPfad, DestPath: zielPfad, Hash: true})

		default:
			if verbose {
				fmt.Printf("[ÜBERSPRINGE] %s (Größe gleich, Hash deaktiviert)\n", relPfad)
			}
		}

		return nil
	})

	fmt.Printf("Phase 1 abgeschlossen. %d direkte Kopien geplant, %d Dateien benötigen Hash-Check.\n", len(tasksToCopy), len(tasksToHash))
	return tasksToHash, tasksToCopy, targetPaths, err
}

// PHASE 2: Hash-Berechnung (Pipeline-Sender)
// FIX #9: Der `tasks []CopyTask`-Parameter ist konzeptuell überflüssig – der echte
// Input kommt über taskChan. Er wird nur noch für die interne Buffer-Größe übergeben
// und ist klar als solcher dokumentiert. hashResultsChan nutzt nun numWorkers*2
// als dynamischen Puffer statt der irreführenden len(tasks).
func phase2_calculateHashes(tasks []CopyTask, taskChan <-chan CopyTask, copyChan chan<- CopyTask, resultChan chan<- CopyResult, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(copyChan)

	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	}

	var hashWg sync.WaitGroup
	hashResultsChan := make(chan HashResult, numWorkers*2)

	for i := 0; i < numWorkers; i++ {
		hashWg.Add(1)
		go func() {
			defer hashWg.Done()
			for task := range taskChan {
				equal, err := tieredHashEqual(task.SourcePath, task.DestPath)
				hashResultsChan <- HashResult{Task: task, Equal: equal, Error: err}
			}
		}()
	}

	go func() {
		hashWg.Wait()
		close(hashResultsChan)
	}()

	for res := range hashResultsChan {
		if res.Error != nil {
			resultChan <- CopyResult{
				SourcePath: res.Task.SourcePath,
				Reason:     "Hash-Fehler",
				Error:      res.Error,
			}
			continue
		}

		if !res.Equal {
			copyChan <- CopyTask{
				SourcePath: res.Task.SourcePath,
				DestPath:   res.Task.DestPath,
				Reason:     "Hash unterschiedlich",
			}
		} else {
			resultChan <- CopyResult{
				SourcePath: res.Task.SourcePath,
				Reason:     "Hash gleich",
			}
		}
	}
}

// tieredHashEqual implementiert die gestufte Hash-Strategie:
//  1. Bei kleinen Dateien (< partialHashThreshold): direkt Full-Hash
//  2. Bei großen Dateien: erst Partial-Hash (Anfang + Ende), bei Gleichheit Full-Hash
//
// FIX #7: Seek-Fehler auf Netzlaufwerken (SMB ohne random-access) werden abgefangen
// und lösen einen graceful Fallback auf Full-Hash aus.
func tieredHashEqual(srcPath, dstPath string) (bool, error) {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, fmt.Errorf("stat quelle: %w", err)
	}

	if srcInfo.Size() < partialHashThreshold {
		return fullHashEqual(srcPath, dstPath)
	}

	if verbose {
		relPath, _ := filepath.Rel(srcDir, srcPath)
		fmt.Printf("[HASH] Partial-Hash für %s (%.1f MB)...\n", relPath, float64(srcInfo.Size())/1024/1024)
	}

	srcPartial, err := partialHash(srcPath, srcInfo.Size())
	if err != nil {
		return false, fmt.Errorf("partial-hash quelle: %w", err)
	}

	dstPartial, err := partialHash(dstPath, srcInfo.Size())
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// FIX #7: Seek-Fehler (ESPIPE, "illegal seek") auf Netzlaufwerken abfangen
		if isSeekError(err) {
			if verbose {
				relPath, _ := filepath.Rel(srcDir, srcPath)
				fmt.Printf("[HASH] Seek nicht unterstützt für %s – Fallback auf Full-Hash\n", relPath)
			}
			return fullHashEqual(srcPath, dstPath)
		}
		return false, fmt.Errorf("partial-hash ziel: %w", err)
	}

	if srcPartial != dstPartial {
		if verbose {
			relPath, _ := filepath.Rel(srcDir, srcPath)
			fmt.Printf("[HASH] Partial-Hash verschieden für %s – Full-Hash übersprungen\n", relPath)
		}
		return false, nil
	}

	if verbose {
		relPath, _ := filepath.Rel(srcDir, srcPath)
		fmt.Printf("[HASH] Partial-Hash gleich für %s – starte Full-Hash...\n", relPath)
	}
	return fullHashEqual(srcPath, dstPath)
}

// isSeekError erkennt Fehler durch nicht unterstütztes Seek.
// Relevant für SMB/NFS-Shares und bestimmte FUSE-Dateisysteme.
func isSeekError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "seek") ||
		strings.Contains(msg, "espipe") ||
		strings.Contains(msg, "illegal seek") ||
		strings.Contains(msg, "invalid seek")
}

// partialHash liest partialHashBlockSize Bytes vom Anfang UND Ende der Datei
// und kombiniert sie mit der Dateigröße zu einem SHA-256-Hash.
func partialHash(filePath string, fileSize int64) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()

	// Dateigröße als Seed einbeziehen, verhindert Kollisionen bei gleich großen Dateien
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(fileSize))
	h.Write(sizeBytes)

	// Ersten Block lesen
	headBuf := make([]byte, partialHashBlockSize)
	n, err := io.ReadFull(f, headBuf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("lesefehler anfang: %w", err)
	}
	h.Write(headBuf[:n])

	// Letzten Block lesen (nur wenn Datei größer als 2 × partialHashBlockSize)
	if fileSize > int64(2*partialHashBlockSize) {
		tailOffset := fileSize - int64(partialHashBlockSize)
		if _, err := f.Seek(tailOffset, io.SeekStart); err != nil {
			return "", fmt.Errorf("seek ende: %w", err)
		}
		tailBuf := make([]byte, partialHashBlockSize)
		n, err = io.ReadFull(f, tailBuf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("lesefehler ende: %w", err)
		}
		h.Write(tailBuf[:n])
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// fullHashEqual berechnet den vollständigen SHA-256 für beide Dateien und vergleicht sie.
func fullHashEqual(srcPath, dstPath string) (bool, error) {
	srcHash, err := calculateHash(srcPath, sha256.New())
	if err != nil {
		return false, fmt.Errorf("full-hash quelle: %w", err)
	}
	dstHash, err := calculateHash(dstPath, sha256.New())
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("full-hash ziel: %w", err)
	}
	return fmt.Sprintf("%x", srcHash) == fmt.Sprintf("%x", dstHash), nil
}

// PHASE 3: Kopieren (Dedizierter Worker-Pool)
func phase3_copyWorkers(tasks <-chan CopyTask, results chan<- CopyResult, wg *sync.WaitGroup) {
	defer wg.Done()

	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	}

	var copyWg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		copyWg.Add(1)
		go func() {
			defer copyWg.Done()
			for task := range tasks {
				res := CopyResult{
					SourcePath: task.SourcePath,
					Reason:     task.Reason,
				}
				if err := copyFileAtomic(task.SourcePath, task.DestPath); err != nil {
					res.Error = fmt.Errorf("Kopierfehler: %w", err)
				} else {
					res.Copied = true
				}
				results <- res
			}
		}()
	}
	copyWg.Wait()
}

// calculateHash berechnet den vollständigen Hash für eine einzelne Datei.
func calculateHash(filePath string, h hash.Hash) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("daten-lesefehler: %w", err)
	}

	return h.Sum(nil), nil
}

// copyFileAtomic kopiert eine Datei crash-sicher:
//  1. Schreibt in eine temporäre Datei (.tmp.fastcopy) im Zielverzeichnis
//  2. Sync() – Flush auf Disk
//  3. os.Rename() – atomares Umbenennen zur Zieldatei
//
// FIX #4: Cross-Device Fallback: os.Rename() schlägt fehl, wenn Quelle und Ziel
// auf verschiedenen Dateisystemen liegen (lokale Disk → NAS, SMB, etc.).
// In diesem Fall: manueller Kopier+Lösch-Fallback via crossDeviceMove().
//
// FIX #6: Sync()-Fehler auf Netzlaufwerken (SMB/NFS ohne fsync-Support)
// werden als Warnung behandelt, nicht als fataler Fehler.
func copyFileAtomic(srcPath, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat quelle: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return fmt.Errorf("zielverzeichnis erstellen: %w", err)
	}

	tmpPath := dstPath + ".tmp.fastcopy"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("temp-datei erstellen: %w", err)
	}

	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("quelle öffnen: %w", err)
	}
	defer srcFile.Close()

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		return fmt.Errorf("kopieren: %w", err)
	}

	// FIX #6: Sync-Fehler nur warnen, nicht abbrechen.
	// Netzlaufwerke (SMB, NFS, exFAT) unterstützen fsync oft nicht.
	// Da io.Copy bereits erfolgreich war, sind die Daten übertragen.
	if err := tmpFile.Sync(); err != nil {
		fmt.Printf("[WARNUNG] Sync für %s nicht unterstützt (Netzlaufwerk?): %v\n", dstPath, err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("temp-datei schließen: %w", err)
	}

	// Metadaten auf die temporäre Datei setzen, bevor sie sichtbar wird
	mtime := srcInfo.ModTime()
	if err := os.Chtimes(tmpPath, mtime, mtime); err != nil {
		fmt.Printf("[WARNUNG] Zeitstempel für %s konnte nicht gesetzt werden: %v\n", dstPath, err)
	}
	if err := os.Chmod(tmpPath, srcInfo.Mode()); err != nil {
		fmt.Printf("[WARNUNG] Dateirechte für %s konnten nicht gesetzt werden: %v\n", dstPath, err)
	}

	// FIX #4: Atomares Rename mit Cross-Device-Fallback.
	// POSIX: garantiert atomar. Windows seit Go 1.21: ebenfalls atomar (MoveFileExW).
	// Cross-Device (verschiedene Laufwerke/Shares): Fallback auf crossDeviceMove().
	if err := os.Rename(tmpPath, dstPath); err != nil {
		if isCrossDeviceError(err) {
			if verbose {
				fmt.Printf("[INFO] Cross-Device Rename für %s – nutze Kopier-Fallback\n", dstPath)
			}
			if moveErr := crossDeviceMove(tmpPath, dstPath); moveErr != nil {
				return moveErr
			}
			success = true
			return nil
		}
		return fmt.Errorf("atomares umbenennen: %w", err)
	}

	success = true
	return nil
}

// isCrossDeviceError erkennt Cross-Device-Rename-Fehler plattformübergreifend.
func isCrossDeviceError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cross-device") ||
		strings.Contains(msg, "invalid cross-device link") ||
		strings.Contains(msg, "the system cannot move the file") ||
		strings.Contains(msg, "across drives")
}

// crossDeviceMove ist der Fallback-Kopierer für Dateisystemgrenzen.
// Nicht atomar, aber der einzig mögliche Weg über Gerätegrenzen.
// Im Fehlerfall bleibt die Zieldatei im alten Zustand (wir schreiben direkt in dstPath).
func crossDeviceMove(tmpPath, dstPath string) error {
	src, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("cross-device: temp öffnen: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("cross-device: ziel öffnen: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("cross-device: kopieren: %w", err)
	}

	src.Close()
	os.Remove(tmpPath)
	return nil
}

// cleanupStaleTempFiles entfernt verwaiste .tmp.fastcopy-Dateien aus einem
// früheren abgebrochenen Lauf. Wird einmalig zu Beginn von syncDirectories aufgerufen.
// FIX #8: Ohne -MIR bleiben diese Dateien sonst dauerhaft im Zielverzeichnis liegen.
func cleanupStaleTempFiles(dir string) {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".tmp.fastcopy") {
			if removeErr := os.Remove(path); removeErr == nil {
				count++
				if verbose {
					fmt.Printf("[AUFRÄUMEN] Verwaiste Temp-Datei entfernt: %s\n", path)
				}
			}
		}
		return nil
	})
	if count > 0 {
		fmt.Printf("[AUFRÄUMEN] %d verwaiste Temp-Datei(en) aus früherem Lauf entfernt.\n", count)
	}
}

// PHASE 5 (Optional): Mirror Cleanup
func mirrorCleanup(destination string, sourcePaths map[string]bool) {
	fmt.Println("\nPhase 5: Starte MIRROR-Bereinigung (/MIR)...")

	filepath.Walk(destination, func(zielPfad string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relativerPfad, relErr := filepath.Rel(destination, zielPfad)
		if relErr != nil || relativerPfad == "." {
			return nil
		}

		// .tmp.fastcopy-Dateien wurden bereits durch cleanupStaleTempFiles entfernt;
		// Restbestände (Rennen zwischen laufendem Sync und Mirror) überspringen.
		if strings.HasSuffix(zielPfad, ".tmp.fastcopy") {
			return nil
		}

		normalizedPath := filepath.ToSlash(relativerPfad)

		if !sourcePaths[normalizedPath] {
			if info.IsDir() {
				if err := os.RemoveAll(zielPfad); err != nil {
					fmt.Printf("[FEHLER] Konnte Ordner %s nicht löschen: %v\n", relativerPfad, err)
				} else {
					fmt.Printf("[GELÖSCHT ORDNER] %s\n", relativerPfad)
				}
				return filepath.SkipDir
			}
			if err := os.Remove(zielPfad); err != nil {
				fmt.Printf("[FEHLER] Konnte Datei %s nicht löschen: %v\n", relativerPfad, err)
			} else {
				fmt.Printf("[GELÖSCHT DATEI] %s\n", relativerPfad)
			}
		}

		return nil
	})
}
