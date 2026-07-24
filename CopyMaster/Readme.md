# CopyMaster

CopyMaster ist ein einfaches CLI-Tool zum Kopieren von Dateien mit einer bestimmten Endung aus einem Quellverzeichnis in ein Zielverzeichnis. Es erlaubt zudem das Ausschließen bestimmter Verzeichnisse vom Kopiervorgang.

## Verwendung

Das Tool kann über die Kommandozeile mit verschiedenen Optionen aufgerufen werden:

```sh
./CopyMaster --source=<source_dir> --dest=<destination_dir> --ext=<file_extension> [--exclude=<exclude1,exclude2,...>]
```

### Parameter

- `--source`   : Quellverzeichnis (Pfad mit Dateien, die kopiert werden sollen)
- `--dest`     : Zielverzeichnis (Pfad, in den die Dateien kopiert werden)
- `--ext`      : Dateiendung der zu kopierenden Dateien (z.B. `.txt`)
- `--exclude`  : Komma-separierte Liste von Verzeichnissen, die ausgeschlossen werden sollen
- `--help`     : Zeigt eine Hilfe zur Nutzung an

### Beispiel

```sh
./CopyMaster --source=/home/user/Dokumente --dest=/home/user/Backup --ext=.txt --exclude=geheim,privat
```

Dieses Beispiel kopiert alle `.txt`-Dateien aus `/home/user/Dokumente` nach `/home/user/Backup`, ignoriert jedoch die Verzeichnisse `geheim` und `privat`.

