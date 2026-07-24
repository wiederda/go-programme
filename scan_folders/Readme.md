# 🗂️ scan_folders

Ein kompaktes Go-Tool, das rekursiv Verzeichnisse durchsucht und
**Ordner identifiziert, die von OpenSSL angelegt wurden, aber ungenutzt
geblieben sind**.\
Es erkennt **Ordner, die mindestens eine `.txt`-Datei enthalten, aber
keine Zertifikatsdateien (`.cer`, `.pem`, `.p7b`, `.pfx`)**.\
Gefundene Ordner können automatisch gelöscht oder im *Dry-Run* nur
gelistet werden.

------------------------------------------------------------------------

## 🚀 Funktionen

-   Rekursiver Scan aller Unterordner\
-   Findet OpenSSL-Ordner, die `.txt` enthalten, aber **keine
    Zertifikatsdateien**\
-   Optionales automatisches Löschen (`-dry-run=false`)\
-   Altersfilter (`-days=N` -- nur Ordner löschen, die älter sind als N
    Tage)\
-   Überspringt Zeitstempel-Ordner (`_YYYYMMDD_HHMM`)\
-   Ausführliches Logfile\
-   Läuft unter Windows, Linux und macOS

------------------------------------------------------------------------

## 🧩 Beispielstruktur

    C:\Zertifikate
    ├── OrgA_20240101_1200
    │   ├── OrgA.txt
    │   └── OrgA.cer
    ├── OrgB
    │   └── OrgB.txt   ← Treffer
    └── Logs
        └── system.txt

------------------------------------------------------------------------

## ⚙️ Verwendung

    scan_folders -path=<Verzeichnis> -log=<Datei> -dry-run=true|false -days=<Alter>

**Beispiel:**

    scan_folders -path=C:\Zertifikate -dry-run=true -days=30 -log=scan_output.txt

### Hilfe anzeigen

    scan_folders -h
