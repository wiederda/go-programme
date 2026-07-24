# 🔗 BuildCertificateChain

Ein leichtgewichtiges **Go-Tool**, das automatisch eine **Zertifikatskette (Leaf → Intermediate → Root)** aus einem Ordner mit `.cer`-Dateien aufbaut und das Ergebnis als **JSON-Datei** exportiert.

---

## 🧭 Übersicht

`BuildCertificateChain` durchsucht einen Ordner nach Zertifikaten, identifiziert automatisch das **Leaf-Zertifikat** (anhand des Ordnernamens) und rekonstruiert daraus die komplette Kette bis zur Root-CA.

Das Ergebnis wird als einfache, maschinenlesbare JSON-Datei gespeichert – ideal zur weiteren Verarbeitung in CI/CD-Pipelines, Security-Checks oder Zertifikats-Audits.

---

## 🚀 Funktionen

✅ Erkennt das **Leaf-Zertifikat** anhand des Ordnernamens (z. B. `org/org.cer`)  
✅ Baut automatisch die **Zertifikatskette** auf  
✅ Unterstützt **PEM** und **DER**-Formate  
✅ Gibt das Ergebnis als **JSON** aus  
✅ Robuste Fehlerbehandlung mit klaren Konsolenmeldungen  
✅ Plattformunabhängig (Windows / Linux / macOS)

---

./check_chain -org ./certs/org -out ./chain.json

---

## Beispielausgabe (`chain.json`):

```json
{
  "chain": [
    "org.cer",
    "intermediate.cer",
    "root.cer"
  ]
}


