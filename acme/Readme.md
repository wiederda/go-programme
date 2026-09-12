# ACME Client

Ein schlanker ACME-Client in Go zur automatisierten Ausstellung von TLS-Zertifikaten über ACME.

Der Client ist für **Windows und Linux** ausgelegt und verwendet standardmäßig **Let's Encrypt Staging**, damit Tests durchgeführt werden können, ohne versehentlich produktive Zertifikate anzufordern.

---

## Funktionen

* ACME v2
* Let's Encrypt als Certificate Authority
* Let's Encrypt Staging als Standard
* Let's Encrypt Production über `-production`
* DNS-01 Challenge
* Wiederverwendung vorhandener Private Keys
* Wiederverwendung vorhandener CSRs
* Prüfung, ob Private Key und CSR zusammengehören
* Verwaltung eines separaten ACME-Accounts
* Speicherung des ACME-Account-Status
* DNS-Provider-Unterstützung
* Anzeige verfügbarer DNS-Provider
* Anzeige der Anforderungen eines DNS-Providers
* Windows- und Linux-Pfade
* Unterstützung normaler Windows-Pfade und UNC-Pfade

---

## Voraussetzungen

Für die Ausstellung eines Zertifikats werden benötigt:

* eine eigene bzw. kontrollierbare Domain
* ein Private Key
* eine dazugehörige CSR
* eine E-Mail-Adresse für das ACME-Konto
* ein unterstützter DNS-Provider für DNS-01

Der Private Key und die CSR werden vom ACME-Client **nicht automatisch erzeugt**.

Sie können beispielsweise aus einer bestehenden Zertifikatsverwaltung stammen.

---

## Installation

Der Client wird als einzelne ausführbare Datei bereitgestellt.

Unter Windows beispielsweise:

```text
acme.exe
```

Unter Linux:

```text
acme
```

Die ausführbare Datei kann direkt aus einem Terminal verwendet werden.

---

## Hilfe

Die allgemeine Hilfe wird mit:

```text
acme -h
```

angezeigt.

Die Version:

```text
acme -version
```

---

# Zertifikat anfordern

Ein typischer Aufruf sieht beispielsweise so aus:

```text
acme -domain example.de -email admin@example.de -csr "C:\Zertifikate\example.de\request.csr" -key "C:\Zertifikate\example.de\private.key" -output "C:\Zertifikate\example.de\certificate.pem" -accept-tos
```

Unter Linux:

```text
./acme -domain example.de -email admin@example.de -csr "/etc/acme/example.de/request.csr" -key "/etc/acme/example.de/private.key" -output "/etc/acme/example.de/certificate.pem" -accept-tos
```

Ohne weitere Angabe verwendet der Client **Let's Encrypt Staging**.

Das ist für die Entwicklung und zum Testen vorgesehen.

---

## Wichtige Parameter

### `-domain`

Die Domain, für die das Zertifikat ausgestellt werden soll.

Beispiel:

```text
-domain example.de
```

### `-email`

E-Mail-Adresse für das ACME-Konto.

```text
-email admin@example.de
```

### `-csr`

Pfad zur vorhandenen CSR-Datei.

```text
-csr "C:\Zertifikate\example.de\request.csr"
```

### `-key`

Pfad zum Private Key, der zur CSR gehört.

```text
-key "C:\Zertifikate\example.de\private.key"
```

Der Client überprüft, ob CSR und Private Key zusammengehören.

### `-output`

Ausgabedatei für das ausgestellte Zertifikat.

Standard:

```text
certificate.pem
```

Beispiel:

```text
-output "C:\Zertifikate\example.de\certificate.pem"
```

### `-account-key`

Private Key des ACME-Accounts.

Standard:

```text
account.key
```

### `-account-state`

Datei mit dem gespeicherten ACME-Account-Status.

Standard:

```text
account.json
```

### `-accept-tos`

Bestätigt die ACME Terms of Service.

Ohne diese Option wird keine Zertifikatsanforderung durchgeführt.

### `-production`

Verwendet die Produktionsumgebung von Let's Encrypt.

Ohne diese Option wird Staging verwendet.

---

# Staging und Production

Während der Entwicklung sollte grundsätzlich zunächst Staging verwendet werden.

