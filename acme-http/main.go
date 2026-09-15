package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "0.1.0"

	StagingDirectory    = "https://acme-staging-v02.api.letsencrypt.org/directory"
	ProductionDirectory = "https://acme-v02.api.letsencrypt.org/directory"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		domain = flag.String(
			"domain",
			"",
			"Domain für das Zertifikat",
		)

		email = flag.String(
			"email",
			"",
			"E-Mail-Adresse für den ACME-Account",
		)

		csr = flag.String(
			"csr",
			"",
			"Pfad zum vorhandenen CSR",
		)

		output = flag.String(
			"out",
			"certificate.pem",
			"Ausgabedatei für das Zertifikat",
		)

		accountKey = flag.String(
			"account-key",
			"account.key",
			"Pfad zum ACME Account Private Key",
		)

		accountState = flag.String(
			"account-state",
			"account.json",
			"Pfad zum ACME Account State",
		)

		directory = flag.String(
			"directory",
			StagingDirectory,
			"ACME Directory URL",
		)

		production = flag.Bool(
			"production",
			false,
			"Let's Encrypt Production verwenden",
		)

		acceptTOS = flag.Bool(
			"accept-tos",
			false,
			"Let's Encrypt Nutzungsbedingungen akzeptieren",
		)

		httpListen = flag.String(
			"http-listen",
			"0.0.0.0",
			"HTTP-01 Listener-Adresse",
		)

		httpPort = flag.Int(
			"http-port",
			80,
			"HTTP-01 Listener-Port",
		)

		showVersion = flag.Bool(
			"version",
			false,
			"Version anzeigen",
		)
	)

	flag.Usage = func() {
		out := flag.CommandLine.Output()

		fmt.Fprintln(out, "acme-http - ACME HTTP-01 Client")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Verwendung:")
		fmt.Fprintln(out)
		fmt.Fprintln(out,
			"  acme-http -domain example.com -email admin@example.com -csr cert.csr -out cert.pem -accept-tos",
		)
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Optionen:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("acme-http %s\n", Version)
		return nil
	}

	if *production {
		*directory = ProductionDirectory
	}

	if strings.TrimSpace(*domain) == "" {
		return errors.New(
			"-domain muss angegeben werden",
		)
	}

	if strings.TrimSpace(*email) == "" {
		return errors.New(
			"-email muss angegeben werden",
		)
	}

	if strings.TrimSpace(*csr) == "" {
		return errors.New(
			"-csr muss angegeben werden",
		)
	}

	if strings.TrimSpace(*output) == "" {
		return errors.New(
			"-out darf nicht leer sein",
		)
	}

	if *httpPort < 1 || *httpPort > 65535 {
		return fmt.Errorf(
			"ungültiger HTTP-Port: %d",
			*httpPort,
		)
	}

	if !*acceptTOS {
		return errors.New(
			"die Let's Encrypt Nutzungsbedingungen müssen mit -accept-tos akzeptiert werden",
		)
	}

	listen := *httpListen + ":" + strconv.Itoa(*httpPort)

	paths := NewPaths(
		*accountKey,
		*accountState,
		*csr,
		*output,
	)

	ctx := context.Background()

	fmt.Println("acme-http")
	fmt.Println("---------")
	fmt.Printf("Domain:       %s\n", *domain)
	fmt.Printf("E-Mail:       %s\n", *email)
	fmt.Printf("CSR:          %s\n", paths.CSR)
	fmt.Printf("Zertifikat:   %s\n", paths.Output)
	fmt.Printf("Directory:    %s\n", *directory)
	fmt.Printf("HTTP-01:      %s\n", listen)
	fmt.Println()

	fmt.Println("CSR wird geprüft...")

	csrDER, err := ReadCSR(paths.CSR)
	if err != nil {
		return err
	}

	fmt.Println("CSR ist gültig.")

	client, err := NewACMEClient(
		ctx,
		*directory,
		paths.AccountKey,
	)
	if err != nil {
		return fmt.Errorf(
			"ACME Client konnte nicht erstellt werden: %w",
			err,
		)
	}

	fmt.Println("ACME Directory wird geladen...")

	if err := client.LoadDirectory(ctx); err != nil {
		return fmt.Errorf(
			"ACME Directory konnte nicht geladen werden: %w",
			err,
		)
	}

	fmt.Println("ACME Account wird geladen...")

	state, err := LoadOrCreateAccount(
		ctx,
		client,
		paths.AccountKey,
		paths.AccountState,
		*email,
		*acceptTOS,
	)
	if err != nil {
		return fmt.Errorf(
			"ACME Account konnte nicht geladen/erstellt werden: %w",
			err,
		)
	}

	client.AccountURL = state.URL

	fmt.Printf(
		"ACME Account: %s\n",
		state.URL,
	)

	fmt.Println("Order wird erstellt...")

	order, err := client.CreateOrder(
		ctx,
		[]string{*domain},
	)
	if err != nil {
		return fmt.Errorf(
			"ACME Order konnte nicht erstellt werden: %w",
			err,
		)
	}

	if len(order.Authorizations) == 0 {
		return errors.New(
			"ACME Order enthält keine Authorizations",
		)
	}

	fmt.Printf(
		"Order enthält %d Authorization(s).\n",
		len(order.Authorizations),
	)

	for index, authorizationURL := range order.Authorizations {
		fmt.Printf(
			"\nAuthorization %d/%d\n",
			index+1,
			len(order.Authorizations),
		)

		authz, err := client.GetAuthorization(
			ctx,
			authorizationURL,
		)
		if err != nil {
			return fmt.Errorf(
				"Authorization konnte nicht geladen werden: %w",
				err,
			)
		}

		fmt.Printf(
			"Identifier: %s\n",
			authz.Identifier.Value,
		)

		challenge, err := FindHTTP01Challenge(authz)
		if err != nil {
			return err
		}

		if challenge.Token == "" {
			return errors.New(
				"HTTP-01 Challenge enthält keinen Token",
			)
		}

		thumbprint, err := JWKThumbprint(
			&client.AccountKey.PublicKey,
		)
		if err != nil {
			return fmt.Errorf(
				"JWK Thumbprint konnte nicht erzeugt werden: %w",
				err,
			)
		}

		keyAuthorization := HTTP01KeyAuthorization(
			challenge.Token,
			thumbprint,
		)

		fmt.Printf(
			"Challenge URL: http://%s%s%s\n",
			*domain,
			http01PathPrefix,
			challenge.Token,
		)

		fmt.Println("HTTP-01 Server wird gestartet...")

		server, err := NewHTTPChallengeServer(
			listen,
			challenge.Token,
			keyAuthorization,
		)
		if err != nil {
			return err
		}

		fmt.Printf(
			"HTTP-01 Server läuft auf %s\n",
			server.Address(),
		)

		serverRunning := true

		defer func() {
			if serverRunning {
				stopCtx, cancel := context.WithTimeout(
					context.Background(),
					5*time.Second,
				)
				defer cancel()

				_ = server.Stop(stopCtx)
			}
		}()

		fmt.Println("HTTP-01 Challenge wird aktiviert...")

		if err := client.AcceptChallenge(
			ctx,
			challenge.URL,
		); err != nil {
			return fmt.Errorf(
				"HTTP-01 Challenge konnte nicht aktiviert werden: %w",
				err,
			)
		}

		fmt.Println(
			"Warte auf Validierung durch Let's Encrypt...",
		)

		if err := client.WaitAuthorization(
			ctx,
			authorizationURL,
		); err != nil {
			return fmt.Errorf(
				"HTTP-01 Authorization fehlgeschlagen: %w",
				err,
			)
		}

		fmt.Println("Authorization erfolgreich.")

		stopCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

		if err := server.Stop(stopCtx); err != nil {
			cancel()

			return fmt.Errorf(
				"HTTP-01 Server konnte nicht beendet werden: %w",
				err,
			)
		}

		cancel()
		serverRunning = false

		if err := server.CheckError(); err != nil {
			return err
		}

		fmt.Println("HTTP-01 Server beendet.")
	}

	fmt.Println()
	fmt.Println("Order wird finalisiert...")

	if err := client.FinalizeOrder(
		ctx,
		order.FinalizeURL,
		csrDER,
	); err != nil {
		return fmt.Errorf(
			"Order konnte nicht finalisiert werden: %w",
			err,
		)
	}

	fmt.Println(
		"Warte auf Ausstellung des Zertifikats...",
	)

	order, err = client.WaitOrder(
		ctx,
		order.URL,
	)
	if err != nil {
		return fmt.Errorf(
			"Order konnte nicht abgeschlossen werden: %w",
			err,
		)
	}

	fmt.Println("Zertifikat wurde ausgestellt.")

	fmt.Println("Zertifikat wird geladen...")

	certificate, err := client.FetchCertificate(
		ctx,
		order.CertificateURL,
	)
	if err != nil {
		return fmt.Errorf(
			"Zertifikat konnte nicht geladen werden: %w",
			err,
		)
	}

	if err := EnsureParentDirectory(paths.Output); err != nil {
		return err
	}

	if err := WriteCertificate(
		paths.Output,
		certificate,
	); err != nil {
		return fmt.Errorf(
			"Zertifikat konnte nicht gespeichert werden: %w",
			err,
		)
	}

	fmt.Println()
	fmt.Printf(
		"Zertifikat gespeichert: %s\n",
		paths.Output,
	)

	return nil
}
