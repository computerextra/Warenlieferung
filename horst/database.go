package horst

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Warenlieferung struct {
	Id            int
	Name          string
	Angelegt      string
	Geliefert     string
	AlterPreis    float32
	NeuerPreis    float32
	Preis         string
	Artikelnummer string
}

type NeuerArtikel struct {
	Name          string
	Artikelnummer string
}

type NeuePreise struct {
	NeuerArtikel
	AlterPreis float32
	NeuerPreis float32
}

func getConnectionString() string {
	mysql_user := os.Getenv("MYSQL_USER")
	mysql_password := os.Getenv("MYSQL_PASS")
	mysql_server := os.Getenv("MYSQL_SERVER")
	mysql_db := os.Getenv("MYSQL_DB")

	mysql_port, err := strconv.ParseInt(os.Getenv("MYSQL_PORT"), 0, 64)
	if err != nil {
		log.Fatal("MYSQL_PORT not in .env: ", err)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysql_user, mysql_password, mysql_server, mysql_port, mysql_db)
}

func GetWarenlieferung() ([]Warenlieferung, error) {
	var warenlieferung []Warenlieferung

	connString := getConnectionString()

	conn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, fmt.Errorf("GetWarenlieferung: Open Connection failed: %s", err.Error())
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT * FROM Warenlieferung")
	if err != nil {
		return nil, fmt.Errorf("GetWarenlieferung: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var wl Warenlieferung
		var Name sql.NullString
		var Angelegt sql.NullString
		var Geliefert sql.NullString
		var AlterPreis sql.NullFloat64
		var NeuerPreis sql.NullFloat64
		var Preis sql.NullString
		var Artikelnummer sql.NullString

		if err := rows.Scan(&wl.Id, &Name, &Angelegt, &Geliefert, &AlterPreis, &NeuerPreis, &Preis, &Artikelnummer); err != nil {
			return nil, fmt.Errorf("GetWarenlieferung: Row Error: %s", err)
		}
		if Name.Valid {
			wl.Name = Name.String
		}
		if Angelegt.Valid {
			wl.Angelegt = Angelegt.String
		}
		if Geliefert.Valid {
			wl.Geliefert = Geliefert.String
		} else {
			wl.Geliefert = ""
		}
		if AlterPreis.Valid {
			wl.AlterPreis = float32(AlterPreis.Float64)
		} else {
			wl.AlterPreis = 0
		}
		if NeuerPreis.Valid {
			wl.NeuerPreis = float32(AlterPreis.Float64)
		} else {
			wl.NeuerPreis = 0
		}
		if Preis.Valid {
			wl.Preis = Preis.String
		} else {
			wl.Preis = ""
		}
		if Artikelnummer.Valid {
			wl.Artikelnummer = Artikelnummer.String
		}
		if Name.Valid && Angelegt.Valid && Artikelnummer.Valid {
			warenlieferung = append(warenlieferung, wl)
		}
	}

	return warenlieferung, nil
}

func UpdateWarenlieferung(neueArtikel []Warenlieferung, gelieferteArtikel []Warenlieferung, neuePreise []Warenlieferung) {
	connString := getConnectionString()

	conn, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal("UpdateWarenlieferung: Open Connection Failed: ", err)
	}
	defer conn.Close()

	// INSERT neueArtikel
	if len(neueArtikel) > 0 {
		for i := range neueArtikel {
			sql := fmt.Sprintf("INSERT INTO Warenlieferung (id, Name, angelegt, Artikelnummer) VALUES (%d, '%s', NOW(), '%s')", neueArtikel[i].Id, neueArtikel[i].Name, neueArtikel[i].Artikelnummer)
			_, err := conn.Exec(sql)
			if err != nil {
				log.Fatal("UpdateWarenlieferung: InsertNeu Query Failed: ", err)
			}
		}

	}

	// Update gelieferteArtikel
	for id := range gelieferteArtikel {
		sql := fmt.Sprintf("UPDATE Warenlieferung SET geliefert=NOW(), Name='%s' WHERE id=%d", strings.ReplaceAll(gelieferteArtikel[id].Name, "'", "\""), gelieferteArtikel[id].Id)
		_, err := conn.Exec(sql)
		if err != nil {
			log.Fatal("UpdateWarenlieferung: UpdateGeliefert Query Failed: ", err)
		}
	}

	// UPDATE neuePreise
	for i := range neuePreise {
		if neuePreise[i].NeuerPreis > 0 && neuePreise[i].AlterPreis > 0 {
			if neuePreise[i].NeuerPreis != neuePreise[i].AlterPreis {
				sql := fmt.Sprintf("UPDATE Warenlieferung SET `Preis`=NOW(), `AlterPreis`='%.2f', `NeuerPreis`='%.2f' WHERE `id`=%d", neuePreise[i].AlterPreis, neuePreise[i].NeuerPreis, neuePreise[i].Id)
				_, err := conn.Exec(sql)
				if err != nil {
					log.Println(sql)
					log.Fatal("UpdateWarenlieferung: UpdatePreise Query Failed: ", err)
				}
			}
		}
	}
}

