package main

import (
	"encoding/json"
	"flag"
	"github.com/launchtime/scrapemonster/cmd"
	"github.com/launchtime/scrapemonster/crawler"
	"github.com/launchtime/scrapemonster/scrape"
	"log"
	"os"
	"time"
)

var (
	scraper   scrape.Scraper
	db        *scrape.DB
	printChan = make(chan []byte)
)

// Command-line flags.
var (
	getOptions  = flag.Bool("o", true, "get deal options")
	maxDepth    = flag.Int("d", 10, "max crawl depth")
	maxParallel = flag.Int("p", 10, "max simultaneous HTTP requests")
	quiet       = flag.Bool("q", false, "do not write JSON to stdout")
	sitename    = flag.String("s", "", "site to crawl")
	startURL    = flag.String("url", "", "override default start url")
	storeInDB   = flag.Bool("db", false, "store results in DB")
	timeout     = flag.Uint("t", 5, "HTTP timeout (seconds)")
	verbose     = flag.Bool("v", false, "verbose output")
)

type (
	dealChannel   chan scrape.DealID
	optionChannel chan []*scrape.Option
)

// chatter writes to the log iff the verbose command-line flag was given.
func chatter(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
}

// printer writes everything it receives from printChan to os.Stdout, as long
// as the quiet command-line flag was not given.
func printer(doneChan chan int) {
	defer func() { doneChan <- 1 }()
	for data := range printChan {
		if !*quiet {
			os.Stdout.Write(data)
			os.Stdout.WriteString("\n")
		}
	}
}

func consumeCrawlerResults(resultChan chan *crawler.Result, dealChan dealChannel) {
	for r := range resultChan {
		deal, err := scraper.ParseDeal(r.URL, r.Body)
		if err != nil {
			log.Print(err)
		}
		if deal == nil {
			continue
		}
		// Optionally print the deal as JSON.
		if !*quiet {
			data, err := json.Marshal(deal)
			if err != nil {
				log.Fatal(err)
			}
			printChan <- data
		}
		// Optionally store the deal in the database.
		if *storeInDB {
			err := db.StoreDeal(deal)
			if err != nil {
				log.Fatal(err)
			}
		}
		// Send the deal ID down the pipeline.
		dealChan <- deal.DealID
	}
	close(dealChan)
}

func optionGetter(dealChan dealChannel, g *crawler.Getter,
	optionChan optionChannel, doneChan chan int) {
	defer func() { doneChan <- 1 }()
	for dealID := range dealChan {
		if *getOptions {
			optionChan <- scraper.GetDealOptions(g, dealID)
		}
	}
}

func consumeOptions(optionChan optionChannel, doneChan chan int) {
	defer func() { doneChan <- 1 }()
	for options := range optionChan {
		for _, option := range options {
			// Optionally print the option as JSON.
			if !*quiet {
				data, err := json.Marshal(option)
				if err != nil {
					log.Fatal(err)
				}
				printChan <- data
			}
			// Optionally store the option in the database.
			if *storeInDB {
				err := db.StoreOption(option)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func main() {
	var (
		resultChan = make(chan *crawler.Result)
		dealChan   = make(dealChannel)
		optionChan = make(optionChannel)
		doneChan   = make(chan int)
		err        error
	)

	flag.Parse()

	scraper = cmd.NewScraper(*sitename)

	// Connect to the database, if requested.
	if *storeInDB {
		uri := scrape.GetMySQLConnectionURI()
		chatter("connecting to database: %s", uri)
		db, err = scrape.OpenDatabase(uri)
		if err != nil {
			log.Fatal(err)
		}
	}

	getter := crawler.NewGetter()
	getter.UserAgent = crawler.UserAgentStrings["MSIE8"]
	getter.Timeout = time.Duration(*timeout) * time.Second
	getter.Verbose = *verbose

	crawler := crawler.New(crawler.SimpleFetcher{
		Getter:       getter,
		URLExtractor: scraper,
	})
	crawler.MaxDepth = *maxDepth
	crawler.MaxParallel = *maxParallel
	crawler.URLTransformer = scraper
	crawler.OutputChan = resultChan
	crawler.Verbose = *verbose

	// Boot up the printer.
	chatter("starting printer")
	go printer(doneChan)

	// Start a bunch of optionGetter goroutines.
	chatter("starting %d optionGetter goroutines", *maxParallel)
	for i := 0; i < *maxParallel; i++ {
		go optionGetter(dealChan, getter, optionChan, doneChan)
	}

	// Consume the output of the optionGetter goroutines.
	chatter("starting consumeOptions")
	go consumeOptions(optionChan, doneChan)

	// Consume crawler results, process deals, forward to optionGetters.
	chatter("starting consumeCrawlerResults")
	go consumeCrawlerResults(resultChan, dealChan)

	// Start crawling.
	if *startURL == "" {
		*startURL = scraper.DefaultStartURL()
	}
	chatter("starting crawl at %s", *startURL)
	if err := crawler.Go(*startURL); err != nil {
		log.Fatal(err)
	}

	// Wait for optionGetter goroutines to finish.
	chatter("waiting for optionGetter goroutines")
	for i := 0; i < *maxParallel; i++ {
		<-doneChan
	}

	// Wait for consumeOptions to finish.
	chatter("waiting for consumeOptions")
	close(optionChan)
	<-doneChan

	// Wait for printer to finish.
	chatter("waiting for printer")
	close(printChan)
	<-doneChan
}
