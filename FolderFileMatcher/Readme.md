# FolderFileMatcher

Ein Tool zum Überprüfen von Dateien in einem Stammverzeichnis und seinen Unterordnern. Es speichert Pfade von Dateien, bei denen der Name des Ordners nicht mit dem Namen der darin enthaltenen Dateien (ohne Erweiterung) übereinstimmt. Dateien direkt im Root-Verzeichnis werden ignoriert.

## Verwendung

```bash
FolderFileMatcher [Optionen]
```

### Optionen

| Option         | Beschreibung                                                                                                                                                              | Standardwert             |
|-----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------|
| `--rootDir`     | Stammverzeichnis, von dem aus gestartet werden soll. Unterstützt UNC-Pfade wie `\\servername\freigabe`. Hinweis: In der Kommandozeile müssen Backslashes (\\) doppelt maskiert werden (\\\\). | `./`                     |
| `--extensions`  | Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen.                                                                                        | `mp4,mkv`               |
| `--excludeDir`  | Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen.                                                                                          | `#recycle`              |
| `--outputFile`  | Optionaler Name der Datei, in der die Ergebnisse gespeichert werden.                                                                                                     | `output.txt`            |
| `--help`        | Zeigt die Hilfe mit einer Beschreibung aller Optionen und der Verwendung an.                                                                                            | -                       |

### Beispiel

1. Überprüfen aller Dateien in einem Verzeichnis:
   ```bash
   ./FolderFileMatcher --rootDir=./test --extensions=mp4,mkv --excludeDir=#recycle
   ```

2. Benutzerdefinierte Ausgabedatei:
   ```bash
   ./FolderFileMatcher --rootDir=\\\\servername\\freigabe --extensions=mkv,mp4 --excludeDir=#recycle --outputFile=results.txt
   ```

### Verhalten

- **Dateien im Root-Verzeichnis**:
  Dateien, die direkt im angegebenen `rootDir` liegen, werden ignoriert und nicht in die Prüfung einbezogen.
  
- **Prüfungskriterien**:
  Das Tool speichert Pfade von Dateien, bei denen der Name des Ordners nicht mit dem Namen der Dateien (ohne Erweiterung) in diesem Ordner übereinstimmt. Die betroffenen Pfade werden in der Ausgabedatei (`outputFile`) gespeichert.

- **Ausgeschlossene Verzeichnisse**:
  Ordner, die in der Liste `excludeDir` angegeben sind, werden vollständig übersprungen, einschließlich aller darin enthaltenen Dateien und Unterordner.
