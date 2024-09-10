package sage

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"

	_ "github.com/denisenkom/go-mssqldb"
)

type Artikel struct {
	Id            int
	Artikelnummer string
	Suchbegriff   string
}

type Leichen struct {
	Artikelnummer string
	Artikelname   string
	Bestand       int16
	Verfügbar     int16
	EK            float64
	LetzterUmsatz string
}

type SummenArtikel struct {
	Artikelnummer string
	Artikelname   string
	Bestand       int16
	EK            float64
	Summe         float64
}

type VerfArtikel struct {
	SummenArtikel
	Verfügbar int16
}

type WertArtikel struct {
	Bestand   int16
	Verfügbar int16
	EK        float64
}

type History struct {
	Id     int
	Action string
}

type Price struct {
	Id     int
	Action string
	Price  float32
}

type AlteSeriennummer struct {
	ArtNr       string
	Suchbegriff string
	Bestand     int
	Verfügbar   int
	GeBeginn    string
}

func getConnectionString() string {

	server := os.Getenv("SAGE_SERVER")
	db := os.Getenv("SAGE_DB")
	user := os.Getenv("SAGE_USER")
	password := os.Getenv("SAGE_PASS")
	port, err := strconv.ParseInt(os.Getenv("SAGE_PORT"), 0, 64)
	if err != nil {
		log.Fatal("SAGE_PORT not in .env: ", err)
	}
	return fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", server, db, user, password, port)
}

