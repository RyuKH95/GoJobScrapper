package main

import (
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/rkh/JobScrapper/scrapper"
)

const fileNAME string = "jobs.csv"

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileNAME)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileNAME, fileNAME)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrapper", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
