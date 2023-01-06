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
	
	"html/template"
	_ "context"
	"strconv"
	"encoding/json"
	

)

var m MembersOfParliament
//var memberMap map[string][]string
const mozURL = "https://api.moz.com/linkscape/url-metrics/"

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
	
	for _, member := range m.Members {
		fmt.Println("Processing member:", member.Id)
		go literaturePopulator(member.Id, m.Members)
	}




	http.ListenAndServe(":8080", router)
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
	
	// Build the search query
	query := url.QueryEscape(member.Fullname + " " + member.Party)
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&pretty=1", query)
	
	// Send the request
	resp, err := http.Get(searchURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	
	// Parse the response
	var data struct {
		RelatedTopics []struct {
			Text string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println(err)
		return
	}
	
	// Populate the Results field
	for _, topic := range data.RelatedTopics {
		member.Results = append(member.Results, Result{
			Headline: topic.Text,
			Link: topic.FirstURL,
		})
	}
	fmt.Println(members[5].Results)
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

