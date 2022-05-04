package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/smallfish/simpleyaml"
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
	job_type      JobType
	cloud_profile string // Supposed to remove this; however, it can help differentiate tests by cluster profiles
	test_type     string
}

var ErrorLogger *log.Logger

type JobType string

const (
	PULL     JobType = "pull"
	PERIODIC JobType = "periodic"
	REHEARSE JobType = "rehearse"
)

var all_jobs = make(map[string]Job)

// if build-log.txt does not exist, then failure is FLAKE.
func isFlake(job Job) bool {
	if job.job_type == PERIODIC || strings.Contains(job.name, "periodic") {
		buid_log_url := ""
		if strings.Contains(job.name, "e2e") {
			buid_log_url = job.log_artifacts + "artifacts/operator-e2e-" + job.cloud_profile + "-periodic-slack/e2e/"
		} else if strings.Contains(job.name, "unit") {
			buid_log_url = job.log_artifacts + "artifacts/operator-unit-test-periodic/unit-periodic/"
		} else {
			return false
		}
		buildlog_response, err := http.Get(buid_log_url)
		if err != nil {
			print_human_row(job)
			return false
		}
		defer buildlog_response.Body.Close()

		buidlog, err := io.ReadAll(buildlog_response.Body)
		if err != nil {
			print_human_row(job)
			return false
		}
		if strings.Contains(string(buidlog), "build-log.txt") {
			return false
		}
	} else if job.job_type == PULL || strings.Contains(job.name, "pull") {
		//artifacts/operator-e2e-gcp/e2e/
		buid_log_url := ""
		if strings.Contains(job.name, "e2e") {
			buid_log_url = job.log_artifacts + "artifacts/operator-e2e-" + job.cloud_profile + "/e2e/"
		} else if strings.Contains(job.name, "unit") {
			buid_log_url = job.log_artifacts + "artifacts/operator-unit-test/unit/"
		} else {
			return false
		}
		buildlog_response, err := http.Get(buid_log_url)
		if err != nil {
			print_human_row(job)
			return false
		}
		defer buildlog_response.Body.Close()

		buidlog, err := io.ReadAll(buildlog_response.Body)
		if err != nil {
			print_human_row(job)
			return false
		}
		if strings.Contains(string(buidlog), "build-log.txt") {
			return false
		}
	}
	return true
}

//get JobType by passing job as argument
func getJobType(job Job) JobType {
	if strings.HasPrefix(job.name, "pull") {
		return PULL
	} else if strings.HasPrefix(job.name, "rehearse") {
		return REHEARSE
	} else if strings.HasPrefix(job.name, "periodic") {
		return PERIODIC
	} else {
		return "unknown"
	}
}

func getClusterProfile(yaml *simpleyaml.Yaml, job Job) string {
	clusterProfile, err := yaml.Get("metadata").Get("labels").Get("ci-operator.openshift.io/cloud").String()
	if err != nil {
		print_human_row(job)
	}
	if clusterProfile == "azure4" {
		clusterProfile = "azure"
	}
	return clusterProfile
}

func nameJob(yaml *simpleyaml.Yaml, job Job) string {
	name, _ := yaml.Get("metadata").Get("annotations").Get("prow.k8s.io/job").String()

	if len(name) < 1 {
		print_human_row(job)
	}
	return name
}

func getEndTime(yaml *simpleyaml.Yaml, job Job) string {
	// Get Start / Stop time

	end, err := yaml.Get("status").Get("completionTime").String()
	if err != nil {
		print_human_row(job)
	}
	return end
}

func getStartTime(yaml *simpleyaml.Yaml, job Job) string {
	start, err := yaml.Get("status").Get("startTime").String()
	if err != nil {
		print_human_row(job)
	}
	return start
}

func getStatus(yaml *simpleyaml.Yaml, job Job) string {
	status, err := yaml.Get("status").Get("state").String()
	if err != nil {
		print_human_row(job)
	}
	//checking if failure is flake.
	if status == "failure" {
		if isFlake(job) {
			status = "flake"
		}
	}

	return status
}

