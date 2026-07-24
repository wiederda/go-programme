# updater

Ein kleines Kommandozeilen-Tool zum atomaren Austausch einer oder mehrerer Dateien mit automatischem Backup und Rollback bei Fehlern.

## Funktionsweise

`updater` ersetzt eine bestehende Datei (`<alt>`) durch eine neue Datei (`<neu>`). Dabei wird die alte Datei zunächst als Backup gesichert (`<alt>.old`), bevor der eigentliche Austausch stattfindet. Schlägt ein Schritt fehl, wird versucht, den vorherigen Zustand dieser einen Datei wiederherzustellen (Rollback).

Der Ablauf pro Datei im Detail:

1. Prüfen, ob die alte Datei existiert.
2. Prüfen, ob die neue Datei existiert.
3. Die aktuellen Zugriffsrechte (Permissions) der alten Datei merken.
4. Ein eventuell vorhandenes altes Backup (`<alt>.old`) löschen.
5. Die alte Datei zu `<alt>.old` umbenennen (Backup).
6. Die neue Datei an die Stelle der alten Datei umbenennen (Austausch).
7. Die ursprünglichen Zugriffsrechte auf die neue Datei übertragen.
8. Eine abschließende Prüfung durchführen (Datei vorhanden, Rechte korrekt).

Schritte 6–8 enthalten ein Rollback: Falls der Austausch, das Setzen der Rechte oder die abschließende Prüfung fehlschlägt, wird versucht, die alte Datei aus dem Backup wiederherzustellen. Dieses Rollback gilt immer nur für die gerade bearbeitete Einzeldatei.

**Das Backup (`<alt>.old`) wird nach einem erfolgreichen Lauf bewusst nicht gelöscht.** Es bleibt liegen und dient als manuelle Rollback-Möglichkeit. Erst beim nächsten Aufruf für dieselbe Datei wird es automatisch entfernt (Schritt 4), bevor ein neues Backup angelegt wird.

## Aufruf

### Einzeldatei

```
updater <alt> <neu>
```

| Parameter | Bedeutung                                   |
|-----------|------------------------------------------------|
| `<alt>`   | Pfad zur bestehenden Datei, die ersetzt wird |
| `<neu>`   | Pfad zur neuen Datei, die eingesetzt wird    |

### Mehrere Dateien (JSON-Konfiguration)

```
updater -j <config.json>
updater --json <config.json>
```

Format der Konfigurationsdatei:

```json
{
  "files": [
    { "old": "C:\\Programme\\App\\a.exe", "new": "C:\\Update\\a.exe" },
    { "old": "C:\\Programme\\App\\b.dll", "new": "C:\\Update\\b.dll" }
  ]
}
```

Jeder Eintrag benötigt `old` und `new`. Die Einträge werden **nacheinander** abgearbeitet, jede einzelne Datei durchläuft dabei genau den oben beschriebenen Ablauf.

**Wichtig:** Schlägt ein Eintrag fehl, wird sofort abgebrochen. Bereits erfolgreich ersetzte Dateien aus vorherigen Einträgen werden **nicht automatisch zurückgerollt** – das Tool bleibt hier bewusst einfach. Sollte in diesem Fall ein Rollback nötig sein, muss dieser manuell erfolgen (z. B. anhand von Log-Ausgaben, welche Dateien bereits ersetzt wurden).

Hilfe anzeigen:

```
updater -h
updater --help
```

## Exit Codes

| Code | Konstante         | Bedeutung                                          |
|------|--------------------|------------------------------------------------------|
| 0    | `exitOK`           | Erfolg                                              |
| 1    | `exitParam`        | Parameterfehler (falsche Anzahl/Art Argumente)      |
| 2    | `exitOldMissing`   | Alte Datei nicht gefunden                           |
| 3    | `exitNewMissing`   | Neue Datei nicht gefunden                           |
| 4    | `exitBackup`       | Fehler beim Erstellen oder Löschen des Backups      |
| 5    | `exitReplace`      | Fehler beim Einsetzen der neuen Datei               |
| 6    | `exitPermission`   | Fehler beim Setzen der Rechte oder bei der Prüfung  |
| 7    | `exitJSON`         | Konfigurationsdatei fehlt, ungültig oder unvollständig |

Fehlermeldungen werden auf `stderr` im Format `ERROR: <meldung>` ausgegeben, der Exit Code entspricht der Tabelle oben. Im JSON-Modus enthält die Meldung zusätzlich die Nummer und den Alt-Pfad des betroffenen Eintrags.

## Beispiele

Einzeldatei:

```
$ updater config.yaml config.yaml.new
OK
```

Mehrere Dateien:

```
$ updater -j updates.json
OK: C:\Programme\App\a.exe
OK: C:\Programme\App\b.dll
OK
```

Bei Erfolg wird am Ende `OK` auf `stdout` ausgegeben und der Prozess endet mit Exit Code `0`. Im JSON-Modus wird zusätzlich jede erfolgreich ersetzte Datei einzeln mit `OK: <alt>` protokolliert, sodass im Fehlerfall nachvollziehbar ist, welche Dateien bereits ersetzt wurden.

## Sicherheitsaspekte

- **Kein Datenverlust:** Solange die alte Datei existierte, bleibt durch das Backup (`<alt>.old`) sowohl im Fehlerfall als auch nach einem erfolgreichen Lauf eine wiederherstellbare Kopie erhalten – bis zum nächsten Aufruf für dieselbe Datei.
- **Rollback nur pro Datei:** Schlägt der Austausch, das Setzen der Rechte oder die Prüfung einer Datei fehl, versucht `updater` automatisch, genau diese Datei aus ihrem Backup wiederherzustellen. Der Rückgabewert des Rollbacks selbst wird nicht geprüft.
- **Kein Rollback über mehrere Dateien hinweg:** Im JSON-Modus werden bereits erfolgreich abgeschlossene Einträge bei einem späteren Fehler nicht automatisch zurückgerollt. Dies ist eine bewusste Design-Entscheidung, um das Tool einfach zu halten.
- **Rechteübernahme:** Die Zugriffsrechte (Permissions) der alten Datei werden auf die neue Datei übertragen, damit sich z. B. ausführbare Dateien oder restriktiv gesetzte Konfigurationsdateien nach dem Austausch identisch verhalten.
- **Kein Schutz vor gleichzeitigem Zugriff:** Das Tool ist nicht auf parallele Aufrufe gegen dieselbe Datei ausgelegt (keine Lock-Datei, keine Mutex-Logik).
- **Plattformübergreifend:** Das Tool funktioniert auf Windows.