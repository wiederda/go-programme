package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "1.0.0"

type ChainResult struct {
	Chain []string `json:"chain"`
}

func main() {
	orgPath := flag.String("org", "", "Pfad zum Ordner mit den .cer-Zertifikaten (z. B. ./org)")
	outPath := flag.String("out", "", "Pfad zur JSON-Ausgabedatei (z. B. ./chain.json)")
	showHelp := flag.Bool("help", false, "Zeigt diese Hilfe an")
	showVersion := flag.Bool("version", false, "Zeigt die Programmversion an")

	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVersion {
		fmt.Println("BuildCertificateChain", version)
		return
	}

	if *orgPath == "" || *outPath == "" {
		fmt.Println("❌ Fehler: Bitte --org und --out angeben.")
		fmt.Println("Tipp: --help für Hilfe.")
		os.Exit(1)
	}

	result, err := BuildCertificateChain(*orgPath)
	if err != nil {
		fmt.Printf("❌ Fehler beim Aufbau der Zertifikatskette: %v\n", err)
		os.Exit(2)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("❌ Fehler beim Serialisieren der JSON-Ausgabe: %v\n", err)
		os.Exit(3)
	}

	if err := os.WriteFile(*outPath, jsonData, 0644); err != nil {
		fmt.Printf("❌ Fehler beim Schreiben der JSON-Datei: %v\n", err)
		os.Exit(4)
	}

	fmt.Printf("✅ Zertifikatskette erfolgreich erstellt: %s\n", *outPath)
}

func printHelp() {
	fmt.Println(`
🔹 BuildCertificateChain - erstellt eine Zertifikatskette (Leaf → Intermediate → Root)

Verwendung:
  buildchain --org <Pfad_zum_Ordner> --out <Pfad_zur_JSON_Datei>

Optionen:
  --org <path>      Pfad zum Ordner mit den .cer-Dateien
  --out <path>      Zielpfad für die erzeugte JSON-Datei
  --help            Zeigt diese Hilfe an
  --version         Zeigt die aktuelle Version an

Beispiel:
  buildchain --org "./certs/org" --out "./output/chain.json"

Beschreibung:
  Dieses Tool lädt alle .cer-Dateien aus dem angegebenen Ordner,
  erkennt das Leaf-Zertifikat anhand des Ordnernamens und
  baut automatisch die Zertifikatskette auf.
  Die Reihenfolge wird als JSON-Datei gespeichert:
  {
    "chain": ["leaf.cer", "intermediate.cer", "root.cer"]
  }
`)
}

func BuildCertificateChain(orgPath string) (*ChainResult, error) {
	files, err := os.ReadDir(orgPath)
	if err != nil {
		return nil, fmt.Errorf("Ordner nicht lesbar: %w", err)
	}

	type certInfo struct {
		File string
		Cert *x509.Certificate
	}

	var certs []certInfo

	// 🔹 Zertifikate einlesen
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(strings.ToLower(f.Name()), ".cer") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(orgPath, f.Name()))
		if err != nil {
			continue
		}

		// PEM oder DER
		var cert *x509.Certificate
		block, _ := pem.Decode(data)
		if block != nil {
			cert, err = x509.ParseCertificate(block.Bytes)
		} else {
			cert, err = x509.ParseCertificate(data)
		}

		if err != nil {
			continue
		}

		certs = append(certs, certInfo{File: f.Name(), Cert: cert})
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("keine Zertifikate im Ordner gefunden")
	}

	// 🔹 Leaf bestimmen
	orgPathClean := strings.TrimRight(orgPath, string(os.PathSeparator))
	leafName := filepath.Base(orgPathClean) + ".cer"

	var leaf *certInfo
	for i := range certs {
		if strings.EqualFold(certs[i].File, leafName) {
			leaf = &certs[i]
			break
		}
	}
	if leaf == nil {
		return nil, fmt.Errorf("Leaf-Zertifikat %s nicht gefunden", leafName)
	}

	chain := []string{leaf.File}
	current := leaf.Cert
	visited := map[string]bool{}

	// 🔹 Chain aufbauen: Leaf → Intermediate → Root
	for {
		subject := normalizeDN(current.Subject.String())
		if visited[subject] {
			break
		}
		visited[subject] = true

		issuerDN := normalizeDN(current.Issuer.String())
		if subject == issuerDN {
			// self-signed Root erreicht
			break
		}

		var next *certInfo
		for i := range certs {
			if normalizeDN(certs[i].Cert.Subject.String()) == issuerDN {
				next = &certs[i]
				break
			}
		}
		if next == nil {
			break
		}

		chain = append(chain, next.File)
		current = next.Cert
	}

	return &ChainResult{Chain: chain}, nil
}

// normalizeDN: Wandelt DNs in Kleinschreibung und entfernt überflüssige Leerzeichen
func normalizeDN(dn string) string {
	dn = strings.ToLower(strings.TrimSpace(dn))
	parts := strings.Split(dn, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ",")
}