func getStateInt(status string) int {
	state_int := 10

	switch status {
	case "success":
		state_int = 0
	case "pending":
		state_int = 1
	case "failure":
		state_int = 2
	case "aborted":
		state_int = 3
	case "flake":
		state_int = 4
	default:
		state_int = 10
	}
	return state_int
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

		//number of URLs (children ) are diff based on the Job_type
		//length of children helps to scrape URLS
		var children = 0
		document.Find("#links-card").Each(func(i int, s *goquery.Selection) {
			children = s.Children().Length()
			//fmt.Printf("children: %v\n", s.Children().Length())

		})

		// PULL job_type has more than 3 children, and periodic job_type has less than 3 children
		if children > 3 {

			// Get the Prow job YAML link
			document.Find("#links-card > a:nth-child(2)").Each(func(i int, s *goquery.Selection) {
				yaml_link, ok := s.Attr("href")
				if ok {
					job.log_yaml = log_yaml_base + yaml_link
					//log.Printf(job.log_yaml)
				}
			})

			// Get the Prow job Artifacts link
			document.Find("#links-card > a:nth-child(5)").Each(func(i int, s *goquery.Selection) {
				artifact_link, ok := s.Attr("href")
				if ok {
					job.log_artifacts = artifact_link
					//log.Printf(job.log_artifacts)
				}
			})

			// Get pull request
			document.Find("#links-card > a:nth-child(4)").Each(func(i int, s *goquery.Selection) {
				pull_request, ok := s.Attr("href")
				if ok {
					job.pull_request = pull_request
					//log.Printf(job.log_artifacts)
				}
			})

		} else {
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

		}

		all_jobs[id] = job
	}
}

func getYAMLDetails(all_jobs map[string]Job) {
	for id, job := range all_jobs {
		//log.Printf(id)
		response, err := http.Get(job.log_yaml)
		if err != nil {
			print_human_row(job)
		}
		defer response.Body.Close()

		yaml_data, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			print_human_row(job)
		}

		yaml, err := simpleyaml.NewYaml(yaml_data)
		if err != nil {
			print_human_row(job)
		}

		//set name
		job.name = nameJob(yaml, job)

		//get cluster profile
		job.cloud_profile = getClusterProfile(yaml, job)

		//set Start and End Time
		job.start_time = getStartTime(yaml, job)
		job.end_time = getEndTime(yaml, job)

		//setting job_type = periodic, pull, rehearse
		job.job_type = getJobType(job)

		//get status
		job.state = getStatus(yaml, job)

		job.state_int = getStateInt(job.state)

		all_jobs[id] = job
	}
}

func print_human(all_jobs map[string]Job) {
	for _, my_job := range all_jobs {
		fmt.Printf("%+v\n\n\n", my_job)
	}
}

func print_human_row(my_job Job) {
	ErrorLogger.Printf("%+v\n", my_job)
}

func print_db(all_jobs map[string]Job) {
	for _, my_job := range all_jobs {

		// ensure all the rows have the required data for entry
		if my_job.end_time == "" {
			print_human_row(my_job)
			break
		}

		if my_job.name == "" {
			print_human_row(my_job)
			break
		}

		if my_job.state_int == 10 {
			// job state not known.
			// this also causes duplicate entries for build_id
			// the state is eventually written by prow and the job in question
			// will be captured, most likely as aborted.
			print_human_row(my_job)
			break
		}

		// datestamps
		st, _ := time.Parse(time.RFC3339, my_job.start_time)
		et, _ := time.Parse(time.RFC3339, my_job.end_time)
		duration := fmt.Sprint(et.Sub(st).Seconds())
		timestamp := fmt.Sprint(st.Unix() * 1000000000)
		timestamp_int, _ := strconv.Atoi(timestamp)
		if timestamp_int < 1 {
			fmt.Println("timestamp is wrong, break out")
			break
		}

		// log.Printf(my_job.start_time)
		// log.Printf(my_job.end_time)
		// log.Printf(st.String())
		// log.Printf(et.String())
		// log.Printf("%f", st.Unix())

		// influxdb line format
		// https://docs.influxdata.com/influxdb/v2.1/reference/syntax/line-protocol/

		build_string := "build," +
			"job_name=" + my_job.name +
			",build_id=" + strconv.Quote(my_job.id) +
			",pull_request=" + strconv.Quote(my_job.pull_request) +
			",start_time=" + my_job.start_time +
			",end_time=" + my_job.end_time +
			",duration=" + duration + //seconds
			",state_int=" + strconv.Itoa(my_job.state_int) +
			",state=" + my_job.state +
			" " + //space required for influxdb format
			"job_name=" + strconv.Quote(my_job.name) +
			",build_id=" + strconv.Quote(my_job.id) +
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
