package helper

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"warenlieferung/horst"
	"warenlieferung/sage"

	"github.com/go-mail/mail"
)

func SendMail(neueArtikel []horst.NeuerArtikel, gelieferteArtikel []horst.NeuerArtikel, neuePreise []horst.NeuePreise) {

	// Zieh die aktuellen Bestände
	wertBestand, wertVerfügbar, err := sage.GetLagerWert()
	if err != nil {
		log.Fatal("SendMail: Fehler beim ermitteln der LagerWerte!", err)
	}

	// Ziehe die Top 10 der teuersten Artikel auf Lager
	teureArtikel, err := sage.GetHighestSum()
	if err != nil {
		log.Fatal("SendMail: Fehler beim ermitteln der Teuersten Artikel!", err)
	}

	teureVerfArtikel, err := sage.GetHighestVerfSum()
	if err != nil {
		log.Fatal("SendMail: Fehler beim ermitteln der Teuersten Artikel!", err)
	}

	leichen, err := sage.GetLeichen()
	if err != nil {
		log.Fatal("SendMail: Fehler beim ermitteln der Leichen", err)
	}

	var body string

	if len(neueArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Neue Artikel</h2><ul>", body)

		for i := range neueArtikel {
			body = fmt.Sprintf("%s<li><b>%s</b> - %s</li>", body, neueArtikel[i].Artikelnummer, neueArtikel[i].Name)
		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	if len(gelieferteArtikel) > 0 {
		body = fmt.Sprintf("%s<br><br><h2>Gelieferte Artikel</h2><ul>", body)

		for i := range gelieferteArtikel {
			body = fmt.Sprintf("%s<li><b>%s</b> - %s</li>", body, gelieferteArtikel[i].Artikelnummer, gelieferteArtikel[i].Name)
		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	if len(neuePreise) > 0 {
		body = fmt.Sprintf("%s<br><br><h2>Preisänderungen</h2><ul>", body)

		for i := range neuePreise {
			var alterPreis string
			var neuerPreis string

			alterPreis = fmt.Sprintf("%.2f", neuePreise[i].AlterPreis)
			neuerPreis = fmt.Sprintf("%.2f", neuePreise[i].NeuerPreis)

			if neuerPreis != alterPreis {
				body = fmt.Sprintf("%s<li><b>%s</b> - %s: %s ➡️ %s ", body, neuePreise[i].Artikelnummer, neuePreise[i].Name, alterPreis, neuerPreis)

				absolute := neuePreise[i].NeuerPreis - neuePreise[i].AlterPreis
				prozent := ((neuePreise[i].NeuerPreis / neuePreise[i].AlterPreis) * 100) - 100
				body = fmt.Sprintf("%s(%.2f %% // %.2f €)</li>", body, prozent, absolute)
			}

		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	body = fmt.Sprintf("%s<h2>Aktuelle Lagerwerte</h2><p><b>Lagerwert Verfügbare Artikel:</b> %.2f €</p><p><b>Lagerwert alle lagernde Artikel:</b> %.2f €</p>", body, wertVerfügbar, wertBestand)
	body = fmt.Sprintf("%s<p>Wert in aktuellen Aufträgen: %.2f €", body, wertBestand-wertVerfügbar)

	SN, err := sage.GetAlteSeriennummern()
	if err != nil {
		log.Fatal("SendMail: Fehler beim ermitteln von GetAlteSeriennummern!", err)
	}

	if len(SN) > 0 {
		body = fmt.Sprintf("%s<h2>Artikel mit alten Seriennummern</h2><p>Nachfolgende Artikel sollten mit erhöhter Prioriät verkauf werden, da die Seriennummern bereits sehr alt sind. Gegebenenfalls sind die Artikel bereits außerhalb der Herstellergarantie!</p>", body)
		body = fmt.Sprintf("%s<p>Folgende Werte gelten:</p>", body)
		body = fmt.Sprintf("%s<p>Wortmann: Angebene Garantielaufzeit + 2 Monate ab Kaufdatum CompEx</p>", body)
		body = fmt.Sprintf("%s<p>Lenovo: Angegebene Garantielaufzeit ab Kauf CompEx</p>", body)
		body = fmt.Sprintf("%s<p>Bei allen anderen Herstellern gilt teilweise das Kaufdatum des Kunden. <br>Falls sich dies ändern sollte, wird es in der Aufzählung ergänzt.</p>", body)

		body = fmt.Sprintf("%s<p>Erklärungen der Farben:</p>", body)
		body = fmt.Sprintf("%s<p><span style='background-color: \"#f45865\"'>ROT:</span> Artikel ist bereits seit mehr als 2 Jahren lagernd und sollte schnellstens Verkauft werden!</p>", body)
		body = fmt.Sprintf("%s<p><span style='background-color: \"#a3a53a\"'>GELB:</span> Artikel ist bereits seit mehr als 1 Jahr lagernd!</p>", body)

		body = fmt.Sprintf("%s<table><thead>", body)
		body = fmt.Sprintf("%s<tr>", body)
		body = fmt.Sprintf("%s<th>Artikelnummer</th>", body)
		body = fmt.Sprintf("%s<th>Name</th>", body)
		body = fmt.Sprintf("%s<th>Bestand</th>", body)
		body = fmt.Sprintf("%s<th>Verfügbar</th>", body)
		body = fmt.Sprintf("%s<th>Garantiebeginn des ältesten Artikels</th>", body)
		body = fmt.Sprintf("%s</tr>", body)
		body = fmt.Sprintf("%s</thead>", body)
		body = fmt.Sprintf("%s</thbody>", body)
		for i := range SN {
			year, _, _ := time.Now().Date()
			tmp := strings.Split(strings.Replace(strings.Split(SN[i].GeBeginn, "T")[0], "-", ".", -1), ".")
			year_tmp, err := strconv.Atoi(tmp[0])
			if err != nil {
				log.Fatal("SendMail: Fehler beim voncertieren von string zu int (year) in GetAlteSeriennummern!", err)
			}

			GarantieBeginn := fmt.Sprintf("%s.%s.%s", tmp[2], tmp[1], tmp[0])
			diff := year - year_tmp
			if diff >= 2 {
				body = fmt.Sprintf("%s<tr style='background-color: \"#f45865\"'>", body)
			} else if diff >= 1 {
				body = fmt.Sprintf("%s<tr style='background-color: \"#a3a53a\"'>", body)
			} else {
				body = fmt.Sprintf("%s<tr>", body)
			}
			body = fmt.Sprintf("%s<td>%s</td>", body, SN[i].ArtNr)
			body = fmt.Sprintf("%s<td>%s</td>", body, SN[i].Suchbegriff)
			body = fmt.Sprintf("%s<td>%v</td>", body, SN[i].Bestand)
			body = fmt.Sprintf("%s<td>%v</td>", body, SN[i].Verfügbar)
			body = fmt.Sprintf("%s<td>%s</td>", body, GarantieBeginn)
			body = fmt.Sprintf("%s</tr>", body)

		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(teureArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Top 10: Die teuersten Artikel inkl. aktive Aufträge</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Einzelpreis</th><th>Summe</th></tr></thead><tbody>", body)

		for i := range teureArtikel {
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%.2f €</td><td>%.2f €</td></tr>", body, teureArtikel[i].Artikelnummer, teureArtikel[i].Artikelname, teureArtikel[i].Bestand, teureArtikel[i].EK, teureArtikel[i].Summe)
		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(teureVerfArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Top 10: Die teuersten Artikel exkl. aktive Aufträge</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Einzelpreis</th><th>Summe</th></tr></thead><tbody>", body)

		for i := range teureVerfArtikel {
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%.2f €</td><td>%.2f €</td></tr>", body, teureVerfArtikel[i].Artikelnummer, teureVerfArtikel[i].Artikelname, teureVerfArtikel[i].Bestand, teureVerfArtikel[i].EK, teureVerfArtikel[i].Summe)

		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(leichen) > 0 {
		body = fmt.Sprintf("%s<h2>Top 20: Leichen bei CE</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Verfügbar</th><th>Letzter Umsatz:</th><th>Wert im Lager:</th></tr></thead><tbody>", body)
		for i := range leichen {
			summe := float64(leichen[i].Verfügbar) * leichen[i].EK
			var LetzterUmsatz string
			if leichen[i].LetzterUmsatz == "1899-12-30T00:00:00Z" {
				LetzterUmsatz = "nie"
			} else {
				tmp := strings.Split(strings.Replace(strings.Split(leichen[i].LetzterUmsatz, "T")[0], "-", ".", -1), ".")
				LetzterUmsatz = fmt.Sprintf("%s.%s.%s", tmp[2], tmp[1], tmp[0])
			}
			bestand := leichen[i].Bestand
			verf := leichen[i].Verfügbar
			artNr := leichen[i].Artikelnummer
			name := leichen[i].Artikelname
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%.2f€</td></tr>", body, artNr, name, bestand, verf, LetzterUmsatz, summe)
		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	m := mail.NewMessage()

	smtp_from := os.Getenv("SMTP_FROM")
	smtp_to := os.Getenv("SMTP_TO")
	smtp_server := os.Getenv("SMTP_SERVER")
	smtp_user := os.Getenv("SMTP_USER")
	smtp_pass := os.Getenv("SMTP_PASS")

	smtp_port, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 0, 64)
	if err != nil {
		log.Fatal("SMTP_PORT not in .env: ", err)
	}

	m.SetHeader("From", smtp_from)
	m.SetHeader("To", smtp_to)
	m.SetHeader("Subject", fmt.Sprintf("Warenlieferung vom %d.%d.%d", time.Now().Day(), time.Now().Month(), time.Now().Year()))
	m.SetBody("text/html", body)

	d := mail.NewDialer(smtp_server, int(smtp_port), smtp_user, smtp_pass)

	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Mail konnte nciht gesendet werden! WEIL DU FETT BIST!!!! : %s", err)
	}

}
