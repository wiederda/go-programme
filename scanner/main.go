package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	sha256simd "github.com/minio/sha256-simd"
)

type Config struct {
	DSN           string   `json:"dsn"`
	IncludeSuffix []string `json:"include_suffix"` // z.B. [".mp3", ".flac"]
}

type FileEntry struct {
	Name          string // relativ zum Verzeichnis (Local+Relativ)
	Size          int64
	Hash          string
	VerzeichnisID int
}

type Verzeichnis struct {
	ID      int
	Local   string
	Network string
	Relativ string
}

func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var cfg Config
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

func calculateFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256simd.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func fileHasSuffix(name string, suffixes []string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, s := range suffixes {
		if ext == strings.ToLower(s) {
			return true
		}
	}
	return false
}

func getFileEntries(base string, suffixes []string) ([]FileEntry, error) {
	var entries []FileEntry

	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if len(suffixes) > 0 && !fileHasSuffix(path, suffixes) {
			return nil
		}

		hash, err := calculateFileHash(path)
		if err != nil {
			return err
		}

		entries = append(entries, FileEntry{
			Name: path, // noch absolut, wird später relativ gesetzt
			Size: info.Size(),
			Hash: hash,
		})
		return nil
	})

	return entries, err
}

func syncFiles(ctx context.Context, db *sql.DB, files []FileEntry) error {
	// Map key = "verzeichnis_id::dateiname"
	existing := make(map[string]bool)

	rows, err := db.QueryContext(ctx, `
		SELECT verzeichnis_id, dateiname FROM dateien WHERE geloescht = false
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var vid int
		var name string
		if err := rows.Scan(&vid, &name); err == nil {
			key := fmt.Sprintf("%d::%s", vid, name)
			existing[key] = true
		}
	}

	for _, f := range files {
		key := fmt.Sprintf("%d::%s", f.VerzeichnisID, f.Name)
		existing[key] = false // Noch aktuell

		_, err := db.ExecContext(ctx, `
			INSERT INTO dateien (verzeichnis_id, dateiname, groesse, hash, geloescht)
			VALUES ($1, $2, $3, $4, false)
			ON CONFLICT (verzeichnis_id, dateiname) DO UPDATE
			SET groesse = EXCLUDED.groesse,
			    hash = EXCLUDED.hash,
			    geloescht = false
		`, f.VerzeichnisID, f.Name, f.Size, f.Hash)
		if err != nil {
			return err
		}
	}

	for key, stillExists := range existing {
		if stillExists {
			parts := strings.SplitN(key, "::", 2)
			if len(parts) != 2 {
				continue
			}
			vidStr := parts[0]
			name := parts[1]

			vid, err := strconv.Atoi(vidStr)
			if err != nil {
				continue
			}

			_, err = db.ExecContext(ctx, `
				UPDATE dateien
				SET geloescht = true
				WHERE verzeichnis_id = $1 AND dateiname = $2
			`, vid, name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var mode string

func init() {
	flag.StringVar(&mode, "mode", "local", "Scan-Modus: 'local' oder 'network'")
	flag.Parse()

	if mode != "local" && mode != "network" {
		fmt.Fprintln(os.Stderr, "Ungültiger Wert für --mode. Nur 'local' oder 'network' erlaubt.")
		os.Exit(1)
	}
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Laden der Konfiguration: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Verbinden zur Datenbank: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT id, local, network, relativ FROM verzeichnisse")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Lesen der Verzeichnisse: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var directories []Verzeichnis
	for rows.Next() {
		var v Verzeichnis
		if err := rows.Scan(&v.ID, &v.Local, &v.Network, &v.Relativ); err != nil {
			fmt.Fprintf(os.Stderr, "Fehler beim Lesen eines Verzeichniseintrags: %v\n", err)
			continue
		}
		directories = append(directories, v)
	}

	if len(directories) == 0 {
		fmt.Fprintln(os.Stderr, "Keine Verzeichnisse in der Tabelle 'verzeichnisse' gefunden.")
		os.Exit(1)
	}

	for _, v := range directories {
		var base string
		if mode == "network" {
			base = v.Network
		} else {
			base = v.Local
		}

		scanDir := filepath.Join(base, v.Relativ)
		fmt.Println("Scanne Verzeichnis:", scanDir)

		entries, err := getFileEntries(scanDir, cfg.IncludeSuffix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fehler beim Scannen von %s: %v\n", scanDir, err)
			continue
		}

		// Pfade relativ zum Local+Relativ berechnen und in entries setzen
		for i, entry := range entries {
			relBase := filepath.Join(v.Local, v.Relativ)
			relPath, err := filepath.Rel(relBase, entry.Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fehler bei Relativierung von %s: %v\n", entry.Name, err)
				continue
			}
			// Nur den relativen Pfad speichern, mit Slashes vereinheitlicht
			entries[i].Name = filepath.ToSlash(relPath)
			entries[i].VerzeichnisID = v.ID
		}

		err = syncFiles(ctx, db, entries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fehler beim Aktualisieren der Datenbank für %s: %v\n", scanDir, err)
		}
	}

	fmt.Println("Scan abgeschlossen.")
}
