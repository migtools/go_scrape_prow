package main

//Job is the requirement to test the data. So converting data TestJob to Job.
func TestJobTOJob(tstJob TestJob) Job {
	job := Job{tstJob.Id, tstJob.State, tstJob.State_int, tstJob.Log_url, tstJob.Log_yaml,
		tstJob.Log_artifacts, tstJob.Start_time, tstJob.End_time, tstJob.Name, tstJob.Pull_request, tstJob.Job_type, tstJob.Cloud_profile, tstJob.Test_type, tstJob.Target_Name}
	return job
}

//This struct is duplicate of Job. TestJob is only to read data from yaml file for UNIT TEST.
type TestJob struct {
	Id            string  `yaml:"jobId"`
	State         string  `yaml:"jobState"`
	State_int     int     `yaml:"jobState_int"`
	Log_url       string  `yaml:"jobLog_url"`
	Log_yaml      string  `yaml:"jobLog_yaml"`
	Log_artifacts string  `yaml:"jobLog_artifacts"`
	Start_time    string  `yaml:"jobStart_time"`
	End_time      string  `yaml:"jobEnd_time"`
	Name          string  `yaml:"jobName"`
	Pull_request  string  `yaml:"jobPull_request"`
	Job_type      JobType `yaml:"jobType"`
	Cloud_profile string  `yaml:"jobCloud_profile"`
	Test_type     string  `yaml:"testType"`
	Target_Name   string  `yaml:"targetName"`
}
