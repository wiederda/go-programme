package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/* ============================================================
   Hilfsfunktionen
   ============================================================ */

// Hilfetext anzeigen
func printHelp() {
	fmt.Println("datumscheck – Datum, Wochentag & Feiertage (Schleswig-Holstein)")
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("Verwendung:")
	fmt.Println("  datumscheck <Datum im Format TT.MM.JJJJ>")
	fmt.Println("  datumscheck <Feiertag> <Jahr>")
	fmt.Println()
	fmt.Println("Beispiele:")
	fmt.Println("  datumscheck 24.12.2028")
	fmt.Println("Ergebnis: Der 24.12.2028 fällt auf einen Sonntag")
	fmt.Println("  datumscheck Ostern 2027")
	fmt.Println("  datumscheck Reformationstag 2030")
}

// Ostersonntag nach Meeus/Jones/Butcher berechnen
func getEasterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

// Feiertag → Datum
func getHolidayDate(holiday string, year int) (time.Time, error) {
	holiday = strings.ToLower(holiday)
	easter := getEasterSunday(year)

	switch holiday {

	// Feste Feiertage Schleswig-Holstein
	case "neujahr":
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local), nil
	case "tag der arbeit":
		return time.Date(year, 5, 1, 0, 0, 0, 0, time.Local), nil
	case "reformationstag", "reformation":
		return time.Date(year, 10, 31, 0, 0, 0, 0, time.Local), nil
	case "tag der deutschen einheit":
		return time.Date(year, 10, 3, 0, 0, 0, 0, time.Local), nil
	case "1. weihnachtstag", "weihnachten 1":
		return time.Date(year, 12, 25, 0, 0, 0, 0, time.Local), nil
	case "2. weihnachtstag", "weihnachten 2":
		return time.Date(year, 12, 26, 0, 0, 0, 0, time.Local), nil

	// Bewegliche Feiertage
	case "karfreitag":
		return easter.AddDate(0, 0, -2), nil
	case "ostern", "ostersonntag":
		return easter, nil
	case "ostermontag":
		return easter.AddDate(0, 0, 1), nil
	case "christi himmelfahrt":
		return easter.AddDate(0, 0, 39), nil
	case "pfingsten", "pfingstsonntag":
		return easter.AddDate(0, 0, 49), nil
	case "pfingstmontag":
		return easter.AddDate(0, 0, 50), nil
	}

	return time.Time{}, fmt.Errorf("Feiertag '%s' unbekannt", holiday)
}

// Prüfen, ob ein Datum ein Feiertag ist
func checkHoliday(t time.Time) string {
	year := t.Year()
	feiertage := map[string]time.Time{
		"Neujahr":                   time.Date(year, 1, 1, 0, 0, 0, 0, time.Local),
		"Tag der Arbeit":            time.Date(year, 5, 1, 0, 0, 0, 0, time.Local),
		"Reformationstag":           time.Date(year, 10, 31, 0, 0, 0, 0, time.Local),
		"Tag der Deutschen Einheit": time.Date(year, 10, 3, 0, 0, 0, 0, time.Local),
		"1. Weihnachtsfeiertag":     time.Date(year, 12, 25, 0, 0, 0, 0, time.Local),
		"2. Weihnachtsfeiertag":     time.Date(year, 12, 26, 0, 0, 0, 0, time.Local),

		"Karfreitag":          getEasterSunday(year).AddDate(0, 0, -2),
		"Ostersonntag":        getEasterSunday(year),
		"Ostermontag":         getEasterSunday(year).AddDate(0, 0, 1),
		"Christi Himmelfahrt": getEasterSunday(year).AddDate(0, 0, 39),
		"Pfingstsonntag":      getEasterSunday(year).AddDate(0, 0, 49),
		"Pfingstmontag":       getEasterSunday(year).AddDate(0, 0, 50),
	}

	for name, d := range feiertage {
		if d.Format("02.01.2006") == t.Format("02.01.2006") {
			return name
		}
	}

	return ""
}

/* ============================================================
   Hauptlogik
   ============================================================ */

func main() {
	if len(os.Args) == 1 || os.Args[1] == "-h" {
		printHelp()
		return
	}

	// Fall 1: <Datum>
	if len(os.Args) == 2 && strings.Contains(os.Args[1], ".") {
		dateString := os.Args[1]
		const layout = "02.01.2006"

		t, err := time.Parse(layout, dateString)
		if err != nil {
			fmt.Println("Ungültiges Datum. Format muss TT.MM.JJJJ sein.")
			return
		}

		weekday := t.Weekday()
		germanDays := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
		dayName := germanDays[weekday]

		// Grammatik: Vergangenheit vs. Zukunft
		now := time.Now()
		var text string
		if t.Before(now) {
			text = "fiel auf einen"
		} else if t.After(now) {
			text = "fällt auf einen"
		} else {
			text = "ist ein"
		}

		fmt.Printf("Der %s %s %s.\n", dateString, text, dayName)

		if feiertag := checkHoliday(t); feiertag != "" {
			fmt.Printf("→ Das ist der Feiertag: %s\n", feiertag)
		}

		return
	}

	// Fall 2: Feiertag + Jahr
	if len(os.Args) == 3 {
		holiday := os.Args[1]
		year, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Ungültiges Jahr.")
			return
		}

		date, err := getHolidayDate(holiday, year)
		if err != nil {
			fmt.Println(err)
			return
		}

		germanDays := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
		dayName := germanDays[date.Weekday()]

		fmt.Printf("%s %d ist am %s (%s).\n",
			strings.Title(holiday),
			year,
			date.Format("02.01.2006"),
			dayName)

		return
	}

	printHelp()
}
