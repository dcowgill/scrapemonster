package main

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/launchtime/scrapemonster/scrape"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

// Command-line flags.
var (
	dayFlag        = flag.String("day", "", "day to dump in yyyy-mm-dd format (default: today)")
	dumpDir        = flag.String("dir", "/tmp", "destination directory")
	shouldCompress = flag.Bool("compress", false, "gzip compress output files")
)

const YYYY_MM_DD = "2006-01-02"

func getDay() time.Time {
	if *dayFlag != "" {
		d, err := time.Parse(YYYY_MM_DD, *dayFlag)
		must(err)
		return d
	}
	return time.Now()
}

func formatNullable(v interface{}) string {
	switch t := v.(type) {
	case *string:
		if t != nil {
			return *t
		}
	case *int:
		if t != nil {
			return strconv.Itoa(*t)
		}
	default:
		return fmt.Sprintf("%s", t)
	}
	return ""
}

func openCsvFile(name string, day time.Time) (filename string, w io.WriteCloser) {
	var err error
	filename = fmt.Sprintf("%s%c%s_%s.csv", *dumpDir,
		os.PathSeparator, day.Format(YYYY_MM_DD), name)
	if *shouldCompress {
		filename += ".gz"
	}
	w, err = os.Create(filename)
	must(err)
	if *shouldCompress {
		w = gzip.NewWriter(w)
	}
	return
}

func must(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func writeDealsCsv(db *scrape.DB, day time.Time) {
	log.Printf("retrieving deals")
	rows, err := db.GetDealDailySnapshots(day)
	must(err)

	records := make([][]string, 0, len(rows)+1)
	records = append(records, []string{
		"Site",
		"DealID",
		"Day",
		"Description",
		"Category",
		"Subcategory",
		"Locale",
		"OriginalPrice",
		"DiscountPrice",
		"NumSold",
		"IsExpired",
		"IsAdult",
	})
	for _, r := range rows {
		records = append(records, []string{
			r.Site,
			strconv.FormatInt(r.DealID, 10),
			r.Day.Format(YYYY_MM_DD),
			formatNullable(r.Description),
			formatNullable(r.Category),
			formatNullable(r.Subcategory),
			formatNullable(r.Locale),
			formatNullable(r.OriginalPrice),
			formatNullable(r.DiscountPrice),
			formatNullable(r.NumSold),
			strconv.FormatBool(r.IsExpired),
			strconv.FormatBool(r.IsAdult),
		})
	}
	writeCsv("deals", day, records)
}

func writeOptionsCsv(db *scrape.DB, day time.Time) {
	log.Printf("retrieving options")
	rows, err := db.GetOptionDailySnapshots(day)
	must(err)

	records := make([][]string, 0, len(rows)+1)
	records = append(records, []string{
		"Site",
		"DealID",
		"OptionID",
		"Day",
		"Description",
		"Price",
		"NumAvailable",
		"NumSold",
	})
	for _, r := range rows {
		records = append(records, []string{
			r.Site,
			strconv.FormatInt(r.DealID, 10),
			strconv.FormatInt(r.OptionID, 10),
			r.Day.Format(YYYY_MM_DD),
			formatNullable(r.Description),
			formatNullable(r.Price),
			formatNullable(r.NumAvailable),
			formatNullable(r.NumSold),
		})
	}
	writeCsv("options", day, records)
}

func writeCsv(what string, day time.Time, records [][]string) {
	filename, fileWriter := openCsvFile(what, day)
	log.Printf("writing %s to %s", what, filename)
	csvWriter := csv.NewWriter(fileWriter)
	must(csvWriter.WriteAll(records))
	csvWriter.Flush()
	must(fileWriter.Close())
}

func main() {
	flag.Parse()

	day := getDay()
	log.Printf("dumping snapshots for %s", day.Format(YYYY_MM_DD))

	uri := scrape.GetMySQLConnectionURI()
	log.Printf("connecting to database: %s", uri)
	db, err := scrape.OpenDatabase(uri)
	must(err)

	writeDealsCsv(db, day)
	writeOptionsCsv(db, day)
}
