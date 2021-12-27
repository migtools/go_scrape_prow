package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
)

type Jobs struct {
	id   string
	pass bool
	url  string
}

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
		log.Printf(my_url)

		u, _ := url.Parse(my_url)
		id := u.Path[strings.LastIndex(u.Path, "/")+1:]
		log.Printf(id)

		j := Jobs{id, false, u.String()}
		log.Printf("%+v\n", j)

	})

}
