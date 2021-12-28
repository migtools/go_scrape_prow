package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/smallfish/simpleyaml"
)

type Job struct {
	id            string
	pass          bool
	log_url       string
	log_yaml      string
	log_artifacts string
	start_time    string
	end_time      string
}

var all_jobs = make(map[string]Job)

func main() {
	geziyor.NewGeziyor(&geziyor.Options{
		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered("https://prow.ci.openshift.org/?job=*oadp*", g.Opt.ParseFunc)
		},
		ParseFunc: getProwJobs,
	}).Start()
}

func getProwJobs(g *geziyor.Geziyor, r *client.Response) {

	// to dump entire html body
	//fmt.Println(string(r.Body))
	rows := r.HTMLDoc.Find("#builds > tbody > tr")
	log.Printf("length: %d", rows.Length())
	rows.Each(func(i int, s *goquery.Selection) {
		link := s.Find("td:nth-child(8) > a")
		my_url, _ := link.Attr("href")
		//log.Printf(my_url)

		u, _ := url.Parse(my_url)
		id := u.Path[strings.LastIndex(u.Path, "/")+1:]
		//log.Printf(id)

		this_job := Job{id, false, u.String(), "", "", "", ""}
		all_jobs[id] = this_job

	})
	getJobDetails(all_jobs)
	getYAMLDetails(all_jobs)
	printAllJobs(all_jobs)
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

		job.start_time = start
		job.end_time = end

		// update object w/ success, failure status
		all_jobs[id] = job
	}
}

func printAllJobs(all_jobs map[string]Job) {
	log.Printf("FINAL DICTIONARY VALUES")
	for _, my_job := range all_jobs {
		log.Printf("%+v\n", my_job)
	}
}
