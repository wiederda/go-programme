package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func ValidateCSRAndKey(csrPath, keyPath string) error {
	csrDER, err := ReadCSR(csrPath)
	if err != nil {
		return err
	}

	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return fmt.Errorf("CSR konnte nicht geparst werden: %w", err)
	}

	if err := csr.CheckSignature(); err != nil {
		return fmt.Errorf("CSR-Signatur ist ungültig: %w", err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("Private Key konnte nicht gelesen werden: %w", err)
	}

	key, err := ParsePrivateKey(keyData)
	if err != nil {
		return err
	}

	var keyPublic crypto.PublicKey

	switch k := key.(type) {
	case *rsa.PrivateKey:
		keyPublic = &k.PublicKey

	case *ecdsa.PrivateKey:
		keyPublic = &k.PublicKey

	default:
		return errors.New("nicht unterstützter Private-Key-Typ")
	}

	if !PublicKeysEqual(csr.PublicKey, keyPublic) {
		return errors.New("Private Key und CSR passen nicht zusammen")
	}

	return nil
}

func ReadCSR(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("CSR konnte nicht gelesen werden: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("CSR enthält keinen PEM-Block")
	}

	if block.Type != "CERTIFICATE REQUEST" &&
		block.Type != "NEW CERTIFICATE REQUEST" {
		return nil, fmt.Errorf(
			"ungültiger CSR PEM-Typ: %s",
			block.Type,
		)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("CSR konnte nicht gelesen werden: %w", err)
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf(
			"CSR-Signatur ist ungültig: %w",
			err,
		)
	}

	return block.Bytes, nil
}

func ParsePrivateKey(data []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Private Key enthält keinen PEM-Block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf(
				"RSA Private Key ist ungültig: %w",
				err,
			)
		}

		return key, nil

	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf(
				"EC Private Key ist ungültig: %w",
				err,
			)
		}

		return key, nil

	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf(
				"PKCS#8 Private Key ist ungültig: %w",
				err,
			)
		}

		switch k := key.(type) {
		case *rsa.PrivateKey:
			return k, nil

		case *ecdsa.PrivateKey:
			return k, nil

		default:
			return nil, errors.New(
				"PKCS#8 enthält einen nicht unterstützten Key-Typ",
			)
		}

	default:
		return nil, fmt.Errorf(
			"unbekannter Private-Key PEM-Typ: %s",
			block.Type,
		)
	}
}

func ParseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	key, err := ParsePrivateKey(data)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New(
			"ACME Account Key muss ein RSA Private Key sein",
		)
	}

	return rsaKey, nil
}

func MarshalRSAPrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	der := x509.MarshalPKCS1PrivateKey(key)

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: der,
	}), nil
}

func PublicKeysEqual(a, b crypto.PublicKey) bool {
	switch x := a.(type) {
	case *rsa.PublicKey:
		y, ok := b.(*rsa.PublicKey)
		if !ok {
			return false
		}

		return x.N.Cmp(y.N) == 0 &&
			x.E == y.E

	case *ecdsa.PublicKey:
		y, ok := b.(*ecdsa.PublicKey)
		if !ok {
			return false
		}

		if x.Curve != y.Curve {
			return false
		}

		return x.X.Cmp(y.X) == 0 &&
			x.Y.Cmp(y.Y) == 0
	}

	return false
}
