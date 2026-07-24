# FastCopy ⚡

> Schnelles, parallelisiertes Dateisynchronisations-Tool auf Basis von Go – mit SHA-256-Inhaltsvergleich und optionalem Mirror-Modus.

---

## Inhaltsverzeichnis

- [Überblick](#überblick)
- [Features](#features)
- [Architektur & Pipeline](#architektur--pipeline)
- [Installation](#installation)
- [Verwendung](#verwendung)
- [Flags & Optionen](#flags--optionen)
- [Beispiele](#beispiele)
- [Funktionsweise im Detail](#funktionsweise-im-detail)
- [Fehlerbehandlung](#fehlerbehandlung)
- [Bekannte Einschränkungen](#bekannte-einschränkungen)
- [Lizenz](#lizenz)

---

## Überblick

**FastCopy** ist ein kommandozeilenbasiertes Synchronisations-Tool, das Dateien von einem Quellverzeichnis in ein Zielverzeichnis kopiert – intelligent, schnell und zuverlässig.

Anstatt blind alle Dateien zu überschreiben, analysiert FastCopy jede Datei in drei Stufen:

1. **Größenvergleich** – Unterschiedliche Dateigröße → sofort kopieren
2. **Hash-Vergleich** (SHA-256) – Gleiche Größe, aber unterschiedlicher Inhalt → kopieren
3. **Identisch** – Gleiche Größe und gleicher Hash → überspringen

Die gesamte Pipeline ist vollständig parallelisiert und nutzt alle verfügbaren CPU-Kerne.

---

## Features

- **Mehrstufige Filterung** – Vermeidet unnötige Kopiervorgänge durch Größen- und Hash-Vergleich
- **Parallele Verarbeitung** – Worker-Pools für Hash-Berechnung und Dateikopiervorgänge (skaliert mit CPU-Kernen)
- **Mirror-Modus** – Löscht Dateien und Ordner im Ziel, die nicht mehr in der Quelle existieren
- **SHA-256-Inhaltsprüfung** – Erkennt Änderungen auch bei unveränderter Dateigröße
- **Metadaten-Erhaltung** – Überträgt Dateiberechtigungen (`chmod`) und Zeitstempel (`mtime`)
- **Leere Verzeichnisse** – Optionales Kopieren von leeren Ordnern
- **Verbose-Modus** – Detaillierte Protokollierung aller übersprungenen Dateien
- **Plattformübergreifend** – Läuft auf Linux, macOS und Windows

---

## Architektur & Pipeline

FastCopy verarbeitet Dateien in einer 5-phasigen Pipeline:

```
┌──────────────────────────────────────────────────────┐
│  PHASE 1 – Scan & Filter                             │
│  filepath.Walk() über Quellverzeichnis               │
│  → Entscheidung: direkt kopieren ODER Hash prüfen    │
└───────────────────┬──────────────────────────────────┘
                    │
        ┌───────────┴────────────┐
        ▼                        ▼
  tasksToCopy              tasksToHash
        │                        │
        │              ┌─────────▼──────────┐
        │              │  PHASE 2 – Hashing  │
        │              │  N Worker-Goroutinen│
        │              │  (runtime.NumCPU()) │
        │              └────────┬────────────┘
        │                       │
        │          Hash gleich? ├──► resultChan (Übersprungen)
        │                       │
        │          Hash ≠?      └──► copyTaskChan
        │                              │
        └──────────────────────────────┘
                                       │
                          ┌────────────▼────────────┐
                          │  PHASE 3 – Copy Workers  │
                          │  N Worker-Goroutinen     │
                          └────────────┬─────────────┘
                                       │
                          ┌────────────▼─────────────┐
                          │  PHASE 4 – Aggregator     │
                          │  Zählt Ergebnisse,        │
                          │  gibt Fehler aus          │
                          └────────────┬──────────────┘
                                       │
                          ┌────────────▼──────────────┐
                          │  PHASE 5 (optional) – MIR  │
                          │  Löscht verwaiste Dateien  │
                          └────────────────────────────┘
```

Die Phasen 2 und 3 laufen **gleichzeitig** über Go-Channels: Sobald Phase 2 einen Hash-Unterschied erkennt, wird die Datei direkt in den Copy-Worker-Pool eingereiht, ohne auf den Abschluss aller Hash-Berechnungen zu warten.

---

## Installation

### Voraussetzungen

- [Go](https://go.dev/dl/) 1.18 oder neuer

### Aus dem Quellcode bauen

```bash
# Repository klonen
git clone https://github.com/beispiel/fastcopy.git
cd fastcopy

# Binary kompilieren
go build -o fastcopy .

# (Optional) Systemweit installieren
sudo mv fastcopy /usr/local/bin/
```

### Direktausführung (ohne Build)

```bash
go run main.go -src /pfad/quelle -dest /pfad/ziel
```

---

## Verwendung

```
fastcopy -src <QUELLE> -dest <ZIEL> [OPTIONEN]
```

### Minimales Beispiel

```bash
# Linux / macOS
fastcopy -src /home/user/dokumente -dest /backup/dokumente

# Windows – Pfade direkt ohne Escaping angeben
fastcopy -src C:\Users\Max\Dokumente -dest D:\Backup\Dokumente
fastcopy -src C:\Users\Max\Projekte -dest \\NAS\Backup\Projekte
```

> **Windows-Hinweis:** FastCopy normalisiert Pfade intern mit `filepath.Clean`. Backslashes müssen **nicht** escaped werden – `C:\Users\Max` funktioniert direkt in der Kommandozeile (cmd.exe und PowerShell).

---

## Flags & Optionen

| Flag         | Typ    | Standard | Beschreibung                                                                                   |
|--------------|--------|----------|-----------------------------------------------------------------------------------------------|
| `-src`       | string | –        | **Pflichtfeld.** Pfad zum Quellverzeichnis.                                                   |
| `-dest`      | string | –        | **Pflichtfeld.** Pfad zum Zielverzeichnis. Wird automatisch erstellt, falls nicht vorhanden.  |
| `-MIR`       | bool   | `false`  | Mirror-Modus: Löscht Dateien/Ordner im Ziel, die in der Quelle nicht (mehr) existieren.      |
| `-E`         | bool   | `false`  | Kopiert auch leere Verzeichnisse in das Ziel.                                                 |
| `-NO-HASH`   | bool   | `false`  | Überspringt den SHA-256-Vergleich. Dateien mit gleicher Größe werden nie kopiert.             |
| `-v`         | bool   | `false`  | Verbose-Modus: Gibt übersprungene Dateien mit Begründung aus.                                 |

> **Hinweis zu `-NO-HASH`:** Dieser Modus ist deutlich schneller, erkennt aber keine inhaltlichen Änderungen bei gleichbleibender Dateigröße (z. B. bei in-place editierten Binärdateien oder bestimmten Log-Dateien).

---

## Beispiele

### Einfache Sicherungskopie

```bash
fastcopy -src /home/user/projekte -dest /mnt/nas/backup/projekte
```

### Lokale Disk auf NAS synchronisieren (Cross-Device)

FastCopy erkennt automatisch, wenn Quelle und Ziel auf verschiedenen Dateisystemen liegen, und wechselt zum Kopier-Fallback.

```bash
# Linux: lokale SSD → gemountetes NAS
fastcopy -src /home/user/daten -dest /mnt/nas/backup -MIR

# Windows: lokales Laufwerk → Netzwerkshare
fastcopy -src C:\Users\Max\Daten -dest \\NAS\Backup\Max -MIR
```

### Mirror: Ziel exakt wie Quelle halten

Löscht im Ziel alle Dateien und Ordner, die in der Quelle nicht mehr vorhanden sind.

```bash
fastcopy -src /var/www/html -dest /backup/www -MIR
```

### Schnell-Sync ohne Hash (nur Größenvergleich)

Ideal für große Medienbibliotheken, bei denen Inhaltsänderungen unwahrscheinlich sind.

```bash
fastcopy -src /media/fotos -dest /backup/fotos -NO-HASH
```

### Vollständige Synchronisation mit Protokollierung

```bash
fastcopy -src /home/user -dest /backup/home -MIR -E -v
```

### Leere Ordnerstruktur übertragen

```bash
fastcopy -src /templates/struktur -dest /projekte/neu -E -NO-HASH
```

---

## Funktionsweise im Detail

### Phase 1 – Scan & Filter

`filepath.Clean()` normalisiert Eingabepfade vor dem Start (vereinheitlicht Separatoren, entfernt trailing Slashes, behandelt Windows-Laufwerksbuchstaben korrekt). Anschließend traversiert `filepath.Walk()` das Quellverzeichnis rekursiv. Für jede Datei wird entschieden:

- **Zieldatei existiert nicht** → direkt in `tasksToCopy` einreihen (Grund: `"Neu im Ziel"`)
- **Zieldatei nicht lesbar** (Berechtigungsfehler etc.) → als neu behandeln, Warnung ausgeben
- **Dateigröße unterschiedlich** → direkt in `tasksToCopy` einreihen (Grund: `"Größe unterschiedlich"`)
- **Dateigröße gleich, Hash aktiv** → in `tasksToHash` einreihen
- **Dateigröße gleich, `-NO-HASH` gesetzt** → überspringen
- **Pfad > 260 Zeichen auf Windows** → Warnung, überspringen (kein Panic)

### Phase 2 – Gestufte Hash-Berechnung

Für jede Datei in `tasksToHash` wird eine **zweistufige Hash-Strategie** angewendet:

**Kleine Dateien (< 4 MB):** direkter Full-SHA-256 beider Dateien.

**Große Dateien (≥ 4 MB):**
1. **Partial-Hash** – liest jeweils 64 KB vom Anfang und Ende beider Dateien + Dateigröße als Seed → SHA-256
2. Partial-Hashes **unterschiedlich** → sofort als geändert markieren, Full-Hash wird gespart
3. Partial-Hashes **gleich** → Full-SHA-256 als endgültige Bestätigung

```
Datei ≥ 4 MB
     │
     ▼
Partial-Hash (128 KB gelesen)
     │
     ├── verschieden → kopieren  (Full-Hash eingespart ✓)
     │
     └── gleich
          │
          ▼
     Full-Hash (gesamte Datei)
          │
          ├── verschieden → kopieren
          └── gleich → überspringen
```

Praktischer Effekt: Dateien, die sich am Anfang oder Ende unterscheiden (Logs, Videos, Datenbank-Dumps), werden mit nur ~128 KB I/O erkannt statt mit einem vollen Read.

### Phase 3 – Atomarer Kopier-Worker-Pool

`N = max(runtime.NumCPU(), 4)` parallele Goroutinen kopieren Dateien **crash-sicher**:

1. Temporäre Datei `<zieldatei>.tmp.fastcopy` im **selben Verzeichnis** anlegen
2. Inhalt via `io.Copy()` schreiben
3. `Sync()` – Flush auf Disk (Fehler auf Netzlaufwerken werden als Warnung behandelt)
4. Metadaten auf die temporäre Datei setzen (`Chtimes`, `Chmod`)
5. `os.Rename()` – atomares Umbenennen zur endgültigen Zieldatei

**Cross-Device-Fallback:** Liegt das Ziel auf einem anderen Dateisystem als die temporäre Datei (z. B. lokale SSD → NAS-Share), schlägt `os.Rename()` mit `invalid cross-device link` fehl. FastCopy erkennt diesen Fehler automatisch und wechselt zu `crossDeviceMove()`: direktes Kopieren in die Zieldatei + Löschen der Tmp-Datei. Nicht atomar, aber der einzig mögliche Weg über Dateisystemgrenzen.

> **Plattform-Hinweis:** `os.Rename()` ist auf POSIX-Systemen (Linux, macOS) garantiert atomar. Auf Windows seit Go 1.21 ebenfalls (`MoveFileExW` mit `MOVEFILE_REPLACE_EXISTING`).

### Phase 4 – Ergebnis-Aggregator

Eine dedizierte Goroutine liest alle `CopyResult`-Einträge vom Channel und zählt erfolgreich kopierte Dateien, übersprungene Dateien und Fehler. Pfade werden via `filepath.Rel()` relativ dargestellt – korrekt auf allen Plattformen, auch Windows mit gemischten Separatoren.

### Phase 5 – Mirror-Bereinigung (optional)

Falls `-MIR` gesetzt ist, wird das Zielverzeichnis nach Abschluss der Synchronisation traversiert. Alle Pfade, die **nicht** in der während Phase 1 gesammelten `targetPaths`-Map enthalten sind, werden gelöscht. `.tmp.fastcopy`-Dateien werden dabei explizit übersprungen.

- Verzeichnisse: `os.RemoveAll()` + `filepath.SkipDir`
- Dateien: `os.Remove()`

---

## Fehlerbehandlung

| Situation                                    | Verhalten                                                                          |
|----------------------------------------------|------------------------------------------------------------------------------------|
| Quell- oder Zieldatei nicht zugänglich       | `[WARNUNG]` im Scan, Datei wird übersprungen – kein Programmabbruch               |
| `os.Stat` der Zieldatei schlägt fehl         | Datei wird als neu behandelt und kopiert, `[WARNUNG]` wird ausgegeben              |
| Hash-Berechnung schlägt fehl                 | Fehler gemeldet, Datei wird **nicht** kopiert                                      |
| Seek nicht unterstützt (Netzlaufwerk)        | Graceful Fallback auf Full-Hash                                                    |
| `Sync()` nicht unterstützt (SMB/NFS/exFAT)  | `[WARNUNG]`, Kopiervorgang gilt trotzdem als erfolgreich                           |
| Cross-Device `Rename()` schlägt fehl         | Automatischer Fallback auf `crossDeviceMove()` (Kopieren + Löschen)               |
| Kopiervorgang schlägt fehl                   | `[FEHLER]` gemeldet, Programm läuft weiter                                         |
| Zeitstempel/Chmod nicht setzbar              | `[WARNUNG]`, Datei gilt als erfolgreich kopiert                                    |
| Pfad > 260 Zeichen (Windows)                 | `[WARNUNG]`, Datei/Ordner wird übersprungen                                        |
| Zielverzeichnis nicht erstellbar             | Fataler Fehler, Exit-Code 1                                                        |
| `-src` oder `-dest` nicht angegeben          | Fehlerausgabe + Usage-Hilfe, Exit-Code 1                                           |
| Quelle = Ziel                                | Fehlerausgabe, Exit-Code 1                                                         |
| Ziel ist Unterverzeichnis der Quelle         | Fehlerausgabe, Exit-Code 1 (verhindert rekursive Selbstkopie)                      |

---

## Bekannte Einschränkungen

- **Symlinks** werden von `filepath.Walk()` nicht als Symlinks verarbeitet – das Ziel des Links wird traversiert. Symbolische Links selbst werden nicht als solche ins Ziel übertragen.
- **Zugriffszeiten (`atime`)** werden auf den Wert der Änderungszeit (`mtime`) gesetzt, da eine separate `atime` aus `os.FileInfo` unter Go nicht portabel verfügbar ist.
- **Windows Long Paths (> 260 Zeichen):** Pfade über dem Limit werden übersprungen und gewarnt. Zur vollständigen Unterstützung muss in der Windows Registry `HKLM\SYSTEM\CurrentControlSet\Control\FileSystem\LongPathsEnabled = 1` gesetzt sein.
- **Cross-Device-Kopien** sind nicht atomar. Ein Absturz während `crossDeviceMove()` kann eine unvollständige Zieldatei hinterlassen. Auf demselben Dateisystem ist die atomare Rename-Strategie davon nicht betroffen.
- **Partielle Hash-Kollisionen**: Dateien, die sich ausschließlich in der Mitte unterscheiden, haben denselben Partial-Hash. Der nachgelagerte Full-Hash erkennt die Änderung korrekt – der Partial-Hash ist nur eine Vorstufe.
- **Temporäre Dateien**: Bei hartem Abbruch (SIGKILL, Stromausfall) können `.tmp.fastcopy`-Dateien verbleiben. FastCopy räumt diese beim nächsten Start automatisch auf.