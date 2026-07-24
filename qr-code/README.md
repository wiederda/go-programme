# QR-Code CLI Tool

Ein einfaches Go-Kommandozeilentool zum Schreiben und Lesen von
QR-Codes.

## Funktionen

-   **QR-Code erstellen**
    -   Übergabe von Text via `-text`
    -   Ausgabe als PNG-Datei (`-out`)
    -   Konfigurierbare Größe (`-size`)
-   **QR-Code lesen**
    -   PNG-Datei via `-in` einlesen
    -   Inhalt automatisch erkennen

## Installation

Zuerst benötigte Bibliotheken installieren:

``` bash
go get github.com/skip2/go-qrcode
go get github.com/liyue201/goqr
```

## Beispiele

### QR-Code schreiben

``` bash
qr-code.exe -mode write -text "Hallo Welt" -out hallo.png -size 300
```

### QR-Code lesen

``` bash
qr-code.exe -mode read -in hallo.png
```

## Flags

  Flag    Beschreibung
  ------- -------------------------------
  -mode   write oder read
  -text   Text zum Schreiben (write)
  -in     Eingabedatei für Lesen (read)
  -out    Ausgabedatei für Schreiben
  -size   Größe des QR-Codes in Pixeln
