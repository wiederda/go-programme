package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type AccountState struct {
	Directory string `json:"directory"`
	URL       string `json:"account_url"`
	Email     string `json:"email"`
}

func LoadOrCreateAccount(
	ctx context.Context,
	client *ACMEClient,
	keyPath string,
	statePath string,
	email string,
	acceptTOS bool,
) (*AccountState, error) {

	keyPath = filepath.Clean(keyPath)
	statePath = filepath.Clean(statePath)

	key, err := LoadOrCreateAccountKey(keyPath)
	if err != nil {
		return nil, err
	}

	client.AccountKey = key

	if state, err := LoadAccountState(statePath); err == nil {
		if state.Directory == client.DirectoryURL && state.URL != "" {
			fmt.Println("Vorhandenes ACME Account wird verwendet.")
			return state, nil
		}
	}

	// Zuerst versuchen wir, ein bereits existierendes Konto
	// anhand des Account Keys zu finden.
	accountURL, err := client.FindExistingAccount(ctx)
	if err == nil && accountURL != "" {
		state := &AccountState{
			Directory: client.DirectoryURL,
			URL:       accountURL,
			Email:     email,
		}

		if err := SaveAccountState(statePath, state); err != nil {
			return nil, err
		}

		return state, nil
	}

	if !acceptTOS {
		return nil, errors.New("ACME Terms of Service wurden nicht akzeptiert")
	}

	fmt.Println("Kein vorhandenes ACME Account gefunden.")
	fmt.Println("Erstelle neues ACME Account...")

	accountURL, err = client.CreateAccount(ctx, email)
	if err != nil {
		return nil, err
	}

	state := &AccountState{
		Directory: client.DirectoryURL,
		URL:       accountURL,
		Email:     email,
	}

	if err := SaveAccountState(statePath, state); err != nil {
		return nil, err
	}

	return state, nil
}

func LoadOrCreateAccountKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		key, err := ParseRSAPrivateKey(data)
		if err != nil {
			return nil, fmt.Errorf("ACME Account Key konnte nicht gelesen werden: %w", err)
		}

		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("ACME Account Key kann nicht gelesen werden: %w", err)
	}

	fmt.Println("Erzeuge neuen ACME Account Key...")

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("Account Key konnte nicht erzeugt werden: %w", err)
	}

	data, err = MarshalRSAPrivateKey(key)
	if err != nil {
		return nil, err
	}

	if err := EnsureParentDirectory(path); err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("Account Key konnte nicht gespeichert werden: %w", err)
	}

	return key, nil
}

func LoadAccountState(path string) (*AccountState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state AccountState

	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("Account Statusdatei ist ungültig: %w", err)
	}

	return &state, nil
}

func SaveAccountState(path string, state *AccountState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("Account Status konnte nicht erzeugt werden: %w", err)
	}

	if err := EnsureParentDirectory(path); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("Account Status konnte nicht gespeichert werden: %w", err)
	}

	return nil
}
