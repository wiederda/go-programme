package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite" // Verwende modernc.org/sqlite anstelle von mattn/go-sqlite3
)

// Funktion zur Anzeige von fehlenden Buch-IDs
func showMissingBookIDs(metadataDBPath, appDBPath string) {
	conn, err := sql.Open("sqlite", metadataDBPath) // Verwende "sqlite" anstelle von "sqlite3"
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS app_db", appDBPath))
	if err != nil {
		log.Fatal(err)
	}

	rows, err := conn.Query(`
		SELECT id FROM books
		WHERE id NOT IN (SELECT book_id FROM app_db.book_shelf_link)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	foundBooks := false
	fmt.Println("Fehlende Buch-IDs:")
	for rows.Next() {
		foundBooks = true
		var bookID int
		if err := rows.Scan(&bookID); err != nil {
			log.Fatal(err)
		}
		fmt.Println(bookID)
	}

	if !foundBooks {
		fmt.Println("Keine fehlenden Buch-IDs gefunden.")
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

// Funktion zum Hinzufügen unsortierter Bücher zum Regal
func assignUnsortedBooks(metadataDBPath, appDBPath string, unsortedShelfID int) {
	conn, err := sql.Open("sqlite", metadataDBPath) // Verwende "sqlite" anstelle von "sqlite3"
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS app_db", appDBPath))
	if err != nil {
		log.Fatal(err)
	}

	var shelfID int
	err = conn.QueryRow("SELECT id FROM app_db.shelf WHERE id = ?", unsortedShelfID).Scan(&shelfID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("Fehler: Regal mit ID %d existiert nicht in app.db!\n", unsortedShelfID)
		} else {
			log.Fatal(err)
		}
		return
	}

	rows, err := conn.Query(`
		SELECT books.id, books.title
		FROM books
		LEFT JOIN (SELECT DISTINCT book_id FROM app_db.book_shelf_link) AS bsl ON books.id = bsl.book_id
		WHERE bsl.book_id IS NULL
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	foundBooks := false
	for rows.Next() {
		foundBooks = true
		var bookID int
		var title string
		if err := rows.Scan(&bookID, &title); err != nil {
			log.Fatal(err)
		}

		_, err := conn.Exec(`
			INSERT INTO app_db.book_shelf_link (book_id, shelf, date_added)
			VALUES (?, ?, datetime('now'))`,
			bookID, unsortedShelfID,
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Buch '%s' wurde dem Regal '_unsortiert' zugeordnet.\n", title)
	}

	if !foundBooks {
		fmt.Println("Keine unsortierten Bücher gefunden.")
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	metadataDBPath := flag.String("metadata", "", "Pfad zur metadata.db (erforderlich)")
	appDBPath := flag.String("app", "", "Pfad zur app.db (erforderlich)")
	unsortedShelfID := flag.Int("shelf", 9, "ID des Regals für '_unsortiert' (Standard: 9)")
	showMissing := flag.Bool("missing", false, "Zeigt nur die fehlenden Buch-IDs an")
	showHelp := flag.Bool("help", false, "Zeigt diese Hilfe an")

	flag.Parse()

	if *showHelp || *metadataDBPath == "" || *appDBPath == "" {
		fmt.Println("CalibreUnsortedFixer - Organisiert unsortierte Bücher in Calibre")
		fmt.Println("\nVerwendung:")
		fmt.Println("  CalibreUnsortedFixer --metadata=PFAD --app=PFAD [--shelf=ID] [--missing]")
		fmt.Println("\nOptionen:")
		flag.PrintDefaults()
		fmt.Println("\nHinweis:")
		fmt.Println("  - Unter Windows müssen Pfade mit \\ maskiert werden, z.B.:")
		fmt.Println(`    --metadata="C:\\Pfad\\zu\\metadata.db"`)
		fmt.Println(`    --app="C:\\Pfad\\zu\\app.db"`)
		fmt.Println("\nBeispiele:")
		fmt.Println(`  CalibreUnsortedFixer --metadata="C:\\_Lokale_Daten_ungesichert\\source\\__TEST\\metadata.db" --app="C:\\_Lokale_Daten_ungesichert\\source\\__TEST\\app.db"`)
		fmt.Println(`  CalibreUnsortedFixer --metadata="C:\\metadata.db" --app="C:\\app.db" --missing`)
		os.Exit(0)
	}

	if *showMissing {
		showMissingBookIDs(*metadataDBPath, *appDBPath)
	} else {
		assignUnsortedBooks(*metadataDBPath, *appDBPath, *unsortedShelfID)
	}
}
