package helper

import (
	"log"
	"time"
	"warenlieferung/horst"
	"warenlieferung/sage"
)

func SortProducts() ([]horst.Warenlieferung, []horst.Warenlieferung, []horst.Warenlieferung) {
	SageProducts, err := sage.GetAllProductsFromSage()
	if err != nil {
		log.Fatal(err)
	}

	TodaysHistory, err := sage.GetLagerHistory()
	if err != nil {
		log.Fatal(err)
	}

	DatabaseProducts, err := horst.GetWarenlieferung()
	if err != nil {
		log.Fatal(err)
	}

	Prices, err := sage.GetPrices()
	if err != nil {
		log.Fatal(err)
	}

	var neueArtikel []horst.Warenlieferung
	var gelieferteArtikel []horst.Warenlieferung
	var geliefert []int
	var neuePreise []horst.Warenlieferung

	if len(DatabaseProducts) <= 0 {
		for i := range SageProducts {
			var neu horst.Warenlieferung
			neu.Id = SageProducts[i].Id
			neu.Artikelnummer = SageProducts[i].Artikelnummer
			neu.Name = SageProducts[i].Suchbegriff
			neueArtikel = append(neueArtikel, neu)
		}
	} else {
		for i := range TodaysHistory {
			if TodaysHistory[i].Action == "Insert" {

				geliefert = append(geliefert, TodaysHistory[i].Id)
			}
		}
		for i := range SageProducts {
			var found bool
			found = false
			// Check if Product is in Database
			for y := 0; y < len(geliefert); y++ {
				if SageProducts[i].Id == geliefert[y] {
					var prod horst.Warenlieferung
					prod.Id = SageProducts[i].Id
					prod.Name = SageProducts[i].Suchbegriff
					gelieferteArtikel = append(gelieferteArtikel, prod)
				}
			}
			for x := 0; x < len(DatabaseProducts); x++ {
				if SageProducts[i].Id == DatabaseProducts[x].Id {
					found = true
					break
				}
			}
			// Check if id is in geliefert
			if !found {
				var neu horst.Warenlieferung
				neu.Id = SageProducts[i].Id
				neu.Artikelnummer = SageProducts[i].Artikelnummer
				neu.Name = SageProducts[i].Suchbegriff
				neueArtikel = append(neueArtikel, neu)
			}
		}
		// Check if Prices are different
		for i := range Prices {
			var tmp horst.Warenlieferung
			var found bool
			idx := 0
			if len(neuePreise) > 0 {
				// Find current item
				for x := 0; x < len(neuePreise); x++ {
					if neuePreise[x].Id == Prices[i].Id {
						found = true
						tmp = neuePreise[x]
						idx = x
					}
				}
			}
			if !found {
				tmp.Id = Prices[i].Id
				tmp.Preis = time.Now().Format(time.DateOnly)
			}
			if Prices[i].Action == "Insert" {
				tmp.NeuerPreis = Prices[i].Price
			}
			if Prices[i].Action == "Delete" {
				tmp.AlterPreis = Prices[i].Price
			}

			if idx > 0 {
				if tmp.AlterPreis > 0 {
					neuePreise[idx].AlterPreis = tmp.AlterPreis
				}
				if tmp.NeuerPreis > 0 {
					neuePreise[idx].NeuerPreis = tmp.NeuerPreis
				}
			} else {
				neuePreise = append(neuePreise, tmp)
			}
		}
	}

	return neueArtikel, gelieferteArtikel, neuePreise
}
