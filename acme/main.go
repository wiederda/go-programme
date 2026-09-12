package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version = "0.1.0"

	LEStaging    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	LEProduction = "https://acme-v02.api.letsencrypt.org/directory"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dns" {
		handleDNSCommand(os.Args[2:])
		return
	}

	var (
		directory   string
		email       string
		domain      string
		accountKey  string
		accountJSON string
		csrFile     string
		keyFile     string
		output      string
		production  bool
		acceptTOS   bool
		version     bool
	)

	flag.StringVar(&directory, "directory", LEStaging, "ACME Directory URL")
	flag.StringVar(&email, "email", "", "E-Mail-Adresse für das ACME-Konto")
	flag.StringVar(&domain, "domain", "", "Domain")
	flag.StringVar(&accountKey, "account-key", "account.key", "ACME Account Private Key")
	flag.StringVar(&accountJSON, "account-state", "account.json", "ACME Account Statusdatei")
	flag.StringVar(&csrFile, "csr", "", "CSR-Datei")
	flag.StringVar(&keyFile, "key", "", "Private Key zur CSR-Prüfung")
	flag.StringVar(&output, "output", "certificate.pem", "Ausgabedatei für das Zertifikat")
	flag.BoolVar(&production, "production", false, "Let's Encrypt Production verwenden")
	flag.BoolVar(&acceptTOS, "accept-tos", false, "ACME Terms of Service akzeptieren")
	flag.BoolVar(&version, "version", false, "Version anzeigen")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ACME Client %s\n\n", Version)
		fmt.Fprintln(os.Stderr, "Verwendung:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "DNS Provider:")
		fmt.Fprintln(os.Stderr, "  acme dns -list")
		fmt.Fprintln(os.Stderr, "  acme dns <anbieter>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Beispiel Windows:")
		fmt.Fprintln(os.Stderr, `  acme.exe -domain example.de -email admin@example.de -csr "C:\Zertifikate\example.de\request.csr" -key "C:\Zertifikate\example.de\private.key" -output "C:\Zertifikate\example.de\certificate.pem" -accept-tos`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Beispiel Linux:")
		fmt.Fprintln(os.Stderr, `  ./acme -domain example.de -email admin@example.de -csr "/etc/acme/example.de/request.csr" -key "/etc/acme/example.de/private.key" -output "/etc/acme/example.de/certificate.pem" -accept-tos`)
	}

	flag.Parse()
}

func handleDNSCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Verwendung:")
		fmt.Println("  acme dns -list")
		fmt.Println("  acme dns <anbieter>")
		return
	}

	if args[0] == "-list" {
		PrintDNSProviderList()
		return
	}

	providerName := args[0]

	provider, ok := GetDNSProvider(providerName)
	if !ok {
		fatal(fmt.Sprintf(
			"Unbekannter DNS Provider: %s\n\nVerwenden Sie \"acme dns -list\", um die verfügbaren Provider anzuzeigen.",
			providerName,
		))
	}

	PrintDNSProviderDetails(provider)
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "FEHLER:", message)
	os.Exit(1)
}
