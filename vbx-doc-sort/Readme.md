# vbx-doc-sort

`vbx-doc-sort` sortiert die Funktionsabschnitte der VBX-Markdown-Dokumentation alphabetisch.

Das Programm durchsucht Markdown-Dateien nach Funktionsüberschriften und ordnet die kompletten Funktionsblöcke anhand des Funktionsnamens. Der Inhalt innerhalb eines Funktionsblocks bleibt dabei unverändert.

## Was wird sortiert?

Unterstützt werden sowohl Funktionen mit Namespace:

```text
### app.StartupPath()
### array.Sort(values)
### cert.GenerateKey(outFile, algo, bits)
```

als auch globale Funktionen:

```text
## Format(expression, style)
## Left(s, n)
## Right(s, n)
```

Die Sortierung erfolgt alphabetisch und unabhängig von Groß-/Kleinschreibung.

## Markdown-Codeblöcke

Für die Erkennung wird der Markdown-AST von Goldmark verwendet.

Dadurch werden Überschriften innerhalb von fenced Codeblöcken nicht als echte Funktionsüberschriften erkannt.

Beispielsweise bleibt folgendes vollständig unberücksichtigt:

````markdown
```vbx
## Das ist keine echte Funktionsüberschrift
Print Format(file.AccessTime(path), "YYYY-MM-DD HH:mm:ss")
```
````

## Ausgabe

Die sortierten Dateien werden in ein separates Ausgabeverzeichnis geschrieben.

Die Originaldateien werden nicht verändert.

Der Bereich vor der ersten erkannten Funktion bleibt unverändert. Die einzelnen Funktionsblöcke werden anschließend in alphabetischer Reihenfolge ausgegeben.

## Verwendung

```text
vbx-doc-sort <input>... <output> [optionen]
```

Das letzte Positionsargument ist immer das Ausgabeverzeichnis.

### Verzeichnis verarbeiten

```text
vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert"
```

Alle `.md`-Dateien des Verzeichnisses werden verarbeitet.

### Einzelne Datei verarbeiten

```text
vbx-doc-sort.exe "C:\test\md\cert.md" "C:\test\sortiert"
```

### Mehrere Dateien verarbeiten

```text
vbx-doc-sort.exe "C:\test\md\cert.md" "C:\test\md\7z.md" "C:\test\sortiert"
```

Damit können einzelne problematische oder noch nicht bearbeitete Dateien gezielt verarbeitet werden.

## Optionen

### Hilfe

```text
-h
--help
```

Zeigt die verfügbaren Optionen und Beispiele an.

### Version

```text
-version
--version
```

Zeigt die Programmversion an.

### Dry-Run

```text
-dry-run
--dry-run
```

Analysiert und sortiert die Dateien nur virtuell. Es werden keine Dateien geschrieben.

Beispiel:

```text
vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert" -dry-run
```

### Dateien ausschließen

```text
-exclude <datei>
--exclude <datei>
```

Schließt eine Datei anhand ihres Dateinamens aus.

Die Option kann mehrfach angegeben werden:

```text
vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert" ^
    -exclude Allgemein.md ^
    -exclude README.md
```

Dabei wird nur der Dateiname berücksichtigt, nicht der vollständige Pfad.

## Ausgabe

Während der Verarbeitung wird für jede Datei das Ergebnis angezeigt:

```text
OK    cert.md                  12 Funktionen gefunden, 12 verarbeitet - Reihenfolge geändert
OK    geo.md                    1 Funktionen gefunden, 1 verarbeitet - bereits sortiert
```

Am Ende wird eine Zusammenfassung ausgegeben:

```text
Zusammenfassung:
  Dateien:                 44
  Funktionen gefunden:    671
  Funktionen verarbeitet: 671
  Funktionen übersprungen: 0
  Funktionen fehlerhaft:   0

  Dateien sortiert:         41
  Dateien unverändert:      3
  Dateien übersprungen:     0
  Dateien mit Fehlern:      0
```

Bei Fehlern werden zusätzlich die betroffenen Funktionsnamen angezeigt.

## Technische Grundlage

Das Programm ist in Go geschrieben und verwendet [Goldmark](https://github.com/yuin/goldmark) für die Markdown-Analyse.

Die Sortierung erfolgt auf Basis der tatsächlichen Markdown-AST-Struktur und nicht durch eine einfache Suche nach `##` oder `###`. Dadurch werden Überschriften innerhalb von Codebeispielen zuverlässig ignoriert.