Standard:

```text
Let's Encrypt Staging
```

Für ein produktives Zertifikat muss ausdrücklich:

```text
-production
```

angegeben werden.

Damit soll verhindert werden, dass während der Entwicklung unbeabsichtigt produktive Zertifikate angefordert oder Rate-Limits belastet werden.

---

# DNS-01

Der Client verwendet für die Domainvalidierung DNS-01.

Dabei muss für die Domain ein TXT-Eintrag unter:

```text
_acme-challenge.<domain>
```

gesetzt werden.

Beispiel:

```text
_acme-challenge.example.de
```

Der dafür benötigte TXT-Wert wird vom ACME-Verfahren vorgegeben.

Der DNS-Provider muss die Verwaltung entsprechender TXT-Einträge ermöglichen.

---

# DNS Provider

DNS-Provider werden über den Client angezeigt und später automatisch für die DNS-01-Challenge verwendet.

## Verfügbare Provider anzeigen

```text
acme dns -list
```

Beispiel:

```text
DNS Provider
============

ipv64
```

## Informationen zu einem Provider

Mit:

```text
acme dns ipv64
```

werden die Anforderungen des Providers angezeigt.

Beispiel:

```text
DNS Provider: ipv64
Beschreibung: IPv64 DNS API

Authentifizierung:
  API Token

Erforderlich:
  token        API Token

Unterstützt:
  DNS-01          ja
  TXT setzen      ja
  TXT löschen     ja
```

Die Anforderungen können sich je nach DNS-Provider unterscheiden.

Ein Provider kann beispielsweise ein API-Token benötigen, während ein anderer mehrere Zugangsdaten oder zusätzliche Angaben benötigt.

---

# IPv64

IPv64 ist der erste DNS-Provider, der unterstützt werden soll.

Für IPv64 ist eine DNS-API vorhanden, über die der für DNS-01 benötigte TXT-Eintrag automatisch verwaltet werden kann.

Vorgesehen ist die Verwendung eines IPv64 API-Tokens.

Damit soll der Ablauf später vollständig automatisiert werden:

```text
ACME Client
    │
    ├── Challenge von Let's Encrypt
    │
    ├── TXT-Eintrag bei IPv64 setzen
    │
    ├── DNS-Validierung
    │
    ├── TXT-Eintrag wieder entfernen
    │
    └── Zertifikat abrufen
```

---

# ACME Account

Für ACME wird ein eigener Account benötigt.

Der Account ist unabhängig vom Private Key des Zertifikats.

Standardmäßig werden dafür:

```text
account.key
account.json
```

verwendet.

Der Account-Key wird bei Bedarf automatisch erstellt.

Der gespeicherte Account kann bei späteren Zertifikatsanforderungen wiederverwendet werden.

---

# Dateipfade

Der Client unterstützt Windows und Linux.

## Windows

Normale Windows-Pfade können verwendet werden:

```text
C:\Zertifikate\example.de\private.key
C:\Zertifikate\example.de\request.csr
C:\Zertifikate\example.de\certificate.pem
```

Auch UNC-Pfade sind möglich:

```text
\\Server\Freigabe\Zertifikate\example.de\private.key
```

## Linux

Beispiel:

```text
/etc/acme/example.de/private.key
/etc/acme/example.de/request.csr
/etc/acme/example.de/certificate.pem
```

Es ist keine Umwandlung von Windows-Pfaden in Linux-ähnliche Schreibweisen notwendig.

---

# Sicherheit

Private Keys und API-Tokens müssen geschützt werden.

Insbesondere sollten folgende Dateien nicht in öffentliche oder gemeinsam genutzte Repositories gelangen:

```text
*.key
account.json
```

Auch Zugangsdaten der DNS-Provider dürfen nicht veröffentlicht werden.

Für Tests sollte die Let's Encrypt Staging-Umgebung verwendet werden.

---

# Geplante Erweiterungen

Für zukünftige Versionen sind unter anderem vorgesehen:

* weitere DNS-Provider
* vollständige automatische DNS-01-Verwaltung
* weitere Möglichkeiten zur Zertifikatsverwaltung
* zusätzliche ACME-Funktionen

Die Liste kann sich während der Entwicklung noch ändern.
