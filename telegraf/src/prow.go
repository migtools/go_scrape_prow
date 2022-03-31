package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
)

// type error interface {
// 	Error() string
// }

// type errorString struct {
// 	s string
// }

// func (e *errorString) Error() string {
// 	return e.s
// }

func main() {
	//error logging
	file, err := os.OpenFile("/tmp/prow_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// get cli args
	var url_to_scrape string
	var print_for_human bool

	flag.StringVar(&url_to_scrape, "url_to_scrape", "https://prow.ci.openshift.org/?job=*oadp*", "prow url to scrape, e.g. ")
	flag.BoolVar(&print_for_human, "print_for_human", false, "print for a human, not influxdb!!!!!")

	flag.Parse()

	// start dem spinners

	// required if stopSpinner is called
	//var spinner *yacspin.Spinner
	if print_for_human {
		spinner, err := start_spinner()
		spinner = spinner // get around compile error
		if err != nil {
			log.Printf("spinner failed")
		}
	}

	// start web scraping
	start_geziyor(url_to_scrape)

	// This may not be required
	// stop spinner
	// if print_for_human {
	// 	stopSpinnerOnSignal(spinner)
	// }

	// print output
	if print_for_human {
		print_human(all_jobs)
	} else {
		print_db(all_jobs)
	}
}

func start_geziyor(url_to_scrape string) {
	geziyor.NewGeziyor(&geziyor.Options{
		LogDisabled: true,
		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered(url_to_scrape, g.Opt.ParseFunc)
		},
		ParseFunc: getProwJobs,
	}).Start()
}

func getProwJobs(g *geziyor.Geziyor, r *client.Response) {

	// to dump entire html body
	//fmt.Println(string(r.Body))
	rows := r.HTMLDoc.Find("#builds > tbody > tr")
	//log.Printf("length: %d", rows.Length())
	rows.Each(func(i int, s *goquery.Selection) {
		link := s.Find("td:nth-child(8) > a")
		my_url, _ := link.Attr("href")
		//log.Printf(my_url)

		u, _ := url.Parse(my_url)
		id := u.Path[strings.LastIndex(u.Path, "/")+1:]
		//log.Printf(id)

		this_job := Job{id, "", 4, u.String(), "", "", "", "", "", "not_found", "", "", ""}
		all_jobs[id] = this_job

	})
	getJobDetails(all_jobs)
	getYAMLDetails(all_jobs)
}