func GetLeichen() ([]Leichen, error) {
	var artikel []Leichen

	conn, err := sql.Open("sqlserver", getConnectionString())
	if err != nil {
		return nil, fmt.Errorf("GetLeichen: Open Connection failed: %s", err.Error())
	}

	defer conn.Close()
	rows, err := conn.Query("SELECT TOP 20 ARTNR, SUCHBEGRIFF, BESTAND, VERFUEGBAR, LetzterUmsatz, EKPR01 FROM sg_auf_artikel WHERE VERFUEGBAR > 0 ORDER BY LetzterUmsatz ASC")
	if err != nil {
		return nil, fmt.Errorf("GetLeichen: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var art Leichen
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var EK sql.NullFloat64
		var LetzerUmsatz sql.NullString

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &Verfügbar, &LetzerUmsatz, &EK); err != nil {
			return nil,
				fmt.Errorf("GetLeichen: Scan failed: %s", err.Error())
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Verfügbar.Valid {
			art.Verfügbar = Verfügbar.Int16
		}
		if LetzerUmsatz.Valid {
			art.LetzterUmsatz = LetzerUmsatz.String
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && LetzerUmsatz.Valid && Verfügbar.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetLeichen: Row Error: %s", err)
	}

	return artikel, nil
}

func GetHighestVerfSum() ([]VerfArtikel, error) {
	var artikel []VerfArtikel

	conn, err := sql.Open("sqlserver", getConnectionString())
	if err != nil {
		return nil, fmt.Errorf("GetHighestVerfSum: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT TOP 10 ARTNR, SUCHBEGRIFF, BESTAND, VERFUEGBAR, EKPR01, VERFUEGBAR * EKPR01 as Summe FROM sg_auf_artikel WHERE VERFUEGBAR > 0 ORDER BY Summe DESC")
	if err != nil {
		return nil, fmt.Errorf("GetHighestVerfSum: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var art VerfArtikel
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var EK sql.NullFloat64
		var Summe sql.NullFloat64

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &Verfügbar, &EK, &Summe); err != nil {
			return nil,
				fmt.Errorf("GetHighestVerfSum: Scan failed: %s", err.Error())
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Summe.Valid {
			art.Summe = Summe.Float64
		}
		if Verfügbar.Valid {
			art.Verfügbar = Verfügbar.Int16
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && Summe.Valid && Verfügbar.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetHighestVerfSum: Row Error: %s", err)
	}

	return artikel, nil
}

func GetHighestSum() ([]SummenArtikel, error) {
	var artikel []SummenArtikel

	conn, err := sql.Open("sqlserver", getConnectionString())
	if err != nil {
		return nil, fmt.Errorf("GetHighestSum: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT TOP 10 ARTNR, SUCHBEGRIFF, BESTAND, EKPR01, BESTAND * EKPR01 as Summe FROM sg_auf_artikel WHERE BESTAND > 0 ORDER BY Summe DESC")
	if err != nil {
		return nil, fmt.Errorf("GetHighestSum: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var art SummenArtikel
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var EK sql.NullFloat64
		var Summe sql.NullFloat64

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &EK, &Summe); err != nil {
			return nil,
				fmt.Errorf("GetHighestSum: Scan failed: %s", err.Error())
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Summe.Valid {
			art.Summe = Summe.Float64
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && Summe.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetHighestSum: Row Error: %s", err)
	}

	return artikel, nil
}

func GetLagerWert() (float64, float64, error) {
	var wertBestand float64
	var wertVerfügbar float64
	wertBestand = 0
	wertVerfügbar = 0

	conn, err := sql.Open("sqlserver", getConnectionString())
	if err != nil {
		return 0, 0, fmt.Errorf("GetLagerWert: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT BESTAND, VERFUEGBAR, EKPR01 FROM sg_auf_artikel WHERE BESTAND > 0")
	if err != nil {
		return 0, 0, fmt.Errorf("GetLagerWert: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var Ek sql.NullFloat64

		if err := rows.Scan(&Bestand, &Verfügbar, &Ek); err != nil {
			return 0, 0,
				fmt.Errorf("GetLagerWert: Scan failed: %s", err.Error())
		}
		if Bestand.Valid && Ek.Valid {
			wertBestand = wertBestand + (float64(Bestand.Int16) * Ek.Float64)
		}
		if Verfügbar.Valid && Ek.Valid {
			wertVerfügbar = wertVerfügbar + (float64(Verfügbar.Int16) * Ek.Float64)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("GetLagerWert: Row Error: %s", err)
	}

	return wertBestand, wertVerfügbar, nil
}

func GetAllProductsFromSage() ([]Artikel, error) {
	flag.Parse()

	var artikel []Artikel

	connString := getConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("GetAllProductsFromSage: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT SG_AUF_ARTIKEL_PK, ARTNR, SUCHBEGRIFF FROM sg_auf_artikel")
	if err != nil {
		return nil, fmt.Errorf("GetAllProductsFromSage: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var art Artikel
		var Artikelnummer sql.NullString
		var Suchbegriff sql.NullString

		if err := rows.Scan(&art.Id, &Artikelnummer, &Suchbegriff); err != nil {
			return nil,
				fmt.Errorf("GetAllProductsFromSage: Scan failed: %s", err.Error())
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Suchbegriff.Valid {
			art.Suchbegriff = Suchbegriff.String
		}
		if Suchbegriff.Valid && Artikelnummer.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllProductsFromSage: Row Error: %s", err)
	}

	return artikel, nil
}

func GetLagerHistory() ([]History, error) {
	var history []History

	queryString := fmt.Sprintf("SELECT SG_AUF_ARTIKEL_FK, Hist_Action FROM sg_auf_lager_history WHERE BEWEGUNG >= 0 AND BEMERKUNG LIKE 'Warenlieferung:%%' AND convert(varchar, Hist_Datetime, 105) = convert(varchar, getdate(), 105)")

	connString := getConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("GetLagerHistory: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()

	rows, err := conn.Query(queryString)
	if err != nil {
		return nil, fmt.Errorf("GetLagerHistory: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var hist History
		var Action sql.NullString

		if err := rows.Scan(&hist.Id, &Action); err != nil {
			return nil, fmt.Errorf("GetLagerHistory: Row Error: %s", err)
		}
		if Action.Valid {
			hist.Action = Action.String
			history = append(history, hist)
		}
	}

	return history, nil
}

func GetPrices() ([]Price, error) {
	var prices []Price

	queryString := "SELECT Hist_Action, SG_AUF_ARTIKEL_FK, PR01 FROM sg_auf_vkpreis_history WHERE convert(varchar, Hist_Datetime, 105) = convert(varchar, getdate(), 105)"

	connString := getConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("GetPrices: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()

	rows, err := conn.Query(queryString)
	if err != nil {
		return nil, fmt.Errorf("GetPrices: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var price Price
		var Action sql.NullString
		var p sql.NullFloat64

		if err := rows.Scan(&Action, &price.Id, &p); err != nil {
			return nil, fmt.Errorf("GetPrices: Row Error: %s", err)
		}
		if p.Valid {
			price.Price = float32(p.Float64)
		}
		if Action.Valid {
			price.Action = Action.String
		}
		if p.Valid && Action.Valid {
			prices = append(prices, price)
		}
	}

	return prices, nil
}

func GetAlteSeriennummern() ([]AlteSeriennummer, error) {
	var artikel []AlteSeriennummer

	conn, err := sql.Open("sqlserver", getConnectionString())
	if err != nil {
		return nil, fmt.Errorf("GetAlteSeriennummern: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT sg_auf_artikel.ARTNR, sg_auf_artikel.SUCHBEGRIFF, sg_auf_artikel.BESTAND, sg_auf_artikel.VERFUEGBAR, sg_auf_snr.GE_Beginn FROM sg_auf_artikel INNER JOIN sg_auf_snr ON sg_auf_artikel.SG_AUF_ARTIKEL_PK = sg_auf_snr.SG_AUF_ARTIKEL_FK  WHERE sg_auf_artikel.VERFUEGBAR > 0 AND sg_auf_snr.SNR_STATUS != 2 AND sg_auf_snr.GE_Beginn <= DATEADD(month, DATEDIFF(month, 0, DATEADD(MONTH,-1,GETDATE())), 0) ORDER BY sg_auf_snr.GE_Beginn ")
	if err != nil {
		return nil, fmt.Errorf("GetAlteSeriennummern: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var art AlteSeriennummer
		var Artikelnummer sql.NullString
		var Suchbegriff sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var Garantie sql.NullString

		if err := rows.Scan(&Artikelnummer, &Suchbegriff, &Bestand, &Verfügbar, &Garantie); err != nil {
			return nil,
				fmt.Errorf("GetAlteSeriennummern: Scan failed: %s", err.Error())
		}

		if Artikelnummer.Valid && Suchbegriff.Valid && Bestand.Valid && Verfügbar.Valid && Garantie.Valid {
			art.ArtNr = Artikelnummer.String
			art.Suchbegriff = Suchbegriff.String
			art.Bestand = int(Bestand.Int16)
			art.Verfügbar = int(Verfügbar.Int16)
			art.GeBeginn = Garantie.String
			if !slices.Contains(artikel, art) {
				artikel = append(artikel, art)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAlteSeriennummern: Row Error: %s", err)
	}

	return artikel, nil
}
