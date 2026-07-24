# MediaFolderRenamer

**MediaFolderRenamer** ist ein Tool, das die Umbenennung von Medienordnern und Dateien basierend auf benutzerdefinierten Regeln ermöglicht. Es unterstützt Funktionen wie das Umbenennen von Ordnern basierend auf Dateinamen sowie das Umbenennen von Dateien basierend auf Ordnernamen.

## Verwendung
```plaintext
MediaFolderRenamer [Optionen]
```

## Optionen

| Option            | Beschreibung                                                                                                   | Standardwert          |
|--------------------|---------------------------------------------------------------------------------------------------------------|-----------------------|
| **`--rootDir`**    | Stammverzeichnis, von dem aus gestartet werden soll. Unterstützt UNC-Pfade wie `\\servername\freigabe`. Hinweis: In der Kommandozeile müssen Backslashes (\\) doppelt maskiert werden (\\\\). | `./`                     |
| **`--extensions`** | Durch Kommas getrennte Liste von Dateierweiterungen, die überprüft werden sollen.                             | `mp4,mkv`            |
| **`--excludeDir`** | Durch Kommas getrennte Liste von Ordnernamen, die ausgeschlossen werden sollen.                               | `#recycle`           |
| **`--renameByFile`** | Umbenennen der Ordner basierend auf den Dateinamen im Ordner.                                               | `false` (deaktiviert)|
| **`--renameByFolder`** | Umbenennen von Dateien im Ordner basierend auf dem Ordnernamen.                                           | `false` (deaktiviert)|
| **`--help`**       | Zeigt diese Hilfe an.                                                                                        | -                    |

## Beispiele
**Beispiel1:**
```plaintext
./MediaFolderRenamer --rootDir=\\servername\freigabe --extensions=mkv,mp4 --excludeDir=#recycle --renameByFolder
```
### Beschreibung des Beispiel:
- Das Stammverzeichnis ist `\\servername\freigabe`.
- Nur Dateien mit den Erweiterungen `mkv` und `mp4` werden überprüft.
- Ordner mit dem Namen `#recycle` werden ausgeschlossen.
- Dateien im Ordner werden basierend auf dem Ordnernamen umbenannt.

**Beispiel2:**
```plaintext
./MediaFolderRenamer --rootDir=\\servername\freigabe --extensions=mkv,mp4 --excludeDir=#recycle --renameByFile
```
### Beschreibung des Beispiel:
- Das Stammverzeichnis ist `\\servername\freigabe`.
- Nur Dateien mit den Erweiterungen `mkv` und `mp4` werden überprüft.
- Ordner mit dem Namen `#recycle` werden ausgeschlossen.
- Ordner werden basierend auf den Dateinamen im Ordner umbenannt

## Voraussetzungen
- **Dateiberechtigungen:** Stellen Sie sicher, dass Sie Schreibrechte für die Zielverzeichnisse und -dateien haben.

## Fehlerbehebung
- **Nicht unterstützte Erweiterungen:** Stellen Sie sicher, dass die angegebenen Dateierweiterungen im Verzeichnis vorhanden sind.
