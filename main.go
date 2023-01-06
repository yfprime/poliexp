package main

import (
	"html"
	"encoding/xml"
	
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	_ "time"
	"github.com/gorilla/mux"
	"html/template"
	_ "context"
	"strconv"
	
	

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
	// Build the search query
	query := url.QueryEscape(fmt.Sprintf("%s %s", members[id-1].Fullname, members[id-1].Party))

	// Send the search request to OpenSearch
	resp, err := http.Get(fmt.Sprintf("https://opensearch.org/search?q=%s", query))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// Parse the search results
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	
	// Find the search results in the page
	results := []Result{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" && n.Parent.Data == "li" {
			// This is a search result headline
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "title" {
					// This is the link element for the search result
					for _, a := range n.FirstChild.Attr {
						if a.Key == "href" {
							// Add the search result to the results slice
							results = append(results, Result{
								Headline: n.FirstChild.FirstChild.Data,
								Link:     a.Val,
								Content:  "",
							})
							break
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// Set the Results field for the member
	members[id-1].Results = results
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

