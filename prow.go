package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/smallfish/simpleyaml"
	"github.com/theckman/yacspin"
)

type Job struct {
	id            string
	pass          bool
	log_url       string
	log_yaml      string
	log_artifacts string
	start_time    string
	end_time      string
	name          string
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
	spinner, err := start_spinner()
	if err != nil {
		log.Printf("spinner failed")
	}

	// start web scraping
	start_geziyor(url_to_scrape)

	// stop spinner
	spinner.Stop()

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
		ColorAll:          false,
		Colors:            []string{},
		CharSet:           yacspin.CharSets[59],
		Prefix:            " ",
		Suffix:            " ",
		SuffixAutoColon:   true,
		Message:           " Getting your jobs",
		StopMessage:       "",
		StopCharacter:     "",
		StopColors:        []string{"fgGreen"},
		StopFailMessage:   "",
		StopFailCharacter: "",
		StopFailColors:    []string{},
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

		this_job := Job{id, false, u.String(), "", "", "", "", ""}
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
		// log.Printf("Value: %#v\n", status)
		if status != "success" {
			job.pass = false
		} else {
			job.pass = true
		}

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
		//log.Printf("\n%+v\n", my_job)
		//log.Println(my_job)
		fmt.Printf("%+v\n", my_job)
	}
}

func print_db(all_jobs map[string]Job) {
	//TO-DO convert time stamps
	// https://pkg.go.dev/github.com/lestrrat-go/strftime#section-documentation
	// https://pkg.go.dev/github.com/levenlabs/golib/timeutil
	for _, my_job := range all_jobs {
		build_string := "build," +
			"job_name=" + my_job.name +
			",build_id=" + my_job.id +
			",pull_request=None" + //TO-DO
			",start_time=" + my_job.start_time +
			",end_time=" + my_job.end_time +
			",duration=" + //TO-DO
			",state_int=" + //TO-DO
			",state=" + //TO-DO
			" " + //space required for influxdb format
			"job_name=" + "\"" + my_job.name + "\"" +
			",build_id=" + "\"" + my_job.id + "\"" +
			",pull_request=None" + //TO-DO
			",start_time=" + "\"" + my_job.start_time + "\"" +
			",end_time=" + "\"" + my_job.end_time + "\"" +
			",duration=" + //TO-DO
			",state_int=" + //TO-DO
			",state=" + //TO-DO
			",log=" + "\"" + my_job.log_url + "\"" +
			" "
			// TO-DO post timestamp here

		fmt.Println(build_string)
	}
}
