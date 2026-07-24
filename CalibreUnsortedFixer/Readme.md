# CalibreUnsortedFixer

Ein Kommandozeilenwerkzeug zum Erkennen und Zuordnen unsortierter Bücher in einer Calibre-Installation.

Das Tool vergleicht die Einträge aus der `metadata.db` mit den Regalzuordnungen in der `app.db`. Bücher ohne Zuordnung werden automatisch einem definierten Regal – standardmäßig `_unsortiert` – hinzugefügt.

## Funktionen

* Erkennt Bücher ohne Regalzuordnung.
* Zeigt fehlende Buch-IDs an.
* Ordnet unsortierte Bücher automatisch einem Regal zu.
* Frei konfigurierbare Regal-ID.
* Verwendet reines Go mit `modernc.org/sqlite` und benötigt keine CGO-Abhängigkeiten.

---

## Voraussetzungen

* Go 1.20 oder neuer
* Zugriff auf die Calibre-Datenbanken:

  * `metadata.db`
  * `app.db`

> **Wichtig:** Erstelle vor der Verwendung unbedingt Sicherungskopien beider Datenbanken.

---

## Installation

Repository klonen und kompilieren:

```bash
git clone <repository-url>
cd <repository-name>

go build -o CalibreUnsortedFixer .
```

Unter Windows:

```powershell
go build -o CalibreUnsortedFixer.exe .
```

---

## Verwendung

```text
CalibreUnsortedFixer --metadata=PFAD --app=PFAD [Optionen]
```

### Optionen

| Option       | Standardwert | Beschreibung                          |
| ------------ | ------------ | ------------------------------------- |
| `--metadata` | –            | Pfad zur `metadata.db` (erforderlich) |
| `--app`      | –            | Pfad zur `app.db` (erforderlich)      |
| `--shelf`    | `9`          | ID des Regals für `_unsortiert`       |
| `--missing`  | `false`      | Zeigt nur fehlende Buch-IDs an        |
| `--help`     | `false`      | Zeigt die Hilfe an                    |

---

## Funktionsweise

Das Tool führt folgende Schritte aus:

1. Öffnet die `metadata.db`.
2. Bindet die `app.db` per `ATTACH DATABASE` ein.
3. Ermittelt alle Bücher ohne Eintrag in `book_shelf_link`.
4. Ordnet diese Bücher dem angegebenen Regal zu.

Wenn die Option `--missing` gesetzt ist, werden lediglich die fehlenden Buch-IDs ausgegeben.

---

## Beispiele

### Fehlende Buch-IDs anzeigen

```powershell
CalibreUnsortedFixer.exe `
  --metadata="C:\Calibre\metadata.db" `
  --app="C:\Calibre\app.db" `
  --missing
```

### Unsortierte Bücher dem Standardregal zuordnen

```powershell
CalibreUnsortedFixer.exe `
  --metadata="C:\Calibre\metadata.db" `
  --app="C:\Calibre\app.db"
```

### Benutzerdefinierte Regal-ID verwenden

```powershell
CalibreUnsortedFixer.exe `
  --metadata="C:\Calibre\metadata.db" `
  --app="C:\Calibre\app.db" `
  --shelf=15
```

### Verwendung mit UNC-Pfaden

```powershell
CalibreUnsortedFixer.exe `
  --metadata="\\server\ebooks\metadata.db" `
  --app="\\server\ebooks\app.db"
```

---

## Beispielausgabe

### Fehlende Buch-IDs anzeigen

```text
Fehlende Buch-IDs:
12
48
127
```

### Bücher zuordnen

```text
Buch 'Der Hobbit' wurde dem Regal '_unsortiert' zugeordnet.
Buch '1984' wurde dem Regal '_unsortiert' zugeordnet.
```

### Keine unsortierten Bücher gefunden

```text
Keine unsortierten Bücher gefunden.
```

---

## Datenbankschema

Das Tool verwendet folgende Tabellen:

### `metadata.db`

* `books`

### `app.db`

* `shelf`
* `book_shelf_link`

Erwartete Beziehung:

```text
books.id -> book_shelf_link.book_id
shelf.id -> book_shelf_link.shelf
```

---

## Sicherheitshinweise

* Vor jedem Lauf Sicherungskopien von `metadata.db` und `app.db` erstellen.
* Die Datenbanken dürfen während der Ausführung nicht gleichzeitig von anderen Anwendungen verändert werden.
* Falsche Regal-IDs führen dazu, dass Bücher unerwarteten Regalen zugeordnet werden.
* Das Tool prüft, ob die angegebene Regal-ID existiert, bevor Änderungen vorgenommen werden.

---

## Lizenz

Interne Nutzung oder gemäß den Bedingungen deines Projekts.
