package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	Version = "0.1.0"

	LEStaging    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	LEProduction = "https://acme-v02.api.letsencrypt.org/directory"
)

func main() {
	var (
		dnsProvider string
		dnsServers  string

		directory   string
		email       string
		domain      string
		accountKey  string
		accountJSON string
		csrFile     string
		output      string
		production  bool
		acceptTOS   bool
		version     bool
	)

	flag.StringVar(
		&dnsProvider,
		"dns",
		"",
		"DNS Provider",
	)

	flag.StringVar(
		&dnsServers,
		"dns-server",
		"",
		"DNS Resolver, mehrere mit Komma trennen",
	)

	flag.StringVar(
		&directory,
		"directory",
		LEStaging,
		"ACME Directory URL",
	)

	flag.StringVar(
		&email,
		"email",
		"",
		"E-Mail-Adresse für das ACME-Konto",
	)

	flag.StringVar(
		&domain,
		"domain",
		"",
		"Domain",
	)

	flag.StringVar(
		&accountKey,
		"account-key",
		"account.key",
		"ACME Account Private Key",
	)

	flag.StringVar(
		&accountJSON,
		"account-state",
		"account.json",
		"ACME Account Statusdatei",
	)

	flag.StringVar(
		&csrFile,
		"csr",
		"",
		"CSR-Datei",
	)

	flag.StringVar(
		&output,
		"out",
		"certificate.pem",
		"Ausgabedatei für das Zertifikat",
	)

	flag.BoolVar(
		&production,
		"production",
		false,
		"Let's Encrypt Production verwenden",
	)

	flag.BoolVar(
		&acceptTOS,
		"accept-tos",
		false,
		"ACME Terms of Service akzeptieren",
	)

	flag.BoolVar(
		&version,
		"version",
		false,
		"Version anzeigen",
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ACME Client %s\n\n", Version)

		fmt.Fprintln(os.Stderr, "Verwendung:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "DNS Provider:")
		fmt.Fprintln(os.Stderr, "  acme -dns list")
		fmt.Fprintln(os.Stderr, "  acme -dns <anbieter>")
		fmt.Fprintln(os.Stderr, "  acme -dns <anbieter> test <domain>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Zertifikat:")
		fmt.Fprintln(os.Stderr, "  acme -dns <anbieter> -domain <domain> -email <email> -csr <datei> -out <datei>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Optionen:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Beispiel Windows:")
		fmt.Fprintln(
			os.Stderr,
			`  acme.exe -dns ipv64 -domain example.de -email admin@example.de -csr "C:\Zertifikate\example.de\request.csr" -out "C:\Zertifikate\example.de\certificate.pem" -accept-tos`,
		)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Beispiel Linux:")
		fmt.Fprintln(
			os.Stderr,
			`  ./acme -dns ipv64 -domain example.de -email admin@example.de -csr "/etc/acme/example.de/request.csr" -out "/etc/acme/example.de/certificate.pem" -accept-tos`,
		)
	}

	flag.Parse()

	if version {
		fmt.Println(Version)
		return
	}

	if dnsProvider == "" {
		fatal("DNS Provider fehlt: -dns")
	}

	// ------------------------------------------------------------
	// DNS Provider Verwaltung
	// ------------------------------------------------------------

	dnsArgs := flag.Args()

	if dnsProvider == "list" {
		if len(dnsArgs) > 0 {
			fatal("Zusätzliche Parameter sind bei \"-dns list\" nicht erlaubt")
		}

		PrintDNSProviderList()
		return
	}

	provider, ok := GetDNSProvider(dnsProvider)
	if !ok {
		fatal(fmt.Sprintf(
			"Unbekannter DNS Provider: %s\n\nVerwenden Sie \"acme -dns list\", um die verfügbaren Provider anzuzeigen.",
			dnsProvider,
		))
	}

	// "acme -dns ipv64 test domain"
	if len(dnsArgs) > 0 {
		if dnsArgs[0] != "test" {
			fatal(fmt.Sprintf(
				"Unbekannte DNS-Aktion für Provider %q: %s",
				dnsProvider,
				dnsArgs[0],
			))
		}

		if len(dnsArgs) != 2 {
			fatal("Verwendung: acme -dns <anbieter> test <domain>")
		}

		if domain != "" {
			fatal("Bei einem DNS Provider Test darf -domain nicht verwendet werden")
		}

		if err := TestDNSProvider(provider, dnsArgs[1]); err != nil {
			fatal(err.Error())
		}

		return
	}

	// "acme -dns ipv64" ohne Zertifikatsparameter:
	// Provider-Informationen anzeigen.
	if domain == "" &&
		email == "" &&
		csrFile == "" &&
		output == "certificate.pem" {

		PrintDNSProviderDetails(provider)
		return
	}

	// ------------------------------------------------------------
	// Zertifikatsanforderung
	// ------------------------------------------------------------

	if production {
		directory = LEProduction
	}

	if domain == "" {
		fatal("Domain fehlt: -domain")
	}

	if email == "" {
		fatal("E-Mail-Adresse fehlt: -email")
	}

	if csrFile == "" {
		fatal("CSR-Datei fehlt: -csr")
	}

	if !acceptTOS {
		fatal("Die ACME Terms of Service müssen mit -accept-tos akzeptiert werden")
	}

	paths := NewPaths(
		accountKey,
		accountJSON,
		csrFile,
		output,
	)

	fmt.Println("ACME Client", Version)
	fmt.Println()
	fmt.Println("Directory  :", directory)
	fmt.Println("DNS Provider:", provider.Name())
	fmt.Println("Domain     :", domain)
	fmt.Println("CSR        :", paths.CSR)
	fmt.Println("Zertifikat :", paths.Output)

	resolvers := ParseDNSResolvers(dnsServers)

	fmt.Println("DNS Server :", joinStrings(resolvers))
	fmt.Println()

	ctx := context.Background()

	// ------------------------------------------------------------
	// ACME Client
	// ------------------------------------------------------------

	client, err := NewACMEClient(
		ctx,
		directory,
		paths.AccountKey,
	)
	if err != nil {
		fatal(err.Error())
	}

	fmt.Println("ACME Directory wird geladen...")

	if err := client.LoadDirectory(ctx); err != nil {
		fatal(err.Error())
	}

	// ------------------------------------------------------------
	// CSR prüfen
	// ------------------------------------------------------------

	csrDER, err := ReadCSR(paths.CSR)
	if err != nil {
		fatal(fmt.Sprintf(
			"CSR konnte nicht verarbeitet werden: %v",
			err,
		))
	}

	fmt.Println("CSR wurde erfolgreich gelesen.")
	fmt.Println()

	// ------------------------------------------------------------
	// ACME Account
	// ------------------------------------------------------------

	account, err := LoadOrCreateAccount(
		ctx,
		client,
		paths.AccountKey,
		paths.AccountState,
		email,
		acceptTOS,
	)
	if err != nil {
		fatal(err.Error())
	}

	client.AccountURL = account.URL

	fmt.Println("ACME Account:", account.URL)
	fmt.Println()

	// ------------------------------------------------------------
	// Order
	// ------------------------------------------------------------

	order, err := client.CreateOrder(
		ctx,
		[]string{domain},
	)
	if err != nil {
		fatal(err.Error())
	}

	fmt.Println("Order erstellt.")
	fmt.Println("Order URL:", order.URL)
	fmt.Println()

	if len(order.Authorizations) == 0 {
		fatal("ACME Order enthält keine Authorization")
	}

	// ------------------------------------------------------------
	// Authorization
	// ------------------------------------------------------------

	authz, err := client.GetAuthorization(
		ctx,
		order.Authorizations[0],
	)
	if err != nil {
		fatal(err.Error())
	}

	challenge, err := FindDNS01Challenge(authz)
	if err != nil {
		fatal(err.Error())
	}

	txtValue, err := client.DNS01Value(challenge.Token)
	if err != nil {
		fatal(err.Error())
	}

	fqdn := DNS01FQDN(domain)

	fmt.Println("------------------------------------------------------------")
	fmt.Println("DNS-01 Challenge")
	fmt.Println("------------------------------------------------------------")
	fmt.Println()
	fmt.Println("Provider:", provider.Name())
	fmt.Println("Name    :", fqdn)
	fmt.Println("Wert    :", txtValue)
	fmt.Println()

	// ------------------------------------------------------------
	// TXT setzen
	// ------------------------------------------------------------

	fmt.Println("TXT-Eintrag wird gesetzt...")

	if err := provider.Present(
		ctx,
		fqdn,
		txtValue,
	); err != nil {
		fatal(fmt.Sprintf(
			"TXT-Eintrag konnte nicht gesetzt werden: %v",
			err,
		))
	}

	fmt.Println("TXT-Eintrag wurde gesetzt.")
	fmt.Println()

	// Ab hier muss der TXT-Eintrag auf jeden Fall wieder entfernt
	// werden, auch wenn die ACME-Validierung später fehlschlägt.
	cleanup := true

	defer func() {
		if !cleanup {
			return
		}

		fmt.Println()
		fmt.Println("TXT-Eintrag wird aufgeräumt...")

		if err := provider.Cleanup(
			context.Background(),
			fqdn,
			txtValue,
		); err != nil {
			fmt.Fprintln(
				os.Stderr,
				"FEHLER: TXT-Eintrag konnte nicht gelöscht werden:",
				err,
			)
			return
		}

		fmt.Println("TXT-Eintrag wurde gelöscht.")
	}()

	// ------------------------------------------------------------
	// Öffentliche DNS-Auflösung prüfen
	// ------------------------------------------------------------

	if err := WaitForTXT(
		ctx,
		fqdn,
		txtValue,
		resolvers,
		120*time.Second,
		5*time.Second,
	); err != nil {
		fatal(fmt.Sprintf(
			"DNS-01 TXT-Eintrag ist nicht öffentlich erreichbar: %v",
			err,
		))
	}

	fmt.Println()

	// ------------------------------------------------------------
	// Challenge aktivieren
	// ------------------------------------------------------------

	if err := client.AcceptChallenge(
		ctx,
		challenge.URL,
	); err != nil {
		fatal(err.Error())
	}

	fmt.Println("Challenge wurde aktiviert.")
	fmt.Println("Warte auf Validierung...")
	fmt.Println()

	if err := client.WaitAuthorization(
		ctx,
		authz.URL,
	); err != nil {
		fatal(err.Error())
	}

	fmt.Println("DNS-01 Challenge erfolgreich validiert.")

	// Der TXT-Eintrag wird nach erfolgreicher Validierung
	// nicht mehr benötigt.
	if err := provider.Cleanup(
		ctx,
		fqdn,
		txtValue,
	); err != nil {
		fatal(fmt.Sprintf(
			"TXT-Eintrag konnte nach erfolgreicher Validierung nicht gelöscht werden: %v",
			err,
		))
	}

	cleanup = false

	fmt.Println("TXT-Eintrag wurde gelöscht.")
	fmt.Println()

	// ------------------------------------------------------------
	// Order finalisieren
	// ------------------------------------------------------------

	if err := client.FinalizeOrder(
		ctx,
		order.FinalizeURL,
		csrDER,
	); err != nil {
		fatal(err.Error())
	}

	fmt.Println("Order finalisiert.")
	fmt.Println("Warte auf Zertifikatsausstellung...")
	fmt.Println()

	finalOrder, err := client.WaitOrder(
		ctx,
		order.URL,
	)
	if err != nil {
		fatal(err.Error())
	}

	if finalOrder.CertificateURL == "" {
		fatal(
			"ACME Order wurde abgeschlossen, enthält aber keine Certificate URL",
		)
	}

	// ------------------------------------------------------------
	// Zertifikat laden
	// ------------------------------------------------------------

	certificate, err := client.FetchCertificate(
		ctx,
		finalOrder.CertificateURL,
	)
	if err != nil {
		fatal(err.Error())
	}

	if err := WriteCertificate(
		paths.Output,
		certificate,
	); err != nil {
		fatal(err.Error())
	}

	fmt.Println()
	fmt.Println("Zertifikat erfolgreich gespeichert:")
	fmt.Println(paths.Output)
}

func TestDNSProvider(
	provider DNSProvider,
	domain string,
) error {
	ctx := context.Background()

	fqdn := DNS01FQDN(domain)
	value := "vbx-acme-test"

	fmt.Println()
	fmt.Println("DNS Provider Test")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("Provider:", provider.Name())
	fmt.Println("Domain  :", domain)
	fmt.Println("Name    :", fqdn)
	fmt.Println("Wert    :", value)
	fmt.Println()

	fmt.Println("TXT-Eintrag wird gesetzt...")

	if err := provider.Present(
		ctx,
		fqdn,
		value,
	); err != nil {
		return fmt.Errorf(
			"TXT-Eintrag konnte nicht gesetzt werden: %w",
			err,
		)
	}

	fmt.Println("TXT-Eintrag wurde gesetzt.")
	fmt.Println()

	fmt.Println("Der Eintrag kann jetzt geprüft werden.")
	fmt.Println()

	fmt.Println("TXT-Eintrag wird nach dem Test wieder gelöscht.")

	if err := provider.Cleanup(
		ctx,
		fqdn,
		value,
	); err != nil {
		return fmt.Errorf(
			"TXT-Eintrag konnte nicht gelöscht werden: %w",
			err,
		)
	}

	fmt.Println("TXT-Eintrag wurde gelöscht.")
	fmt.Println()
	fmt.Println("DNS Provider Test erfolgreich.")

	return nil
}

func joinStrings(values []string) string {
	if len(values) == 0 {
		return ""
	}

	result := values[0]

	for i := 1; i < len(values); i++ {
		result += ", " + values[i]
	}

	return result
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "FEHLER:", message)
	os.Exit(1)
}
