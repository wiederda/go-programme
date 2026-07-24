# Ordner-Cleanup Tool

Ein einfaches Kommandozeilenprogramm zum automatischen Finden und optionalen Löschen veralteter Ordner.

## Funktionen

* Durchsucht rekursiv einen Basisordner.
* Berücksichtigt nur Ordner, deren Name ein `_` enthält.
* Ignoriert Ordner, die mit `_` beginnen.
* Löscht nur Ordner, deren letztes Änderungsdatum älter als eine definierte Anzahl von Tagen ist.
* Unterstützt lokale Pfade und UNC-Netzwerkpfade.
* Standardmäßig wird ein Testlauf (`dry-run`) ausgeführt.
* Optionale Protokollierung in eine Logdatei.

---

## Voraussetzungen

* Go 1.20 oder neuer

---

## Installation

Repository klonen und kompilieren:

```bash
git clone <repository-url>
cd <repository-name>

go build -o folder-cleanup .
```

Unter Windows:

```powershell
go build -o folder-cleanup.exe .
```

---

## Verwendung

```text
folder-cleanup [Optionen]
```

### Optionen

| Option     | Standardwert | Beschreibung                                                 |
| ---------- | ------------ | ------------------------------------------------------------ |
| `-path`    | `.`          | Basis-Pfad zum Durchsuchen                                   |
| `-days`    | `30`         | Ordner älter als diese Anzahl an Tagen werden berücksichtigt |
| `-dry-run` | `true`       | Zeigt nur an, welche Ordner gelöscht würden                  |
| `-log`     | leer         | Optionaler Pfad zur Logdatei                                 |

---

## Regeln für die Ordnerauswahl

Ein Ordner wird nur berücksichtigt, wenn alle folgenden Bedingungen erfüllt sind:

1. Der Ordnername enthält mindestens ein `_`.
2. Der Ordnername beginnt **nicht** mit `_`.
3. Das Änderungsdatum des Ordners ist älter als die angegebene Anzahl von Tagen.

### Beispiele

| Ordnername    | Ergebnis            |
| ------------- | ------------------- |
| `projekt_alt` | Wird berücksichtigt |
| `backup_2024` | Wird berücksichtigt |
| `_archiv`     | Wird ignoriert      |
| `dokumente`   | Wird ignoriert      |

---

## Beispiele

### Testlauf auf lokalem Pfad

```powershell
folder-cleanup.exe -path "C:\daten\ordner" -days 30 -dry-run=true
```

### Testlauf auf einem UNC-Pfad

```powershell
folder-cleanup.exe -path "\\server\share\ordner" -days 30 -dry-run=true
```

### Ordner tatsächlich löschen

```powershell
folder-cleanup.exe -path "C:\daten\ordner" -days 30 -dry-run=false
```

### Mit Logdatei

```powershell
folder-cleanup.exe -path "C:\daten\ordner" -days 30 -dry-run=false -log "C:\temp\cleanup.log"
```

---

## Beispielausgabe

### Testlauf

```text
2026/06/18 10:15:00 Starte Suche in: C:/daten/ordner | Älter als 30 Tage | DryRun=true
2026/06/18 10:15:02 [TEST] Würde löschen: C:/daten/ordner/projekt_alt (letzte Änderung: 2026-04-10T09:30:00+02:00)
```

### Löschlauf

```text
2026/06/18 10:15:00 Starte Suche in: C:/daten/ordner | Älter als 30 Tage | DryRun=false
2026/06/18 10:15:02 Lösche Ordner: C:/daten/ordner/projekt_alt (letzte Änderung: 2026-04-10T09:30:00+02:00)
```

---

## Hinweise

* Das Tool verwendet `os.RemoveAll()` zum Löschen von Verzeichnissen.
* Gelöschte Ordner können nicht wiederhergestellt werden.
* Führe neue Konfigurationen immer zuerst mit `-dry-run=true` aus.
* UNC-Pfade werden unterstützt.
* Alle Pfade werden intern in ein plattformunabhängiges Format mit `/` umgewandelt.

---

## Lizenz

Interne Nutzung oder gemäß den Bedingungen deines Projekts.
