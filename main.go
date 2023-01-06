package main

import (
	
	"encoding/xml"
	
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	_ "time"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	_ "context"
	"strconv"
	
	"github.com/gocolly/colly"
	

)

var m MembersOfParliament
//var memberMap map[string][]string
const bingURL = "https://api.cognitive.microsoft.com/bing/v7.0/search"

func main() {
	//memberMap = make(map[string][]string)

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/sa", func (w http.ResponseWriter, r *http.Request) {
		memberId , ok := r.URL.Query()["memberId"]
		if !ok || len(memberId[0]) < 1 {
			log.Println("Url Param 'memberId' is missing")
			return
		}
		id ,err := strconv.Atoi(memberId[0])
		if err != nil {
			log.Println("Url Param 'memberId' is not a number")
			return
		}
		sentimentAnalysisHandler(w,r, id)
	
	})
	resp, errResp := http.Get("https://www.ourcommons.ca/Members/en/search/xml?view=List")
	if errResp != nil {
		log.Fatal(errResp)
	}
	respBytes, errBytes := ioutil.ReadAll(resp.Body)
	if errBytes != nil {
		log.Fatal(errBytes)
	}
	xml.Unmarshal(respBytes, &m)
	resp.Body.Close()
	log.Println("Response recorded.")
	http.ListenAndServe(":8080", router)


	for _, member := range m.Members {
		go literaturePopulator(member.Id, m.Members)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := "templates/index.html"
	t, err := template.ParseFiles(p)
	if err != nil {
		log.Fatal(err)
	}
	counter := 1
	for i:= range m.Members {
		m.Members[i].Id = counter
		counter +=1
	}
	for i:= range m.Members {
		m.Members[i].Fullname = m.Members[i].Firstname + " " + m.Members[i].Lastname
	}		
		t.Execute(w, m.Members)
	
}
func sentimentAnalysisHandler(w http.ResponseWriter, r *http.Request, memberId int) {
	p := "templates/sentimentanalysis.html"
	t, err := template.ParseFiles(p)
	if err != nil {
		log.Fatal(err)
	}

	// Find the member with the matching ID
	var member Member
	for _, m := range m.Members {
		if m.Id == memberId {
			member = m
			break
		}
	}

	t.Execute(w, member)
	fmt.Printf("Selected member ID: %v", member)
}
func literaturePopulator(id int, members []Member) {
	var member Member
	for _, m := range members {
		if m.Id == id {
			member = m
			break
		}
	}
	query := fmt.Sprintf("%s %s site:news.google.com", member.Fullname, member.Party)
	searchURL := fmt.Sprintf("https://opensearch.org/search?q=%s", url.QueryEscape(query))
	c := colly.NewCollector()
	c.OnHTML("h3", func(e *colly.HTMLElement) {
		member.Results = append(member.Results, Result{
			Headline: e.Text,
			Link:     e.Attr("href"),
		})
	})
	c.Visit(searchURL)
}




type Member struct {
	Id int
	Firstname string `xml:"PersonOfficialFirstName"`
	Lastname  string `xml:"PersonOfficialLastName"`
	Fullname string
	Party     string `xml:"CaucusShortName"`
	Constituency string `xml:"ConstituencyName"`
	Results []Result  
}

type Result struct {
	Headline string
	Link string
	Content string
}

type MembersOfParliament struct {
	XMLName xml.Name `xml:"ArrayOfMemberOfParliament"`
	Members []Member `xml:"MemberOfParliament"`
}

