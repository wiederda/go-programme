# check_service

Kleines Windows-CLI-Tool zur Überwachung von Diensten mit Logging und optionaler Ausführung von Tasks (z. B. Mailversand oder Recovery-Aktionen).

---

## Überblick

`check_service` prüft den Status eines Windows-Dienstes und schreibt das Ergebnis in eine Logdatei.
Optional können abhängig vom Zustand weitere Aktionen ausgelöst werden:

* Mailversand über geplanten Task
* Recovery-Task (z. B. Neustart eines Dienstes)

---

## Features

* Abfrage von:

  * Dienststatus (`Running`, `Stopped`, `StopPending`, …)
  * Starttyp (`Auto`, `Manuell`, `Deaktiviert`)
* Logging in UTF-8 (mit BOM) und Windows-CRLF
* Steuerbarer Mailversand
* Ausführung von Windows Tasks (`schtasks`)
* Vollständiges Fehlerlogging (keine stillen Fehler)

---

## Voraussetzungen

* Windows
* Zugriff auf:

  * Service Control Manager (SCM)
  * `schtasks.exe`
* Ausreichende Berechtigungen zum Lesen des Dienststatus und Ausführen von Tasks

---

## Nutzung

```bash
check_service.exe Service=<Name> Log=<Pfad> [Task=<Task>] [Mail=<Task>] [MailStatus=true|false|error]
```

---

## Parameter

| Parameter  | Pflicht | Beschreibung                      |
| ---------- | ------- | --------------------------------- |
| Service    | Ja      | Name des Windows-Dienstes         |
| Log        | Ja      | Pfad zur Logdatei                 |
| Task       | Nein    | Task bei `StopPending` (Recovery) |
| Mail       | Nein    | Task für Mailversand              |
| MailStatus | Nein    | Steuerung für Mailversand         |

---

## MailStatus

| Wert  | Verhalten                  |
| ----- | -------------------------- |
| true  | Mail wird immer gesendet   |
| false | Mail wird nie gesendet     |
| error | Mail nur bei `StopPending` |

---

## Ablauf

1. Parameter werden eingelesen
2. Verbindung zum Service Manager wird aufgebaut
3. Dienst wird abgefragt
4. Status wird geloggt
5. Optional:

   * Mail-Task wird ausgeführt
   * Recovery-Task wird ausgeführt (nur bei `StopPending`)

---

## Logging

Format:

```
YYYY-MM-DD HH:MM:SS <Service> <Starttyp> <Status>
```

Zusätzliche Einträge:

* `Mail gesendet`
* Fehlermeldungen von `schtasks`
* Fehler beim Zugriff auf den Service

---

### Beispiel

```
2026-04-24 10:15:00 Wildfly Auto Running
2026-04-24 10:16:00 Wildfly Auto StopPending
2026-04-24 10:16:01 Mail gesendet
2026-04-24 10:16:02 task 'RestartWildfly' failed: Zugriff verweigert
```

---

## Beispiele

### 1. Nur Status loggen

```bash
check_service.exe Service=Spooler Log=C:\Logs\spooler.log
```

---

### 2. Mail immer senden

```bash
check_service.exe Service=Wildfly Log=C:\Logs\wildfly.log Mail=SendMail MailStatus=true
```

---

### 3. Mail nur bei Fehler + Restart

```bash
check_service.exe Service=Wildfly Log=C:\Logs\wildfly.log Task=RestartWildfly Mail=SendMail MailStatus=error
```

---

### 4. Komplett ohne Mail

```bash
check_service.exe Service=SQLServer Log=C:\Logs\sql.log MailStatus=false
```

---

## Verhalten bei Fehlern

### Service nicht erreichbar

* Log:

  ```
  <Service> Unknown Unknown (<Fehler>)
  ```
* Mail wird ggf. trotzdem gesendet (abhängig von `MailStatus`)

---

### Service existiert nicht

* Gleiches Verhalten wie oben
* Fehler wird im Log sichtbar

---

### Fehler bei `schtasks`

* Werden vollständig geloggt
* Beispiel:

  ```
  task 'SendMail' failed: exit status 1 (Zugriff verweigert)
  ```

---

### Ungültiger `MailStatus`

* Fällt intern auf Standard zurück (`true`)
* Wird im Log vermerkt

---

## Edge Cases

### Leerer Mail-Task

```bash
Mail=
```

→ Wird ignoriert, keine Ausführung

---

### Ungültiger Parameter

```bash
MailStatus=foo
```

→ Fallback auf `true` (Mail wird gesendet)

---

### Dienst hängt in `StopPending`

* Mail (bei `error`)
* danach Recovery-Task (falls gesetzt)

---

### Keine Berechtigung auf Service

* Status = `Unknown Unknown`
* Fehler im Log
* Mail optional

---

### Logdatei existiert nicht

* Wird automatisch erstellt
* UTF-8 BOM wird einmalig geschrieben

---

## Technische Details

* Logging:

  * UTF-8 mit BOM
  * CRLF (Windows-kompatibel)
* Task-Ausführung:

  * über `schtasks /run`
* Keine Parallelisierung
* Keine Retry-Logik
* Keine Timeouts für Tasks

---

## Einschränkungen

* Funktioniert nur unter Windows
* Keine native Mail-Funktion (nur via Task)
* Hängende Tasks können den Prozess blockieren
* Keine Exit-Codes für Monitoring integriert

---
