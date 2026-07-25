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
	exitPartial    = 9
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

// logOutput steuert global, ob Fortschritt und Fehler ausgegeben werden.
// Ohne -log läuft updater komplett still (kein stdout, kein stderr) - für
// unbeaufsichtigte Läufe (Cron etc.), wo ohnehin nur der Exit Code zählt.
var logOutput bool

func main() {

	if len(os.Args) == 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help") {

		help()
		os.Exit(exitOK)
	}

	args := os.Args[1:]

	if len(args) > 0 && args[0] == "-log" {
		logOutput = true
		args = args[1:]
	}

	switch {

	case len(args) == 2 && (args[0] == "-j" || args[0] == "--json"):
		runJSON(args[1])

	case len(args) == 2:
		runSingle(args[0], args[1])

	default:
		errorExit("usage", exitParam)
	}
}

// runSingle behandelt den klassischen Ein-Datei-Modus: updater [-log] <alt> <neu>
//
// Standardmäßig läuft dieser Modus komplett still (weder stdout noch
// stderr). Mit -log wird der Erfolg zusätzlich gemeldet.
func runSingle(oldFile, newFile string) {

	if err := update(oldFile, newFile); err != nil {
		errorExit(err.Error(), err.code)
	}

	if logOutput {
		fmt.Println("OK")
	}

	os.Exit(exitOK)
}

// runJSON behandelt den Batch-Modus: updater [-log] -j <config.json>
//
// Jeder Eintrag wird unabhängig von den anderen behandelt: schlägt ein
// Eintrag fehl (Datei fehlt, Backup-Fehler, Berechtigungsfehler, ...),
// wird das gemeldet, der Batch aber NICHT abgebrochen - die restlichen
// Einträge werden trotzdem abgearbeitet. Ein automatisches Rollback
// bereits ersetzter Dateien findet nicht statt.
//
// Standardmäßig läuft dieser Modus komplett still. Mit -log wird jeder
// erfolgreiche und jeder fehlgeschlagene Schritt gemeldet.
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

	hadError := false

	for i, entry := range cfg.Files {

		if entry.Old == "" || entry.New == "" {
			reportEntryError(i, entry.Old, "Eintrag unvollständig (old/new erforderlich)")
			hadError = true
			continue
		}

		if err := update(entry.Old, entry.New); err != nil {
			reportEntryError(i, entry.Old, err.Error())
			hadError = true
			continue
		}

		if logOutput {
			fmt.Printf("OK: %s\n", entry.Old)
		}
	}

	if hadError {
		os.Exit(exitPartial)
	}

	if logOutput {
		fmt.Println("OK")
	}

	os.Exit(exitOK)
}

// reportEntryError meldet den Fehler eines einzelnen Batch-Eintrags auf
// stderr, sofern -log gesetzt ist, ohne das Programm zu beenden.
func reportEntryError(index int, oldFile, msg string) {
	if logOutput {
		fmt.Fprintf(os.Stderr, "ERROR: Eintrag %d (%s): %s\n", index+1, oldFile, msg)
	}
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
	// Kein automatisches Rollback: schlägt dieser Schritt fehl, bleibt
	// der aktuelle Zustand unverändert bestehen. Das Backup (<alt>.old)
	// ist vorhanden, eine Wiederherstellung muss manuell erfolgen.
	if err := os.Rename(newFile, oldFile); err != nil {
		return &updateError{
			msg:  "neue Datei konnte nicht eingesetzt werden: " + err.Error(),
			code: exitReplace,
		}
	}

	// -------------------------
	// Rechte übernehmen
	// -------------------------
	// Kein automatisches Rollback: die neue Datei ist an dieser Stelle
	// bereits eingesetzt, nur die Rechte konnten nicht gesetzt werden.
	// Das Backup (<alt>.old) bleibt für eine manuelle Wiederherstellung
	// erhalten.
	if err := os.Chmod(oldFile, oldMode); err != nil {
		return &updateError{
			msg:  "Dateirechte konnten nicht gesetzt werden: " + err.Error(),
			code: exitPermission,
		}
	}

	// -------------------------
	// Prüfung
	// -------------------------
	// Auch hier kein automatisches Rollback - nur eine Fehlermeldung,
	// damit klar ist, dass die abschließende Prüfung nicht bestanden
	// wurde. Das Backup (<alt>.old) bleibt in jedem Fall (Erfolg wie
	// Fehler) liegen und wird erst beim nächsten Lauf für dieselbe
	// Datei entfernt (siehe oben, Schritt "vorhandenes Backup löschen").
	// Eine Wiederherstellung aus dem Backup muss der Anwender manuell
	// vornehmen.
	check, err := os.Stat(oldFile)

	if err != nil ||
		check.Mode().Perm() != oldMode {

		return &updateError{
			msg:  "Dateiprüfung fehlgeschlagen: " + oldFile,
			code: exitPermission,
		}
	}

	return nil
}

func help() {

	fmt.Println("updater")
	fmt.Println()
	fmt.Println("Aktualisiert eine oder mehrere Dateien atomar mit Backup.")
	fmt.Println()
	fmt.Println("Aufruf:")
	fmt.Println("  updater [-log] <alt> <neu>")
	fmt.Println("  updater [-log] -j <config.json>")
	fmt.Println()
	fmt.Println("  -log   gibt Fortschritt und Fehler aus (OK je Datei,")
	fmt.Println("         abschließend OK, Fehler mit ERROR-Präfix). Ohne")
	fmt.Println("         -log läuft das Tool komplett still - weder")
	fmt.Println("         stdout noch stderr, nur der Exit Code zählt.")
	fmt.Println()
	fmt.Println("JSON-Format:")
	fmt.Println(`  { "files": [ { "old": "...", "new": "..." }, ... ] }`)
	fmt.Println()
	fmt.Println("Im JSON-Modus wird ein fehlerhafter Eintrag gemeldet, der")
	fmt.Println("Batch aber NICHT abgebrochen - die restlichen Einträge")
	fmt.Println("werden trotzdem abgearbeitet. Bereits ersetzte Dateien")
	fmt.Println("werden NICHT automatisch zurückgerollt.")
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
	fmt.Println("  9  mindestens ein Eintrag im JSON-Modus fehlgeschlagen")
}

// errorExit meldet einen Fehler auf stderr, sofern -log gesetzt ist, und
// beendet das Programm mit dem passenden Exit Code. Der Exit Code wird
// immer gesetzt, unabhängig von -log - nur die Textausgabe ist optional.
func errorExit(msg string, code int) {

	if logOutput {
		fmt.Fprintln(os.Stderr, "ERROR:", msg)
	}

	os.Exit(code)
}