func GetDailyWarenlieferung() ([]NeuerArtikel, []NeuerArtikel, []NeuePreise) {
	NeueArtikel, err := getDailyNew()
	if err != nil {
		log.Fatal(err)
	}
	GelieferteArtikel, err := getDailyDelivered()
	if err != nil {
		log.Fatal(err)
	}
	NeuePreise, err := getDailyPrices()
	if err != nil {
		log.Fatal(err)
	}

	return NeueArtikel, GelieferteArtikel, NeuePreise
}

func getDailyPrices() ([]NeuePreise, error) {
	var Warenlieferung []NeuePreise

	connString := getConnectionString()

	conn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, fmt.Errorf("getDailyDelivered: Open Connection failed: %s", err)
	}
	defer conn.Close()

	sqlQuery := fmt.Sprintf("SELECT Name, Artikelnummer, AlterPreis, NeuerPreis FROM Warenlieferung WHERE DATE_FORMAT(Preis, '%%Y-%%m-%%d') = DATE_FORMAT(NOW(), '%%Y-%%m-%%d') AND DATE_FORMAT(angelegt, '%%Y-%%m-%%d') != DATE_FORMAT(NOW(), '%%Y-%%m-%%d') ORDER BY Artikelnummer ASC")

	rows, err := conn.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetDailyNew: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var Ware NeuePreise
		var Name sql.NullString
		var Artikelnummer sql.NullString
		var AlterPreis sql.NullFloat64
		var NeuerPreis sql.NullFloat64

		if err := rows.Scan(&Name, &Artikelnummer, &AlterPreis, &NeuerPreis); err != nil {
			return nil, fmt.Errorf("getDailyNew: Row Error: %s", err)
		}
		if Name.Valid {
			Ware.Name = Name.String
		}
		if Artikelnummer.Valid {
			Ware.Artikelnummer = Artikelnummer.String
		}
		if NeuerPreis.Valid {
			Ware.NeuerPreis = float32(NeuerPreis.Float64)
		}
		if AlterPreis.Valid {
			Ware.AlterPreis = float32(AlterPreis.Float64)
		}
		if Artikelnummer.Valid && Name.Valid && NeuerPreis.Valid && AlterPreis.Valid {
			Warenlieferung = append(Warenlieferung, Ware)
		}
	}
	return Warenlieferung, nil
}

func getDailyDelivered() ([]NeuerArtikel, error) {
	var Warenlieferung []NeuerArtikel

	connString := getConnectionString()

	conn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, fmt.Errorf("getDailyDelivered: Open Connection failed: %s", err)
	}
	defer conn.Close()

	sqlQuery := fmt.Sprintf("SELECT Name, Artikelnummer FROM Warenlieferung WHERE DATE_FORMAT(geliefert, '%%Y-%%m-%%d') = DATE_FORMAT(NOW(), '%%Y-%%m-%%d') AND DATE_FORMAT(angelegt, '%%Y-%%m-%%d') != DATE_FORMAT(NOW(), '%%Y-%%m-%%d') ORDER BY Artikelnummer ASC")

	rows, err := conn.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetDailyNew: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var Ware NeuerArtikel
		var Name sql.NullString
		var Artikelnummer sql.NullString
		if err := rows.Scan(&Name, &Artikelnummer); err != nil {
			return nil, fmt.Errorf("getDailyNew: Row Error: %s", err)
		}
		if Name.Valid {
			Ware.Name = Name.String
		}
		if Artikelnummer.Valid {
			Ware.Artikelnummer = Artikelnummer.String
		}
		if Artikelnummer.Valid && Name.Valid {
			Warenlieferung = append(Warenlieferung, Ware)
		}
	}
	return Warenlieferung, nil
}

func getDailyNew() ([]NeuerArtikel, error) {
	var Warenlieferung []NeuerArtikel

	connString := getConnectionString()

	conn, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, fmt.Errorf("GetDailyNew: Open Connection failed: %s", err)
	}
	defer conn.Close()

	sqlQuery := fmt.Sprintf("SELECT Name, Artikelnummer FROM Warenlieferung WHERE DATE_FORMAT(angelegt, '%%Y-%%m-%%d') = DATE_FORMAT(NOW(), '%%Y-%%m-%%d') ORDER BY Artikelnummer ASC")

	rows, err := conn.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("GetDailyNew: Query failed: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var Ware NeuerArtikel
		var Name sql.NullString
		var Artikelnummer sql.NullString
		if err := rows.Scan(&Name, &Artikelnummer); err != nil {
			return nil, fmt.Errorf("getDailyNew: Row Error: %s", err)
		}
		if Name.Valid {
			Ware.Name = Name.String
		}
		if Artikelnummer.Valid {
			Ware.Artikelnummer = Artikelnummer.String
		}
		if Artikelnummer.Valid && Name.Valid {
			Warenlieferung = append(Warenlieferung, Ware)
		}
	}
	return Warenlieferung, nil
}
