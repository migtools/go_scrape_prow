package main

import (
	"testing"
)

func Test_start_geziyor(t *testing.T) {

	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "Negative test",
			arg:  "https://google1.com/",
			want: 0,
		},
		{
			name: "Postive test test",
			arg:  "https://prow.ci.openshift.org/?type=periodic&job=*oadp*",
			want: 15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start_geziyor(tt.arg)
			if got := len(all_jobs); got != tt.want {
				t.Errorf("Nuber of jobs  = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_getProwJobs(t *testing.T) {
// 	type args struct {
// 		g *geziyor.Geziyor
// 		r *client.Response
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			getProwJobs(tt.args.g, tt.args.r)
// 		})
// 	}
// }
