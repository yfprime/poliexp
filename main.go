package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"html/template"
	"github.com/go-echarts/go-echarts"
)

var m MembersOfParliament
//var memberMap map[string][]string

func main() {
	//memberMap = make(map[string][]string)

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
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
	log.Println(m)
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
	t.Execute(w, m.Members)
	fmt.Println(m.Members)
}

type Member struct {
	Id int
	Firstname string `xml:"PersonOfficialFirstName"`
	Lastname  string `xml:"PersonOfficialLastName"`
	Party     string `xml:"CaucusShortName"`
	Constituency string `xml:"ConstituencyName"`
	
}

type MembersOfParliament struct {
	XMLName xml.Name `xml:"ArrayOfMemberOfParliament"`
	Members []Member `xml:"MemberOfParliament"`
}
