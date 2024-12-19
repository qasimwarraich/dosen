package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"unicode"

	"github.com/gocolly/colly/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/muesli/reflow/wordwrap"
)

var URL string = "https://www.aktionis.ch/deals?c=8-26"

type Deal struct {
	name        string
	description string
	price       string
	discount    string
	store       string
	validity    string
}

func main() {
	ankerOnSale := false
	var ankerDeal Deal

	deals, err := scrapeDeals()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Price (CHF)", "Discount", "Store", "Validity", "Misc"})

	for _, deal := range deals {
		if strings.Contains(deal.name, "Anker") || strings.Contains(deal.name, "anker") {
			ankerOnSale = true
			ankerDeal = Deal{
				name:        deal.name,
				description: deal.description,
				price:       deal.price,
				discount:    deal.description,
				store:       deal.store,
				validity:    deal.validity,
			}
		}

		if strings.Contains(deal.store, "OTTO'S") || deal.description == "" {
			deal.description = tryFormatFromName(deal.name)
		}

		t.AppendRow(table.Row{deal.name, deal.price, deal.discount, deal.store, deal.validity, wordwrap.String(deal.description, 40)})
		t.AppendSeparator()
	}

	if !ankerOnSale {
		fmt.Println("Anker is not on sale :/, here is the best of the rest:")
	} else {
		fmt.Println("🚨🚨🚨🚨🚨🚨")
		fmt.Println("ANKER IS ON SALE!")
		fmt.Println(ankerDeal)
		fmt.Printf(
			"Anker is on sale at %s!. The price is %s, which means a discount of %s!.\ninfo: %s\n until: %s\n",
			ankerDeal.store,
			ankerDeal.price,
			ankerDeal.discount,
			ankerDeal.description,
			ankerDeal.validity,
		)
		fmt.Println("🚨🚨🚨🚨🚨🚨")
		fmt.Println("Here is the best of the rest anyway:")
	}

	t.SetStyle(table.StyleBold)
	t.Render()
	os.Exit(0)
}

func scrapeDeals() ([]Deal, error) {
	deals := []Deal{}
	c := colly.NewCollector(colly.AllowedDomains("www.aktionis.ch"))
	c.OnHTML(".card", func(e *colly.HTMLElement) {
		name := e.ChildText(".card-title")
		newPriceText := strings.Split(e.ChildText(".price-new"), " ")
		if len(newPriceText) == 1 {
			newPrice := newPriceText[0]
			discount := e.ChildText(".price-discount")
			description := e.ChildText(".card-description")
			validity := e.ChildText(".card-date")
			store := e.DOM.Find("img").Nodes[0].Attr[1].Val

			deals = append(deals, Deal{
				name:        name,
				description: description,
				price:       newPrice,
				discount:    discount,
				store:       store,
				validity:    validity,
			})
		}
	})

	err := c.Visit(URL)
	if err != nil {
		slog.Error(err.Error())
		return []Deal{}, fmt.Errorf("error visiting url %s: %w", URL, err)
	}

	return deals, nil
}

func tryFormatFromName(name string) string {
	trimmed := strings.TrimLeftFunc(name, func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	return trimmed
}
