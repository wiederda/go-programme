# OCR CLI Tool

Ein einfaches Kommandozeilenprogramm zur Texterkennung (OCR) aus Bildern mithilfe von Tesseract.

Das Tool übergibt eine Bilddatei an Tesseract, liest die erzeugte Textdatei ein, gibt den erkannten Text auf der Konsole aus und entfernt die temporäre Ausgabedatei anschließend automatisch.

---

## Funktionen

* Extrahiert Text aus Bilddateien per OCR.
* Nutzt eine vorhandene Tesseract-Installation.
* Gibt den erkannten Text direkt in der Konsole aus.
* Löscht temporäre Ausgabedateien automatisch.
* Funktioniert plattformübergreifend.

---

## Voraussetzungen

* Go 1.20 oder neuer
* Installiertes Tesseract OCR
* Tesseract muss im Systempfad (`PATH`) verfügbar sein

---

## Installation

Repository klonen und kompilieren:

```bash
git clone <repository-url>
cd <repository-name>

go build -o ocr-cli .
```

Unter Windows:

```powershell
go build -o ocr-cli.exe .
```

---

## Tesseract installieren

### Windows

1. Tesseract herunterladen und installieren.
2. Das Installationsverzeichnis zur Umgebungsvariable `PATH` hinzufügen.

Beispiel:

```text
C:\Program Files\Tesseract-OCR
```

Prüfen, ob Tesseract verfügbar ist:

```powershell
tesseract --version
```

### Linux

#### Debian / Ubuntu

```bash
sudo apt install tesseract-ocr
```

#### Fedora / Rocky Linux

```bash
sudo dnf install tesseract
```

### macOS

Mit Homebrew:

```bash
brew install tesseract
```

---

## Verwendung

```text
ocr-cli <Bildpfad>
```

### Hilfe anzeigen

```text
ocr-cli --help
```

---

## Beispiele

### Bild analysieren

```bash
ocr-cli ./rechnung.png
```

### Windows-Pfad verwenden

```powershell
ocr-cli.exe "C:\Bilder\dokument.jpg"
```

### Netzwerkpfad verwenden

```powershell
ocr-cli.exe "\\server\share\scan.png"
```

---

## Beispielausgabe

```text
Erkannter Text:

Rechnung Nr. 2026-001
Datum: 18.06.2026

Gesamtbetrag: 149,90 EUR
```

---

## Unterstützte Bildformate

Die unterstützten Formate hängen von der installierten Tesseract-Version ab. Typischerweise werden folgende Formate unterstützt:

* PNG
* JPEG / JPG
* TIFF
* BMP
* GIF
* WebP

---

## Funktionsweise

1. Prüft, ob die angegebene Bilddatei existiert.
2. Ermittelt den Dateinamen ohne Erweiterung.
3. Führt Tesseract mit der Bilddatei aus.
4. Liest die erzeugte Textdatei `<dateiname>_ocr.txt`.
5. Gibt den erkannten Text auf der Konsole aus.
6. Löscht die temporäre Textdatei.

---

## Bekannte Einschränkungen

* Es kann keine OCR-Sprache ausgewählt werden.
* Das Tool verwendet die Standardsprache der Tesseract-Installation.
* Die Ausgabedatei wird im aktuellen Arbeitsverzeichnis erzeugt.
* Bereits vorhandene Dateien mit dem Namen `<dateiname>_ocr.txt` werden überschrieben.
* Die Erkennungsqualität hängt stark von Bildauflösung, Kontrast und Bildqualität ab.

---

## Mögliche Erweiterungen

* Auswahl der OCR-Sprache per Kommandozeilenparameter
* Unterstützung mehrerer Bilddateien gleichzeitig
* Ausgabe in JSON oder CSV
* Speichern der Ergebnisse in einer Datei
* Konfigurierbarer Ausgabeordner
