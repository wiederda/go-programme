# updater

Ein kleines Kommandozeilen-Tool zum atomaren Austausch einer oder mehrerer Dateien mit automatischem Backup. Ein Rollback erfolgt nicht automatisch, sondern liegt in der Verantwortung des Anwenders.

## Funktionsweise

`updater` ersetzt eine bestehende Datei (`<alt>`) durch eine neue Datei (`<neu>`). Dabei wird die alte Datei zunächst als Backup gesichert (`<alt>.old`), bevor der eigentliche Austausch stattfindet. Schlägt ein Schritt danach fehl, wird **kein automatisches Rollback** durchgeführt - der aktuelle Zustand bleibt einfach bestehen, und das Backup steht für eine manuelle Wiederherstellung zur Verfügung.

Der Ablauf pro Datei im Detail:

1. Prüfen, ob die alte Datei existiert.
2. Prüfen, ob die neue Datei existiert.
3. Die aktuellen Zugriffsrechte (Permissions) der alten Datei merken.
4. Ein eventuell vorhandenes altes Backup (`<alt>.old`) löschen.
5. Die alte Datei zu `<alt>.old` umbenennen (Backup).
6. Die neue Datei an die Stelle der alten Datei umbenennen (Austausch).
7. Die ursprünglichen Zugriffsrechte auf die neue Datei übertragen.
8. Eine abschließende Prüfung durchführen (Datei vorhanden, Rechte korrekt).

Schlägt einer der Schritte 6–8 fehl, bricht `updater` mit einer Fehlermeldung und passendem Exit Code ab. Es wird **nicht** versucht, die alte Datei automatisch aus dem Backup wiederherzustellen - das muss der Anwender bei Bedarf selbst tun.

**Das Backup (`<alt>.old`) wird sowohl im Erfolgs- als auch im Fehlerfall nicht sofort gelöscht.** Es bleibt liegen und dient als manuelle Rollback-Möglichkeit. Erst beim nächsten Aufruf für dieselbe Datei wird es automatisch entfernt (Schritt 4), bevor ein neues Backup angelegt wird.

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
- **Kein automatisches Rollback:** Weder pro Datei noch über mehrere Dateien hinweg (JSON-Modus) wird bei einem Fehler automatisch etwas zurückgerollt. Das Tool bricht ab und meldet den Fehler; eine Wiederherstellung aus dem Backup ist bewusst Sache des Anwenders, um das Tool einfach zu halten.
- **Rechteübernahme:** Die Zugriffsrechte (Permissions) der alten Datei werden auf die neue Datei übertragen, damit sich z. B. ausführbare Dateien oder restriktiv gesetzte Konfigurationsdateien nach dem Austausch identisch verhalten.
- **Kein Schutz vor gleichzeitigem Zugriff:** Das Tool ist nicht auf parallele Aufrufe gegen dieselbe Datei ausgelegt (keine Lock-Datei, keine Mutex-Logik).
- **Plattformübergreifend:** Das Tool funktioniert auf Windows, Linux und macOS.