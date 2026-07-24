package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	exitAborted    = 8
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

func main() {

	if len(os.Args) == 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help") {

		help()
		os.Exit(exitOK)
	}

	args := os.Args[1:]

	interactive := false
	if len(args) > 0 && args[0] == "-i" {
		interactive = true
		args = args[1:]
	}

	switch {

	case len(args) == 2 && (args[0] == "-j" || args[0] == "--json"):
		runJSON(args[1], interactive)

	case len(args) == 2:
		runSingle(args[0], args[1], interactive)

	default:
		errorExit("usage", exitParam)
	}
}

// runSingle behandelt den klassischen Ein-Datei-Modus: updater [-i] <alt> <neu>
//
// Standardmäßig läuft dieser Modus still (keine stdout-Ausgabe außer bei
// Fehlern). Mit -i wird vor dem Austausch eine Rückfrage gestellt und der
// Erfolg zusätzlich gemeldet.
func runSingle(oldFile, newFile string, interactive bool) {

	if interactive && !confirm(oldFile, newFile) {
		errorExit("vom Benutzer abgebrochen: "+oldFile, exitAborted)
	}

	if err := update(oldFile, newFile); err != nil {
		errorExit(err.Error(), err.code)
	}

	if interactive {
		fmt.Println("OK")
	}

	os.Exit(exitOK)
}

// runJSON behandelt den Batch-Modus: updater [-i] -j <config.json>
//
// Jeder Eintrag wird unabhängig von den anderen behandelt: schlägt ein
// Eintrag fehl (Datei fehlt, Backup-Fehler, Berechtigungsfehler, ...),
// wird das gemeldet, der Batch aber NICHT abgebrochen - die restlichen
// Einträge werden trotzdem abgearbeitet. Ein automatisches Rollback
// bereits ersetzter Dateien findet nicht statt.
//
// Standardmäßig läuft dieser Modus still (keine stdout-Ausgabe für
// erfolgreiche Einträge). Fehler werden immer auf stderr gemeldet,
// unabhängig von -i. Mit -i wird zusätzlich jeder erfolgreiche Schritt
// gemeldet und vor jedem Austausch eine Rückfrage gestellt.
func runJSON(configFile string, interactive bool) {

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

		if interactive && !confirm(entry.Old, entry.New) {
			reportEntryError(i, entry.Old, "vom Benutzer übersprungen")
			hadError = true
			continue
		}

		if err := update(entry.Old, entry.New); err != nil {
			reportEntryError(i, entry.Old, err.Error())
			hadError = true
			continue
		}

		if interactive {
			fmt.Printf("OK: %s\n", entry.Old)
		}
	}

	if hadError {
		os.Exit(exitPartial)
	}

	if interactive {
		fmt.Println("OK")
	}

	os.Exit(exitOK)
}

// reportEntryError meldet den Fehler eines einzelnen Batch-Eintrags auf
// stderr, ohne das Programm zu beenden.
func reportEntryError(index int, oldFile, msg string) {
	fmt.Fprintf(os.Stderr, "ERROR: Eintrag %d (%s): %s\n", index+1, oldFile, msg)
}

// confirm fragt interaktiv auf stdin nach, ob eine Datei ersetzt werden soll.
func confirm(oldFile, newFile string) bool {

	fmt.Printf("%s -> %s ersetzen? [j/N]: ", oldFile, newFile)

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))

	return line == "j" || line == "ja" || line == "y" || line == "yes"
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
	fmt.Println("  updater [-i] <alt> <neu>")
	fmt.Println("  updater [-i] -j <config.json>")
	fmt.Println()
	fmt.Println("  -i   interaktiv: fragt vor jedem Austausch nach und gibt")
	fmt.Println("       den Fortschritt aus. Ohne -i läuft das Tool still")
	fmt.Println("       (keine stdout-Ausgabe außer bei Fehlern).")
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
	fmt.Println("  8  vom Benutzer abgebrochen (nur -i, Single-Modus)")
	fmt.Println("  9  mindestens ein Eintrag im JSON-Modus fehlgeschlagen")
}

func errorExit(msg string, code int) {

	fmt.Fprintln(os.Stderr, "ERROR:", msg)

	os.Exit(code)
}
