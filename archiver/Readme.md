# Archiver

Ein einfaches CLI-Tool zum Erstellen und Entpacken von Archiven (`zip`, `tar`, `tgz`).

---

## Features

- ZIP / TAR / TGZ Unterstützung
- Streaming (kein komplettes Laden in den Speicher)
- Abbruch mit `Ctrl+C`
- Windows-Pfad Support (`C:\...`)
- Automatische Zielordner-Erstellung beim Entpacken
- Automatischer Zielordner beim Entpacken basiert auf Archivnamen
- Sichere Pfadbehandlung (kein Path Traversal)
- Kein externes Tool erforderlich

---

## Installation

```bash
go build -o archiver
````

---

## Nutzung

### Hilfe anzeigen

```bash
archiver -h
```

---

## Archiv erstellen

```bash
archiver create zip c:\data backup.zip
archiver create tar c:\data backup.tar
archiver create tgz c:\data backup.tgz
```

### Verhalten

* Wenn nur ein Dateiname angegeben wird (`backup.zip`), wird im aktuellen Arbeitsverzeichnis gespeichert
* Wenn der Zielpfad absolut ist, wird er direkt verwendet
* Falls der Source-Pfad dem aktuellen Arbeitsverzeichnis entspricht, wird automatisch ein sicherer Ausweichpfad verwendet

---

## Archiv entpacken

```bash
archiver extract zip backup.zip c:\out
archiver extract tar backup.tar c:\out
archiver extract tgz backup.tgz c:\out
```

### Verhalten

Beim Entpacken wird automatisch ein Unterordner erstellt:

```text
c:\out\<archivname>\
```

Beispiele:

```
backup.zip   → c:\out\backup\
logs.tgz     → c:\out\logs\
data.tar     → c:\out\data\
```

---

## Beispiele

### Create

```bash
cd C:\_Lokale_Daten_ungesichert
archiver create zip c:\test test.zip
```

→ erstellt:

```
C:\_Lokale_Daten_ungesichert\test.zip
```

---

### Extract

```bash
archiver extract zip backup.zip c:\out
```

→ entpackt nach:

```
c:\out\backup\
```

---

## Sicherheitsverhalten

* Pfad-Traversal wird blockiert (`../` außerhalb Zielverzeichnis)
* Keine Schreiboperation außerhalb des definierten Zielbaums
* Alle Zielordner werden automatisch erstellt

---

## Verhalten bei Ctrl+C

* laufende Operation wird sauber abgebrochen
* keine halbfertigen Dateien werden garantiert abgeschlossen (so weit Streaming es erlaubt)

---

## Unterstützte Formate

| Format | Lesen | Schreiben |
| ------ | ----- | --------- |
| zip    | ✔     | ✔         |
| tar    | ✔     | ✔         |
| tgz    | ✔     | ✔         |

---

## Designprinzip

* Deterministisches Verhalten
* Kein verstecktes „Magie-Verzeichnis“
* Archivname bestimmt Standard-Entpackziel
* Arbeitsverzeichnis bestimmt Standard-Output beim Erstellen

```
