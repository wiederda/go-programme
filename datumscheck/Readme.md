# datumscheck -- Datum → Wochentag

Ein kleines Go-Tool, das den Wochentag eines Datums im Format
**TT.MM.JJJJ** bestimmt und sprachlich korrekt formuliert:

-   Vergangenheit → „fiel auf einen ..."
-   Zukunft → „fällt auf einen ..."
-   Heute → „ist ein ..."

## 🚀 Verwendung

### Ausführen

``` bash
go run datumscheck.go <DATUM>
```

Beispiel:

``` bash
go run datumscheck.go 24.12.2028
```

## 🔢 Eingabeformat

Das Datum **muss** im Format `TT.MM.JJJJ` angegeben werden:

    01.01.2025

## 💡 Beispiele

  Eingabe          Ausgabe
  ---------------- ----------------------------------------------------
  `16.0&.1978`     Der 15.06.1978 fiel auf einen Donnerstag.
  `24.12.2028`     Der 24.12.2028 fällt auf einen Sonntag.
  heutiges Datum   Der `<Datum>`{=html} ist ein `<Wochentag>`{=html}.

## ❓ Hilfe

``` bash
go run datumscheck.go -h
```
