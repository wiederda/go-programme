# 🛠️ FindReplace

FindReplace dient zur automatisierten Ersetzung spezifischer Zeilen in Dateien. Es unterstützt sowohl die Bearbeitung einzelner Dateien als auch die Batch-Verarbeitung von Dateien anhand einer Liste.

## ✨ Funktionen

Das Tool kann zwei Hauptaufgaben erfüllen:

1. **Einzelne Datei bearbeiten:** Ersetzt eine definierte alte Zeile durch eine neue Zeile in einer einzelnen Datei.
2. **Batch-Verarbeitung:** Verarbeitet eine Liste von Konfigurationseinträgen (in einer separaten Datei) und führt die Ersetzung für alle relevanten Dateien durch, die zum aktuellen Hostnamen gehören.

### ⚙️ Details der Bearbeitung

* **Verarbeitung:** Das Tool liest die gesamte Datei in den Speicher, führt die Ersetzung durch und schreibt dann die gesamte, modifizierte Datei zurück, wodurch die Originaldatei überschrieben wird.
* **Berechtigungen:** Nach erfolgreicher Bearbeitung wird versucht, die Datei ausführbar zu machen (Setzt die Berechtigung `0755`), es sei denn, die Datei endet auf `.txt`.

## 💡 Verwendung

Das Programm wird über die Kommandozeile aufgerufen und unterscheidet zwischen drei Modi: Hilfe, Einzelbearbeitung und Batch-Verarbeitung.

### 1. Hilfe anzeigen

Um die Nutzungsbeispiele zu sehen:

```bash
./FindReplace --help
```

### 2. Einzelne Datei bearbeiten (Direktaufruf)

Verwenden Sie diesen Modus, wenn Sie nur eine Datei bearbeiten möchten.

**Syntax:**
```bash
./FindReplace <dateiname> <alte_zeile> <neue_zeile>
```

**Beispiel:**
Angenommen, Sie möchten in `config.conf` die Zeile `old_api_key` durch `new_api_key` ersetzen.

```bash
./FindReplace config.conf "old_api_key" "new_api_key"
```

### 3. Batch-Verarbeitung (Mit Listenfile)

Verwenden Sie diesen Modus, um viele Dateien gleichzeitig zu bearbeiten, basierend auf einer strukturierten Liste.

**Syntax:**
```bash
./FindReplace <pfad_zur_listenfile.txt>
```

#### 📄 Format der Listenfile (`file_list.txt`)

Jede Zeile in der Liste muss exakt dem folgenden Format folgen:

```
<hostname> <datei_pfad> <alte_zeile> <neue_zeile>
```

**Beispiel für `file_list.txt`:**

```
server-a /etc/app/config.ini "timeout=30" "timeout=60"
server-b /etc/app/settings.xml "user=guest" "user=admin"
server-c /etc/app/data.yml "version: 1.0" "version: 2.0"
```

**Wichtige Filterungen:**

* **Hostname-Filter:** Das Tool verarbeitet nur die Einträge, deren `hostname` mit dem Hostnamen des aktuell laufenden Systems übereinstimmt.
* **HTTPS-Filter:** Dateien, deren Pfad mit `https://` beginnt, werden ignoriert.

**Ausführung:**

```bash
./FindReplace file_list.txt
```