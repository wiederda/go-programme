# 🔧 Windows Backup Utility

Ein einfaches, aber flexibles Kommandozeilen-Tool zum Kopieren von Dateien und Ordnern nach einer Suchliste.  
Behält die Originalordnerstruktur bei und kann optional ein ZIP-Archiv erstellen.

---

## 🧩 Funktionen

- Kopiert **Dateien und Ordner** gemäß einer Liste (`-searchlist`)
- Behält die **Originalverzeichnisstruktur** (ohne Laufwerksbuchstaben oder UNC-Prefix)
- Unterstützt **UNC-Pfade** (\\Server\Share)
- Optionales **ZIP-Archiv** mit `-zip=<pfad>`
- **Sicherheitsprüfung:** ZIP-Datei selbst wird nicht mitgezippt
- **Fallback-Logik:** Wenn eine Datei nicht exakt gefunden wird, wird nach ähnlichen Namen gesucht
- Fortschrittsanzeige und robuste Fehlerbehandlung

---

## ⚙️ Verwendung

```bash
go run main.go -searchlist=<suchliste.txt> -dest=<zielordner> [-zip=<zipdatei>]
```

### Parameter

| Parameter | Beschreibung |
|------------|---------------|
| `-searchlist` | Pfad zur Textdatei mit absoluten Windows- oder UNC-Pfaden (eine Datei oder ein Ordner pro Zeile) |
| `-dest` | Zielordner, in den kopiert wird |
| `-zip` | *(optional)* Pfad zur ZIP-Datei, die erstellt werden soll |
| `-h` | Zeigt die Hilfe an |

---

## 📄 Beispiel

```bash
go run main.go -searchlist=suchliste.txt -dest=E:\Backup -zip=E:\Backup\Backup.zip
```

### Beispiel `suchliste.txt`

```
C:\Users\TEST\Documents\ProjektA\readme.txt
C:\Users\TEST\Pictures\Urlaub\
\\Server\Share\Teamdaten
```

Ergebnis:

```
E:\Backup\Users\TEST\Documents\ProjektA\readme.txt
E:\Backup\Users\TEST\Pictures\Urlaub\
E:\Backup\Server\Share\Teamdaten\
```

---

## 🧠 Unscharfe Dateisuche

Wenn eine Datei aus der Liste **nicht exakt** gefunden wird:

- Das Programm sucht im gleichen Verzeichnis nach Dateien mit **demselben Namen, aber beliebiger Erweiterung**  
  z. B.:
  ```
  Gesucht:  Bilder\...\20251111
  Gefunden: Bilder\...\20251111.jpg
  ```
- Wird **genau eine passende Datei** gefunden, wird diese kopiert.
- Gibt es **mehrere Treffer**, wird keine Datei kopiert (um Fehlkopien zu vermeiden).

---

## 🗜️ ZIP-Archiv

Wenn `-zip` angegeben ist:

- Nach erfolgreichem Kopieren wird der gesamte Zielordner zu einer ZIP-Datei gepackt.
- Die ZIP-Datei selbst wird **nicht** in das Archiv aufgenommen.
- Beispiel:
  ```bash
  -dest C:\myBackup\Backup
  -zip  C:\myBackup\Backup.zip
  ```
  ➜ Ergebnis: `Backup.zip` enthält alle Dateien und Ordner aus `Backup\`.

---

## 📋 Hinweise

- Unterstützt **normale Windows-Pfade** (`C:\Ordner\Datei.txt`)
- **UNC-Pfade** (`\\Server\Freigabe\...`) funktionieren ebenfalls
- **Leere Zeilen** in der Suchliste werden ignoriert
- Existierende Dateien werden **überschrieben**

---

## 🏗️ Beispielaufruf unter Windows

```powershell
PS> CopyFiles -searchlist="C:\Listen\backup.txt" -dest="E:\Backup" -zip="E:\Backup\Backup.zip"
```
