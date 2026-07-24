# `checkfolders`

**`checkfolders`** ist ein Kommandozeilen-Tool, das rekursiv Verzeichnisse nach dem Vorhandensein von `.jpg`- und/oder `.nfo`-Dateien durchsucht. Es hilft dabei, schnell zu identifizieren, welche Ordner möglicherweise wichtige Medien oder Metadaten fehlen.

## 🚀 Funktionen

Das Tool durchsucht ein angegebenes Wurzelverzeichnis auf fehlende Dateien in allen Unterordnern.

*   **Suche nach `.jpg`:** Prüft, ob Verzeichnisse `.jpg`-Dateien enthalten.
*   **Suche nach `.nfo`:** Prüft, ob Verzeichnisse `.nfo`-Dateien enthalten.
*   **Ausschlüsse:** Ermöglicht das Ignorieren von bestimmten Ordnernamen während der Suche (z.B. `#recycle`, `.git`).
*   **Ausgabe:** Speichert die Pfade zu den fehlenden Ordnern in einer angegebenen Datei.

## ⚙️ Verwendung

Das Programm wird über Kommandozeilen-Flags gesteuert.

### Voraussetzungen

*   Go (Version 1.16 oder neuer)

### Syntax

```bash
checkfolders --root=<Startverzeichnis> [--exclude=<Liste>] [--out=<Ausgabepfad>] [--jpg] [--nfo]
```

## 📋 Parameter (Flags)

| Flag | Beschreibung | Standardwert |
| :--- | :--- | :--- |
| **`--root`** | **Pfad zum Startverzeichnis** der rekursiven Suche. (Obligatorisch) | *Fehlt bei Ausführung* |
| **`--exclude`** | Kommagetrennte Liste von Ordnernamen, die von der Suche ausgeschlossen werden sollen. | `#recycle,.ds_store,thumbs.db` |
| **`--out`** | Pfad zur Ausgabedatei, in die die Ergebnisse geschrieben werden. | `missing_files.txt` |
| **`--jpg`** | Nur Verzeichnisse mit fehlenden `.jpg`-Dateien melden. | `false` |
| **`--nfo`** | Nur Verzeichnisse mit fehlenden `.nfo`-Dateien melden. | `false` |
| **`--help`** | Zeigt die Hilfeseite an und beendet das Programm. | `false` |

## 💡 Beispiele

### Beispiel 1: Grundlegende Suche (alle fehlenden Dateien)

Suche im Verzeichnis `C:\Hörspiele` und speichere das Ergebnis in `D:\result.txt`.

```bash
checkfolders --root="C:\Hörspiele" --out="D:\result.txt"
```

### Beispiel 2: Nur fehlende `.nfo`-Dateien finden

Suche nur nach Ordnern, die keine `.nfo`-Datei enthalten.

```bash
checkfolders --root="\\Server\Share" --nfo
```

### Beispiel 3: JPG-Ausschlüsse und spezifische Ausgabe

Suche im Verzeichnis `C:\Media`, ignoriere Ordner, die `#recycle` oder `.git` enthalten, und melde nur fehlende `.jpg`-Dateien.

```bash
checkfolders --root="C:\Media" --exclude=".git,#recycle" --out="D:\missing_jpegs.txt" --jpg
```

---

## 📄 Code-Struktur (Optional, aber nützlich)

Das Programm ist in Go geschrieben und verwendet die Standardbibliothek `flag` sowie `path/filepath` zur Handhabung der Dateisystemoperationen.

*   **Sprache:** Go
*   **Komplexität:** Mittel (Rekursive Pfadverarbeitung)

***