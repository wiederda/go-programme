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
updater [-log] <alt> <neu>
```

| Parameter | Bedeutung                                   |
|-----------|------------------------------------------------|
| `<alt>`   | Pfad zur bestehenden Datei, die ersetzt wird |
| `<neu>`   | Pfad zur neuen Datei, die eingesetzt wird    |
| `-log`    | Optional: Fortschritt protokollieren (siehe unten) |

### Mehrere Dateien (JSON-Konfiguration)

```
updater [-log] -j <config.json>
updater [-log] --json <config.json>
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

**Wichtig:** Jeder Eintrag wird unabhängig von den anderen behandelt. Schlägt ein Eintrag fehl (z. B. weil eine Datei fehlt), wird der Batch **nicht** abgebrochen, sondern läuft mit den restlichen Einträgen weiter. Bereits erfolgreich ersetzte Dateien werden **nicht automatisch zurückgerollt** – das Tool bleibt hier bewusst einfach. Sollte ein Rollback nötig sein, muss dieser manuell erfolgen (anhand der `ERROR`- bzw. `OK: <datei>`-Zeilen, sofern mit `-log` protokolliert wurde).

### Protokoll-Ausgabe (`-log`)

Standardmäßig läuft `updater` **komplett still** – weder `stdout` noch `stderr`, nur der Exit Code zählt. Das ist gezielt auf unbeaufsichtigte Läufe (Cron, Skripte, Deployment-Pipelines) ausgelegt, bei denen sich ohnehin niemand die Ausgabe anschaut.

Mit dem optionalen Parameter `-log` wird zusätzlich protokolliert:

- Erfolgreich ersetzte Dateien: `OK: <datei>` je Datei, abschließend `OK`.
- Fehler: `ERROR: ...` auf `stderr`.

Ohne `-log` gibt es also **auch bei Fehlern keine Ausgabe** – nur der Exit Code signalisiert, ob (und im JSON-Modus grob welcher Art) ein Fehler aufgetreten ist.

`-log` kann sowohl im Einzeldatei- als auch im JSON-Modus vorangestellt werden, z. B. `updater -log a.txt b.txt` oder `updater -log -j config.json`.

Hilfe anzeigen:

```
updater -h
updater --help
```

## Exit Codes

| Code | Konstante         | Bedeutung                                          |
|------|--------------------|------------------------------------------------------|
| 0    | `exitOK`           | Erfolg (alle Dateien ersetzt)                       |
| 1    | `exitParam`        | Parameterfehler (falsche Anzahl/Art Argumente)      |
| 2    | `exitOldMissing`   | Alte Datei nicht gefunden                           |
| 3    | `exitNewMissing`   | Neue Datei nicht gefunden                           |
| 4    | `exitBackup`       | Fehler beim Erstellen oder Löschen des Backups      |
| 5    | `exitReplace`      | Fehler beim Einsetzen der neuen Datei               |
| 6    | `exitPermission`   | Fehler beim Setzen der Rechte oder bei der Prüfung  |
| 7    | `exitJSON`         | Konfigurationsdatei fehlt, ungültig oder unvollständig |
| 9    | `exitPartial`      | JSON-Modus: mindestens ein Eintrag ist fehlgeschlagen |

**Wichtig:** Die Codes `2`–`6` sind als konkreter Prozess-Exit-Code nur im **Einzeldatei-Modus** relevant. Im **JSON-Modus** ist der Prozess-Exit-Code entweder `0` (alle Einträge erfolgreich) oder `9` (mindestens ein Eintrag fehlgeschlagen) – welcher konkrete Fehlertyp (`2`–`7`) bei welchem Eintrag aufgetreten ist, steht in der jeweiligen `stderr`-Zeile, nicht im Exit Code.

Fehlermeldungen werden nur mit `-log` auf `stderr` im Format `ERROR: <meldung>` ausgegeben; ohne `-log` gibt es dazu keine Ausgabe. Im JSON-Modus enthält jede Zeile zusätzlich die Nummer und den Alt-Pfad des betroffenen Eintrags, z. B.:

```
ERROR: Eintrag 2 (C:\Programme\App\b.dll): neue Datei nicht gefunden: C:\Update\b.dll
```

## Beispiele

Einzeldatei, still (Standard):

```
$ updater config.yaml config.yaml.new
$ echo $?
0
```

Einzeldatei, mit Protokoll (`-log`):

```
$ updater -log config.yaml config.yaml.new
OK
```

Mehrere Dateien, mit Protokoll (`-log`):

```
$ updater -log -j updates.json
OK: C:\Programme\App\a.exe
OK: C:\Programme\App\b.dll
OK
```

Ohne `-log` gibt es bei Erfolg keine Ausgabe – nur der Exit Code (`0`) signalisiert das Ergebnis. Mit `-log` wird zusätzlich jede erfolgreich ersetzte Datei einzeln mit `OK: <alt>` protokolliert, sodass im Fehlerfall nachvollziehbar ist, welche Dateien bereits ersetzt wurden.

Mehrere Dateien mit einem fehlerhaften Eintrag, ohne `-log` (komplett still, nur Exit Code zählt):

```
$ updater -j updates.json
$ echo $?
9
```

Dasselbe mit `-log` – jetzt ist nachvollziehbar, welcher Eintrag fehlgeschlagen ist und welche trotzdem erfolgreich waren:

```
$ updater -log -j updates.json
OK: C:\Programme\App\a.exe
ERROR: Eintrag 2 (C:\Programme\App\b.dll): neue Datei nicht gefunden: C:\Update\b.dll
$ echo $?
9
```

Eintrag 1 wurde hier trotz des Fehlers bei Eintrag 2 ganz normal ersetzt – nur der fehlerhafte Eintrag wird gemeldet, der Rest läuft weiter.

## Sicherheitsaspekte

- **Kein Datenverlust:** Solange die alte Datei existierte, bleibt durch das Backup (`<alt>.old`) sowohl im Fehlerfall als auch nach einem erfolgreichen Lauf eine wiederherstellbare Kopie erhalten – bis zum nächsten Aufruf für dieselbe Datei.
- **Kein automatisches Rollback:** Weder pro Datei noch über mehrere Dateien hinweg (JSON-Modus) wird bei einem Fehler automatisch etwas zurückgerollt. Das Tool bricht ab und meldet den Fehler; eine Wiederherstellung aus dem Backup ist bewusst Sache des Anwenders, um das Tool einfach zu halten.
- **Rechteübernahme:** Die Zugriffsrechte (Permissions) der alten Datei werden auf die neue Datei übertragen, damit sich z. B. ausführbare Dateien oder restriktiv gesetzte Konfigurationsdateien nach dem Austausch identisch verhalten.
- **Kein Schutz vor gleichzeitigem Zugriff:** Das Tool ist nicht auf parallele Aufrufe gegen dieselbe Datei ausgelegt (keine Lock-Datei, keine Mutex-Logik).
- **Plattformübergreifend:** Das Tool funktioniert auf Windows, Linux und macOS.
