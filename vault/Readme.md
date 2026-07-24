# 🔐 Mini-Passwort-Keystore (Vault)

Ein sicheres, lokales Kommandozeilen-Tool zur Speicherung, Abfrage und Verwaltung sensibler Passwörter und Schlüssel. Alle Daten werden durch starke kryptografische Verfahren verschlüsselt und erfordern ein Master-Passwort für den Zugriff.

## ✨ Features

*   **Sichere Speicherung:** Nutzt AES-GCM für die Verschlüsselung der Daten.
*   **Master-Passwort-Schutz:** Der Zugriff erfordert ein Master-Passwort, das mittels PBKDF2 in einen sicheren Schlüssel umgewandelt wird.
*   **CRUD-Funktionalität:** Unterstützt das Hinzufügen (`add`), Abrufen (`get`), Löschen (`del`) und Auflisten (`list`) von Einträgen.
*   **Persistenz:** Speichert den gesamten Vault-Zustand in einer JSON-Datei (`vault.json`).

## 🚀 Installation und Setup

Da es sich um ein Go-Projekt handelt, können Sie das Tool lokal kompilieren.

**Voraussetzungen:**
*   [Go](https://golang.org/dl) (Empfohlen: Go 1.18+)

**Schritte:**

1.  Klonen Sie das Repository:
    ```bash
    git clone <Ihr-Repo-Link>
    cd mini-vault
    ```
2.  Kompilieren Sie das Tool:
    ```bash
    go build -o vault
    ```

Das ausführbare Programm `vault` ist nun im aktuellen Verzeichnis verfügbar.

## 📖 Verwendung (Usage)

Das Tool erfordert immer ein **Master-Passwort** (`-master`) und einen **Befehl** (`-cmd`).

### 1. Eintrag hinzufügen (`add`)

Speichert einen neuen Eintrag oder aktualisiert einen bestehenden Eintrag.

```bash
# Syntax: vault -master <Passwort> -cmd add -name <Schlüssel> -value <Wert>
./vault -master meinSuperPasswort -cmd add -name db_pass -value "geheimes-db-passwort-123"
```

### 2. Eintrag abrufen (`get`)

Ruft den Wert eines Eintrags ab und entschlüsselt ihn.

```bash
# Syntax: vault -master <Passwort> -cmd get -name <Schlüssel>
./vault -master meinSuperPasswort -cmd get -name db_pass
# Ausgabe: Wert: geheimes-db-passwort-123
```

### 3. Alle Einträge auflisten (`list`)

Listet alle Namen der gespeicherten Einträge auf, ohne deren Werte zu enthüllen.

```bash
# Syntax: vault -master <Passwort> -cmd list
./vault -master meinSuperPasswort -cmd list
# Ausgabe:
# --- Gespeicherte Einträge ---
# - db_pass
# - api_key
# ----------------------------
```

### 4. Eintrag löschen (`del`)

Löscht einen Eintrag dauerhaft aus dem Vault.

```bash
# Syntax: vault -master <Passwort> -cmd del -name <Schlüssel>
./vault -master meinSuperPasswort -cmd del -name db_pass
```

---

## 🛡️ Sicherheit und Architektur (Deep Dive)

Das Tool ist so konzipiert, dass es höchste Sicherheitsstandards erfüllt:

1.  **Key Derivation (PBKDF2):** Das Master-Passwort wird nicht direkt als Schlüssel verwendet. Stattdessen wird die **PBKDF2**-Funktion mit 100.000 Iterationen verwendet. Dies macht das Knacken des Passworts durch Brute-Force extrem schwierig.
2.  **Salt:** Ein zufälliger `Salt` wird beim ersten Start generiert und in der Konfigurationsdatei gespeichert. Dieser Salt stellt sicher, dass selbst wenn zwei Benutzer dasselbe Master-Passwort verwenden, die resultierenden Schlüssel unterschiedlich sind.
3.  **Verschlüsselung (AES-GCM):** Wir verwenden **AES-GCM (Galois/Counter Mode)**. Dies ist ein *authentifizierter* Verschlüsselungsmodus, der nicht nur die Vertraulichkeit (Verschlüsselung) gewährleistet, sondern auch die **Integrität** der Daten. Jede Manipulation des Ciphertexts führt zu einem Entschlüsselungsfehler.

## 📄 Struktur der Daten

Die Daten werden in der Datei `vault.json` gespeichert und enthalten folgende Struktur:

```json
{
  "version": 1,
  "salt": "...", // Der zufällige Salt
  "entries": [
    {
      "name": "db_pass",
      "iv": "...", // Initialisierungsvektor (zufällig)
      "ciphertext": "..." // Der verschlüsselte Wert
    }
  ]
}
```

## 📄 Lizenz

Dieses Projekt ist unter der [MIT-Lizenz](LICENSE) veröffentlicht.