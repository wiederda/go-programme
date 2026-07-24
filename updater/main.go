package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	exitOK         = 0
	exitParam      = 1
	exitOldMissing = 2
	exitNewMissing = 3
	exitBackup     = 4
	exitReplace    = 5
	exitPermission = 6
	exitJSON       = 7
)

type updateError struct {
	msg  string
	code int
}

func (e updateError) Error() string {
	return e.msg
}

// fileEntry beschreibt ein einzelnes Alt/Neu-Paar in der JSON-Konfiguration.
type fileEntry struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// jsonConfig ist die Wurzelstruktur der JSON-Konfigurationsdatei.
type jsonConfig struct {
	Files []fileEntry `json:"files"`
}

func main() {

	if len(os.Args) == 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help") {

		help()
		os.Exit(exitOK)
	}

	switch {

	case len(os.Args) == 3 && (os.Args[1] == "-j" || os.Args[1] == "--json"):
		runJSON(os.Args[2])

	case len(os.Args) == 3:
		runSingle(os.Args[1], os.Args[2])

	default:
		errorExit("usage", exitParam)
	}
}

// runSingle behandelt den klassischen Ein-Datei-Modus: updater <alt> <neu>
func runSingle(oldFile, newFile string) {

	if err := update(oldFile, newFile); err != nil {
		errorExit(err.Error(), err.code)
	}

	fmt.Println("OK")
	os.Exit(exitOK)
}

// runJSON behandelt den Batch-Modus: updater -j <config.json>
//
// Die Dateien werden nacheinander abgearbeitet. Schlägt ein Eintrag fehl,
// wird abgebrochen. Bereits erfolgreich ersetzte Dateien werden NICHT
// automatisch zurückgerollt - das ist bewusst so gehalten, ein Rollback
// über mehrere Dateien hinweg müsste im Zweifel manuell erfolgen.
func runJSON(configFile string) {

	data, ioErr := os.ReadFile(configFile)
	if ioErr != nil {
		errorExit("Konfigurationsdatei nicht gefunden: "+ioErr.Error(), exitJSON)
	}

	var cfg jsonConfig
	if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
		errorExit("Konfigurationsdatei ungültig: "+jsonErr.Error(), exitJSON)
	}

	if len(cfg.Files) == 0 {
		errorExit("Konfigurationsdatei enthält keine Einträge", exitJSON)
	}

	for i, entry := range cfg.Files {

		if entry.Old == "" || entry.New == "" {
			errorExit(
				fmt.Sprintf("Eintrag %d unvollständig (old/new erforderlich)", i+1),
				exitJSON,
			)
		}

		if err := update(entry.Old, entry.New); err != nil {
			errorExit(
				fmt.Sprintf("Eintrag %d (%s): %s", i+1, entry.Old, err.Error()),
				err.code,
			)
		}

		fmt.Printf("OK: %s\n", entry.Old)
	}

	fmt.Println("OK")
	os.Exit(exitOK)
}

// update ersetzt eine einzelne Datei atomar mit Backup.
func update(oldFile, newFile string) *updateError {

	// -------------------------
	// Prüfen alte Datei
	// -------------------------
	oldInfo, err := os.Stat(oldFile)
	if err != nil {
		return &updateError{
			msg:  "alte Datei nicht gefunden: " + oldFile,
			code: exitOldMissing,
		}
	}

	// -------------------------
	// Prüfen neue Datei
	// -------------------------
	if _, err := os.Stat(newFile); err != nil {
		return &updateError{
			msg:  "neue Datei nicht gefunden: " + newFile,
			code: exitNewMissing,
		}
	}

	oldMode := oldInfo.Mode().Perm()
	backup := oldFile + ".old"

	// -------------------------
	// vorhandenes Backup löschen
	// -------------------------
	if _, err := os.Stat(backup); err == nil {

		if err := os.Remove(backup); err != nil {
			return &updateError{
				msg:  "altes Backup konnte nicht gelöscht werden: " + backup,
				code: exitBackup,
			}
		}
	}

	// -------------------------
	// alte Datei sichern
	// -------------------------
	if err := os.Rename(oldFile, backup); err != nil {
		return &updateError{
			msg:  "Backup konnte nicht erstellt werden: " + err.Error(),
			code: exitBackup,
		}
	}

	// -------------------------
	// neue Datei einsetzen
	// -------------------------
	if err := os.Rename(newFile, oldFile); err != nil {

		// Rollback
		_ = os.Rename(backup, oldFile)

		return &updateError{
			msg:  "neue Datei konnte nicht eingesetzt werden: " + err.Error(),
			code: exitReplace,
		}
	}

	// -------------------------
	// Rechte übernehmen
	// -------------------------
	if err := os.Chmod(oldFile, oldMode); err != nil {

		// Rollback
		_ = os.Remove(oldFile)
		_ = os.Rename(backup, oldFile)

		return &updateError{
			msg:  "Dateirechte konnten nicht gesetzt werden: " + err.Error(),
			code: exitPermission,
		}
	}

	// -------------------------
	// Prüfung
	// -------------------------
	check, err := os.Stat(oldFile)

	if err != nil ||
		check.Mode().Perm() != oldMode {

		// Rollback
		_ = os.Remove(oldFile)
		_ = os.Rename(backup, oldFile)

		return &updateError{
			msg:  "Dateiprüfung fehlgeschlagen: " + oldFile,
			code: exitPermission,
		}
	}

	// Backup (<alt>.old) bleibt bewusst liegen. Es wird erst beim
	// nächsten Lauf für dieselbe Datei entfernt (siehe oben, Schritt
	// "vorhandenes Backup löschen"), damit ein manuelles Rollback nach
	// einem erfolgreichen Update jederzeit möglich bleibt.

	return nil
}

func help() {

	fmt.Println("updater")
	fmt.Println()
	fmt.Println("Aktualisiert eine oder mehrere Dateien atomar mit Backup.")
	fmt.Println()
	fmt.Println("Aufruf:")
	fmt.Println("  updater <alt> <neu>")
	fmt.Println("  updater -j <config.json>")
	fmt.Println()
	fmt.Println("JSON-Format:")
	fmt.Println(`  { "files": [ { "old": "...", "new": "..." }, ... ] }`)
	fmt.Println()
	fmt.Println("Im JSON-Modus wird bei einem Fehler abgebrochen. Bereits")
	fmt.Println("ersetzte Dateien werden NICHT automatisch zurückgerollt.")
	fmt.Println()
	fmt.Println("Exit Codes:")
	fmt.Println("  0  Erfolg")
	fmt.Println("  1  Parameterfehler")
	fmt.Println("  2  alte Datei fehlt")
	fmt.Println("  3  neue Datei fehlt")
	fmt.Println("  4  Backup Fehler")
	fmt.Println("  5  Austausch Fehler")
	fmt.Println("  6  Rechte Fehler")
	fmt.Println("  7  JSON/Konfigurationsfehler")
}

func errorExit(msg string, code int) {

	fmt.Fprintln(os.Stderr, "ERROR:", msg)

	os.Exit(code)
}
