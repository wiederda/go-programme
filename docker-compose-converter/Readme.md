# Docker Compose Converter

## Verwendung

```
./docker-compose-converter -container <ConterinerID> -output output_file
```
```
./docker-compose-converter -input input_file -output output_file
```

## Flags

- **`-input string`**\
  Pfad zur YAML-Datei mit Containerinformationen (wenn kein direkter Zugriff auf den Docker-Host möglich ist)

- **`-container string`**\
  Name des Docker-Containers oder ID (wenn keine YAML-Datei mit Containerinformationen angegeben ist und direkter Zugriff auf den Docker-Host möglich ist)  

- **`-output string`**  \
  Pfad zur Ausgabe-Docker-Compose-Datei (Standard: `docker-compose.yml`)



