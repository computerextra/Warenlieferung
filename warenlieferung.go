package main

import (
	"fmt"
	"log"
	"warenlieferung/helper"
	"warenlieferung/horst"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	neueArtikel, geliefert, neuePreise := helper.SortProducts()

	fmt.Println("Neue Artikel noch nicht in Datenbank: ", len(neueArtikel))
	fmt.Println("Gelieferte Artikel laut Sage", len(geliefert))
	fmt.Println("Neue Preise laut Sage", len(neuePreise))

	horst.UpdateWarenlieferung(neueArtikel, geliefert, neuePreise)

	fmt.Println("Artikeldatenbank mit Sage Synchronisiert.")

	fmt.Println("Erstelle Mail")
	helper.SendMail(horst.GetDailyWarenlieferung())
	fmt.Println("Mail versendet")
	fmt.Printf("bye\n")
}
