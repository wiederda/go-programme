## Beschreibung
Dieses Programm lädt Skripte und Konfigurationen auch aus privaten Git-Repositories auf Linux-Server herunter und überschreibt vorhandene Dateien. 'Bereinigen' umfasst das Entfernen überflüssiger Leerzeichen und potenziell problematischer Inhalte. Alle Dateien, bis auf `*.txt`, werden ausführbar gemacht.

### Verwendung
```
<Programm> <LISTE_DATEI_URL> <LISTE_DATEI> [TOKEN_DATEI]
```

### Optionen
- `--help`        Zeigt die Hilfe an und beendet das Programm.
- `TOKEN_DATEI`   Ein optionaler Pfad zu einer Datei, die einen Zugriffstoken (Access-Token) enthält, um private Repositories zu authentifizieren.
- `LISTE_DATEI`   Kann mehrere Zeilen enthalten. Leerzeilen und Zeilen, die mit `#` beginnen, werden ignoriert.

### Beispiel
```bash
./get-git https://example.com/list.txt liste.txt
./get-git https://example.com/list.txt liste.txt token.txt
```

### Beispiel für <LISTE_DATEI>
```
<ALL|Hostname> https://example.com/skript.sh skript.sh
<ALL|Hostname> https://example.com/skript.sh skript.sh token.txt
```

### Hinweise
- Stellen Sie sicher, dass die URL korrekt ist, indem Sie sie in einem Browser öffnen oder mit `curl` bzw. `wget` testen, und überprüfen Sie, ob die Datei zugänglich ist.