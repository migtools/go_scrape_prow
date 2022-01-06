package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/smallfish/simpleyaml"
	"github.com/theckman/yacspin"
)

type Job struct {
	id            string
	state         string
	state_int     int
	log_url       string
	log_yaml      string
	log_artifacts string
	start_time    string
	end_time      string
	name          string
	pull_request  string
}

var all_jobs = make(map[string]Job)

func main() {
	// get cli args
	var url_to_scrape string
	var print_for_human bool

	flag.StringVar(&url_to_scrape, "url_to_scrape", "https://prow.ci.openshift.org/?job=*oadp*", "prow url to scrape, e.g. ")
	flag.BoolVar(&print_for_human, "print_for_human", false, "print for a human, not influxdb")

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

func start_spinner() (*yacspin.Spinner, error) {
	// meh have some fun
	cfg := yacspin.Config{
		Frequency:         500 * time.Millisecond,
		Writer:            nil,
		ShowCursor:        false,
		HideCursor:        false,
		SpinnerAtEnd:      false,
		CharSet:           yacspin.CharSets[59],
		Prefix:            " ",
		Suffix:            " ",
		SuffixAutoColon:   true,
		Message:           " Getting your jobs",
		ColorAll:          true,
		Colors:            []string{"fgYellow"},
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopMessage:       "done",
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
		StopFailMessage:   "failed",
		NotTTY:            false,
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to make spinner from struct: %s", err)
	}

	err = spinner.Start()
	time.Sleep(1 * time.Second)
	// end fun
	return spinner, err

}

func stopSpinnerOnSignal(spinner *yacspin.Spinner) {
	// ensure we stop the spinner before exiting, otherwise cursor will remain
	// hidden and terminal will require a `reset`
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh

		spinner.StopFailMessage("interrupted")

		// ignoring error intentionally
		_ = spinner.StopFail()

		os.Exit(0)
	}()
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

		this_job := Job{id, "", 4, u.String(), "", "", "", "", "", "todo"}
		all_jobs[id] = this_job

	})
	getJobDetails(all_jobs)
	getYAMLDetails(all_jobs)
}

func getJobDetails(all_jobs map[string]Job) {
	log_yaml_base := "https://prow.ci.openshift.org"
	for id, job := range all_jobs {
		//log.Printf("%+v\n", job)
		//log.Printf(id)
		response, err := http.Get(job.log_url)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		// Create a goquery document from the HTTP response
		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatal("Error loading HTTP response body. ", err)
		}

		// Get the Prow job YAML link
		document.Find("#links-card > a:nth-child(2)").Each(func(i int, s *goquery.Selection) {
			yaml_link, ok := s.Attr("href")
			if ok {
				job.log_yaml = log_yaml_base + yaml_link
				//log.Printf(job.log_yaml)
			}
		})

		// Get the Prow job Artifacts link
		document.Find("#links-card > a:nth-child(3)").Each(func(i int, s *goquery.Selection) {
			artifact_link, ok := s.Attr("href")
			if ok {
				job.log_artifacts = artifact_link
				//log.Printf(job.log_artifacts)
			}
		})
		all_jobs[id] = job
	}
}

func getYAMLDetails(all_jobs map[string]Job) {
	for id, job := range all_jobs {
		//log.Printf(id)
		response, err := http.Get(job.log_yaml)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()
		yaml_data, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)
		if err != nil {
			panic(err)
		}
		status, err := yaml.Get("status").Get("state").String()
		if err != nil {
			panic(err)
		}

		// get state
		//  0 = success, 1 = pending, 2 = failed 3 = aborted / other
		state_int := 4
		state := ""
		switch status {
		case "success":
			state_int = 0
			state = "success"
		case "pending":
			state_int = 1
			state = "pending"
		case "failed":
			state_int = 2
			state = "failed"
		default:
			state_int = 4
			state = "unknown"
		}

		job.state = state
		job.state_int = state_int

		// Get Start / Stop time
		start, err := yaml.Get("status").Get("startTime").String()
		if err != nil {
			panic(err)
		}
		end, err := yaml.Get("status").Get("completionTime").String()
		if err != nil {
			panic(err)
		}
		name, err := yaml.Get("metadata").Get("annotations").Get("prow.k8s.io/job").String()
		if err != nil {
			panic(err)
		}

		job.start_time = start
		job.end_time = end
		job.name = name

		// update object w/ success, failure status
		all_jobs[id] = job
	}
}

func print_human(all_jobs map[string]Job) {
	for _, my_job := range all_jobs {
		fmt.Printf("%+v\n", my_job)
	}
}

func print_db(all_jobs map[string]Job) {
	for _, my_job := range all_jobs {
		// datestamps
		st, _ := time.Parse(time.RFC3339, my_job.start_time)
		et, _ := time.Parse(time.RFC3339, my_job.end_time)
		duration := fmt.Sprint(et.Sub(st).Seconds())
		timestamp := fmt.Sprint(st.Unix() * 1000000000)

		// log.Printf(my_job.start_time)
		// log.Printf(my_job.end_time)
		// log.Printf(st.String())
		// log.Printf(et.String())
		// log.Printf("%f", st.Unix())

		// influxdb line format
		// https://docs.influxdata.com/influxdb/v2.1/reference/syntax/line-protocol/

		build_string := "build," +
			"job_name=" + my_job.name +
			",build_id=" + my_job.id +
			",pull_request=" + my_job.pull_request +
			",start_time=" + my_job.start_time +
			",end_time=" + my_job.end_time +
			",duration=" + duration + //seconds
			",state_int=" + strconv.Itoa(my_job.state_int) +
			",state=" + my_job.state +
			" " + //space required for influxdb format
			"job_name=" + strconv.Quote(my_job.name) +
			",build_id=" + my_job.id +
			",pull_request=" + strconv.Quote(my_job.pull_request) +
			",start_time=" + strconv.Quote(my_job.start_time) +
			",end_time=" + strconv.Quote(my_job.end_time) +
			",duration=" + duration +
			",state_int=" + strconv.Itoa(my_job.state_int) +
			",state=" + strconv.Quote(my_job.state) +
			",log=" + strconv.Quote(my_job.log_url) +
			" " +
			timestamp // this timestap is the job recorded timestamp in influx

		fmt.Println(build_string)
	}
}