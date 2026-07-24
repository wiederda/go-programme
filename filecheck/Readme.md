# 📂 filecheck: Duplikat-Finder & Backup (Go)

## 📝 Beschreibung

Das `filecheck`-Kommandozeilenprogramm (CLI) dient dazu, rekursiv in einem angegebenen Ordner nach **doppelten Dateien** zu suchen. Der Vergleich erfolgt anhand der **Dateigröße** (als schneller Vorfilter) und des **SHA-256-Hashs** (als definitive Bestätigung der Identität).

Das Tool bietet eine optionale, sichere Backup-Funktion, bei der alle gefundenen Duplikate in einen separaten Backup-Ordner verschoben werden, um dem Benutzer die endgültige Entscheidung über die Löschung zu überlassen.

---

## ✨ Funktionen

* **Plattformübergreifend:** Funktioniert zuverlässig unter Windows (unterstützt `C:\`-Pfade und UNC-Pfade wie `\\Server\Share`) sowie Linux/macOS.
* **Hash-basierter Vergleich:** Nutzt **SHA-256**, um die absolute Gleichheit der Inhalte zu gewährleisten.
* **Performance-optimiert:** Hash-Berechnung nur für Dateien gleicher Größe.
* **Sicherer Backup-Modus:** Verschiebt alle Duplikate in einen definierten Backup-Ordner, strukturiert nach Hash-Gruppen.
* **Protokollierung:** Schreibt den detaillierten Bericht über gefundene Duplikate und Aktionen in eine Log-Datei oder auf die Konsole.