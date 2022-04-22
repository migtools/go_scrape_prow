package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/smallfish/simpleyaml"
)

func Test_isFlake(t *testing.T) {

	tests := []struct {
		name string
		arg  Job
		want bool
	}{
		{
			name: "Positive Test is Flake - periodic",
			arg: Job{"1498722881352241152", "", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_release/25366/rehearse-25366-periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack/1498722881352241152",
				"https://prow.ci.openshift.org/prowjob?prowjob=2b1ee221-998b-11ec-8be4-0a580a8119a8", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_release/25366/rehearse-25366-periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack/1498722881352241152/",
				"2022-03-01T18:12:34Z", "2022-03-01T18:34:41Z", "rehearse-25366-periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack", "https://github.com/openshift/release/pull/25366", "rehearse", "aws", ""},
			want: true,
		},
		{
			name: "Positive Test is not a Flake - periodic",
			arg: Job{"1504306580781273088", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure-periodic-slack/1504306580781273088",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure-periodic-slack/1504306580781273088/",
				"2022-03-01T18:12:34Z", "2022-03-01T18:34:41Z", "periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure-periodic-slack", "", "periodic", "azure", ""},
			want: false,
		},
		{
			name: "Negative data Test for Flake - periodic",
			arg: Job{"1498722881352241152", "", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_release/25366/rehearse-25366-periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack/1498722881352241152",
				"https://prow.ci.openshift.org/prowjob?prowjob=2b1ee221-998b-11ec-8be4-0a580a8119a8", "abc",
				"2022-03-01T18:12:34Z", "2022-03-01T18:34:41Z", "rehearse-25366-periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack", "https://github.com/openshift/release/pull/25366", "rehearse", "aws", ""},
			want: false,
		},
		{
			name: "Positive data Test is not a Flake - pull-name-update",
			arg: Job{"1513678249626963968", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/566/pull-ci-openshift-oadp-operator-master-4.7-operator-e2e-gcp/1513678249626963968",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/566/pull-ci-openshift-oadp-operator-master-4.7-operator-e2e-gcp/1513678249626963968/",
				"2022-03-01T18:12:34Z", "2022-03-01T18:34:41Z", "pull-ci-openshift-oadp-operator-master-4.7-operator-e2e-gcp", "", "pull", "gcp", ""},
			want: false,
		},
		{
			name: "Positive data Test is a Flake - pull-name-update",
			arg: Job{"1505907147571990528", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_release/27052/rehearse-27052-pull-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure/1505907147571990528",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_release/27052/rehearse-27052-pull-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure/1505907147571990528/",
				"2022-03-01T18:12:34Z", "2022-03-01T18:34:41Z", "rehearse-27052-pull-ci-openshift-oadp-operator-master-4.9-operator-e2e-azure", "", "rehearse", "azure", ""},
			want: true,
		},
		{
			name: "Positive for Unit Test - periodic-unit-test",
			arg: Job{"1517352534933508096", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/periodic-ci-openshift-oadp-operator-master-4.8-operator-unit-test-periodic/1517352534933508096",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/logs/periodic-ci-openshift-oadp-operator-master-4.8-operator-unit-test-periodic/1517352534933508096/",
				"", "", "periodic-ci-openshift-oadp-operator-master-4.8-operator-unit-test-periodic", "", "periodic", "", ""},
			want: false,
		},
		{
			name: "Positive for Index Test - pull-index-test",
			arg: Job{"1514611520795840512", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/639/pull-ci-openshift-oadp-operator-master-4.9-ci-index/1514611520795840512",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/639/pull-ci-openshift-oadp-operator-master-4.9-ci-index/1514611520795840512/",
				"", "", "pull-ci-openshift-oadp-operator-master-4.9-ci-index", "", "pull", "", ""},
			want: false,
		},
		{
			name: "Positive for Unit Test - pull-unit-test",
			arg: Job{"1517182064560967680", "failure", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/624/pull-ci-openshift-oadp-operator-master-4.7-operator-unit-test/1517182064560967680",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_oadp-operator/624/pull-ci-openshift-oadp-operator-master-4.7-operator-unit-test/1517182064560967680/",
				"", "", "pull-ci-openshift-oadp-operator-master-4.7-operator-unit-test", "", "pull", "", ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making logs.txt ")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				isFlake(tt.arg)
				if _, err := os.Stat("testData/logs.txt"); err != nil {
					t.Errorf("Error log failed for Empty name")
				} else {
					os.Remove("testData/logs.txt")
				}
			} else {
				if got := isFlake(tt.arg); got != tt.want {
					t.Errorf("isFlake() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_getJobType(t *testing.T) {
	tests := []struct {
		name string
		arg  Job
		want JobType
	}{

		{
			name: "Periodic job type test",
			arg:  Job{"1234567", "", 4, "", "", "", "", "", "periodic-ci-openshift-oadp-operator-master-4.8-operator-e2e-gcp-periodic-slack", "not_found", "", "", ""},
			want: PERIODIC,
		},
		{
			name: "Rehearse job type test",
			arg:  Job{"1234567", "", 4, "", "", "", "", "", "rehearse-25366-pull-ci-openshift-oadp-operator-master-4.9-operator-unit-test", "not_found", "", "", ""},
			want: REHEARSE,
		},
		{

			name: "Pull job type test",
			arg:  Job{"1234567", "", 4, "", "", "", "", "", "pull-ci-openshift-oadp-operator-master-4.8-operator-e2e-gcp-periodic-slack", "not_found", "", "", ""},
			want: PULL,
		},
		{
			name: "Unknown job type test",
			arg:  Job{"1234567", "", 4, "", "", "", "", "", "test", "not_found", "", "", ""},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := getJobType(tt.arg); got != tt.want {
				t.Errorf("getJobType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nameJob(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		job      Job
		want     string
	}{
		{
			name:     "Negative Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/negative.yaml",
		},
		{
			name:     "Positive Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/positive.yaml",
			want:     "periodic-ci-openshift-oadp-operator-master-4.9-operator-e2e-aws-periodic-slack",
		},
	}

	for _, tt := range tests {
		yaml_data, err := ioutil.ReadFile(tt.fileName)
		if err != nil {
			fmt.Println("err in file  ")
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)

		if err != nil {
			fmt.Println("err in yaml ")
		}
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making logs.txt ")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				nameJob(yaml, tt.job)
				if _, err := os.Stat("testData/logs.txt"); err != nil {
					t.Errorf("Error log failed for Empty name")
				} else {
					os.Remove("testData/logs.txt")
				}
			} else {
				if got := nameJob(yaml, tt.job); got != tt.want {
					t.Errorf("nameJob() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func Test_startTime(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		job      Job
		want     string
	}{
		{
			name:     "Negative Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/badData.yaml",
		},
		{
			name:     "Positive Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/positive.yaml",
			want:     "2022-03-23T04:01:07Z",
		},
	}

	for _, tt := range tests {
		yaml_data, err := ioutil.ReadFile(tt.fileName)
		if err != nil {
			fmt.Println("err in file  ")
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)

		if err != nil {
			fmt.Println("err in yaml ")
		}
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/badData_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making badData_logs.txt for startTime")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				getStartTime(yaml, tt.job)
				if _, err := os.Stat("testData/badData_logs.txt"); err != nil {
					t.Errorf("Error log failed for start Time")
				} else {
					os.Remove("testData/badData_logs.txt")
				}
			} else {
				if got := getStartTime(yaml, tt.job); got != tt.want {
					t.Errorf("getstartTime() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func Test_getStatus(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		job      Job
		want     string
	}{
		{
			name:     "Negative status test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/badData.yaml",
		},
		{
			name:     "Positive success status test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/positive.yaml",
			want:     "success",
		},
		{
			name: "Positive failure-flake status test",
			job: Job{"1508830368151638016", "", 0, "https://prow.ci.openshift.org/view/gs/origin-ci-test/pr-logs/pull/openshift_release/27052/rehearse-27052-pull-ci-openshift-oadp-operator-master-4.8-operator-e2e-azure/1508830368151638016",
				"", "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test/pr-logs/pull/openshift_release/27052/rehearse-27052-pull-ci-openshift-oadp-operator-master-4.8-operator-e2e-azure/1508830368151638016/",
				"", "", "rehearse-27052-pull-ci-openshift-oadp-operator-master-4.8-operator-e2e-azure", "not_found", "pull", "azure", ""},
			fileName: "testData/statusFailure.yaml",
			want:     "flake",
		},
		{
			name:     "Positive Pending status test",
			fileName: "testData/statusPending.yaml",
			want:     "pending",
		},
	}

	for _, tt := range tests {
		yaml_data, err := ioutil.ReadFile(tt.fileName)
		if err != nil {
			fmt.Println("err in file  ")
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)

		if err != nil {
			fmt.Println("err in yaml ")
		}
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/badData_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making badData_logs.txt for getStatus")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				getStatus(yaml, tt.job)
				if _, err := os.Stat("testData/badData_logs.txt"); err != nil {
					t.Errorf("Error log failed for geStatus")
				} else {
					os.Remove("testData/badData_logs.txt")
				}
			} else {
				if got := getStatus(yaml, tt.job); got != tt.want {
					t.Errorf("getStatus() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func Test_endTime(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		job      Job
		want     string
	}{
		{
			name:     "Negative Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/badData.yaml",
		},
		{
			name:     "Positive Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/positive.yaml",
			want:     "2022-03-23T05:28:36Z",
		},
	}

	for _, tt := range tests {
		yaml_data, err := ioutil.ReadFile(tt.fileName)
		if err != nil {
			fmt.Println("err in file  ")
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)

		if err != nil {
			fmt.Println("err in yaml ")
		}
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/badData_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making badData_logs.txt for endTime  ")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				getEndTime(yaml, tt.job)
				if _, err := os.Stat("testData/badData_logs.txt"); err != nil {
					t.Errorf("Error log failed for End Time")
				} else {
					os.Remove("testData/badData_logs.txt")
				}

			} else {
				if got := getEndTime(yaml, tt.job); got != tt.want {
					t.Errorf("getEndTime() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func Test_cluster_profile(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		job      Job
		want     string
	}{
		{
			name:     "Negative Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/badData.yaml",
		},
		{
			name:     "Positive Name test",
			job:      Job{"1506481143975776256", "", 4, "", "", "", "", "", "", "not_found", "", "", ""},
			fileName: "testData/positive.yaml",
			want:     "aws",
		},
	}

	for _, tt := range tests {
		yaml_data, err := ioutil.ReadFile(tt.fileName)
		if err != nil {
			fmt.Println("err in file  ")
		}
		yaml, err := simpleyaml.NewYaml(yaml_data)

		if err != nil {
			fmt.Println("err in yaml ")
		}
		t.Run(tt.name, func(t *testing.T) {
			if strings.HasPrefix(tt.name, "Negative") {
				unitTestfile, err := os.OpenFile("testData/badData_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					fmt.Println("err in making badData_logs.txt for clusterProfile  ")
				}
				ErrorLogger = log.New(unitTestfile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
				getEndTime(yaml, tt.job)
				if _, err := os.Stat("testData/badData_logs.txt"); err != nil {
					t.Errorf("Error log failed for cluster Profile")
				} else {
					os.Remove("testData/badData_logs.txt")
				}

			} else {
				if got := getClusterProfile(yaml, tt.job); got != tt.want {
					t.Errorf("getClusterProfile() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}

func Test_getStateInt(t *testing.T) {

	tests := []struct {
		name   string
		status string
		want   int
	}{
		{
			name:   "success state int",
			status: "success",
			want:   0,
		},
		{
			name:   "failure state int",
			status: "failure",
			want:   2,
		},
		{
			name:   "aborted state int",
			status: "aborted",
			want:   3,
		},
		{
			name:   "flake state int",
			status: "flake",
			want:   4,
		},
		{
			name:   "pending state int",
			status: "pending",
			want:   1,
		},
		{
			name:   "unknown state int",
			status: "unknown",
			want:   10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStateInt(tt.status); got != tt.want {
				t.Errorf("getStateInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
