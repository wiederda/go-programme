# File Scanner

Ein Kommandozeilenwerkzeug zum Scannen lokaler oder Netzwerkverzeichnisse und zum Synchronisieren von Dateiinformationen mit einer PostgreSQL-Datenbank.

Für jede gefundene Datei werden folgende Informationen gespeichert:

* relativer Dateipfad
* Dateigröße
* SHA-256-Hash
* zugehöriges Verzeichnis

Nicht mehr vorhandene Dateien werden als gelöscht markiert, anstatt aus der Datenbank entfernt zu werden.

## Funktionen

* Scannen lokaler oder gemounteter Netzwerkpfade
* Filterung nach Dateiendungen
* Berechnung von SHA-256-Hashes mit `sha256-simd`
* Automatische Aktualisierung bestehender Einträge
* Soft-Delete über das Feld `geloescht`
* Plattformunabhängige Pfadverarbeitung

## Voraussetzungen

* Go 1.24 oder neuer
* PostgreSQL 14 oder neuer

## Installation

Repository klonen:

```bash
git clone <repository-url>
cd <repository-name>
```

Abhängigkeiten installieren:

```bash
go mod download
```

Anwendung bauen:

```bash
go build -o filescanner .
```

## Konfiguration

Die Anwendung erwartet eine `config.json` im Arbeitsverzeichnis.

Beispiel:

```json
{
  "dsn": "postgres://user:password@localhost:5432/media?sslmode=disable",
  "include_suffix": [
    ".mp3",
    ".flac",
    ".wav"
  ]
}
```

### Konfigurationsparameter

| Feld             | Beschreibung                                                            |
| ---------------- | ----------------------------------------------------------------------- |
| `dsn`            | PostgreSQL-Verbindungszeichenfolge                                      |
| `include_suffix` | Liste erlaubter Dateiendungen. Leer lassen, um alle Dateien zu scannen. |

## Datenbankschema

### Tabelle `verzeichnisse`

Enthält die zu scannenden Verzeichnisse.

```sql
CREATE TABLE verzeichnisse (
    id SERIAL PRIMARY KEY,
    local TEXT NOT NULL,
    network TEXT NOT NULL,
    relativ TEXT NOT NULL
);
```

| Spalte    | Beschreibung        |
| --------- | ------------------- |
| `id`      | Eindeutige ID       |
| `local`   | Lokaler Basispfad   |
| `network` | Netzwerkpfad        |
| `relativ` | Relativer Unterpfad |

### Tabelle `dateien`

Enthält die erfassten Dateiinformationen.

```sql
CREATE TABLE dateien (
    verzeichnis_id INTEGER NOT NULL REFERENCES verzeichnisse(id),
    dateiname TEXT NOT NULL,
    groesse BIGINT NOT NULL,
    hash TEXT NOT NULL,
    geloescht BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (verzeichnis_id, dateiname)
);
```

## Verwendung

### Lokale Verzeichnisse scannen

```bash
./filescanner --mode local
```

### Netzwerkpfade scannen

```bash
./filescanner --mode network
```

Mögliche Werte für `--mode`:

* `local`
* `network`

Standardwert:

```text
local
```

## Beispiel

Angenommen, die Tabelle `verzeichnisse` enthält:

| id | local    | network    | relativ |
| -- | -------- | ---------- | ------- |
| 1  | `/media` | `Z:\Media` | `Musik` |

Im lokalen Modus wird folgendes Verzeichnis gescannt:

```text
/media/Musik
```

Eine Datei:

```text
/media/Musik/Rock/album/song.mp3
```

wird in der Datenbank gespeichert als:

```text
Rock/album/song.mp3
```

## Ablauf

1. Verzeichnisse aus der Tabelle `verzeichnisse` laden
2. Zielpfad anhand des gewählten Modus bestimmen
3. Dateien rekursiv durchsuchen
4. Dateigröße und SHA-256-Hash berechnen
5. Relative Pfade erzeugen
6. Datensätze einfügen oder aktualisieren
7. Nicht mehr vorhandene Dateien als gelöscht markieren

## Bekannte Einschränkungen

* Dateien werden seriell verarbeitet.
* Für jede Datei wird der vollständige Hash neu berechnet.
* Große Verzeichnisstrukturen können lange Laufzeiten verursachen.
* Änderungen an Dateinamen werden als „gelöscht + neu angelegt“ erkannt.
* Netzwerkpfade müssen vor dem Scan erreichbar und eingebunden sein.

## Mögliche Erweiterungen

* Parallele Hash-Berechnung
* Fortschrittsanzeige
* Wiederaufnahme unterbrochener Scans
* Ausschlusslisten für Verzeichnisse
* Zeitstempel für letzte Änderungen
* Inkrementelle Scans anhand von Dateigröße und Änderungsdatum
* Konfigurierbares Logging
* Graceful Shutdown über Signale
