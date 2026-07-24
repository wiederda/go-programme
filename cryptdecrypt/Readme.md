# Verwendung von cryptdecrypt

## Funktionsweise

Das Tool `cryptdecrypt` wird verwendet, um Texte zu verschlüsseln oder zu entschlüsseln. 

### Argumente

- `--mode`  
  Modus: `'crypt'` für Verschlüsselung oder `'decrypt'` für Entschlüsselung

- `--text`  
  Text zum Verschlüsseln/Entschlüsseln.  
  Für Entschlüsselung im Format `salt:ciphertext`

- `--password`  
  Passwort für Verschlüsselung/Entschlüsselung

- `--help`
  Ruft die Hilfe auf  

### Beispiele

#### Verschlüsselung
```
./cryptdecrypt --mode crypt --text "mein Text" --password "meinPasswort"
```

#### Entschlüsselung
```
./cryptdecrypt --mode decrypt --text "salt:ciphertext" --password "meinPasswort"
```