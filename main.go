package main

import (
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

type RSSFeed struct {
	Items []RSSItem `xml:"channel>item"`
	Link  *url.URL  `xml:"link"`
	ID    uuid.UUID
}

type RSSItem struct {
	Title   string   `xml:"title"`
	Link    *url.URL `xml:"link"`
	PubDate string   `xml:"pubDate"`
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "index", myRSSReader.Feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
	requestedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	index := -1
	for i, v := range myRSSReader.Feeds {
		if v.ID == requestedID {
			index = i
			break
		}
	}
	if index != -1 {
		myRSSReader.Feeds = append(myRSSReader.Feeds[:index], myRSSReader.Feeds[index+1:]...)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	requestURL, err := url.ParseRequestURI(r.FormValue("inputField"))
	if err != nil {
		http.Error(w, "Not a valid url", http.StatusBadRequest)
		return
	}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "li", RSSFeed{Link: requestURL, ID: uuid.New()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	myRSSReader.Feeds = append(myRSSReader.Feeds, RSSFeed{Link: requestURL, ID: uuid.New()})
}

type RSSReader struct {
	Feeds []RSSFeed
}

var sampleURL, err = url.Parse("https://www.test.com")

var myRSSReader = RSSReader{Feeds: []RSSFeed{{Link: sampleURL, Items: []RSSItem{{Title: "Something cool is happenning", PubDate: "2024-01-29"}}}}}

func main() {
	baseUrl := "https://lorem-rss.herokuapp.com/feed?unit=second&interval=30"

	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", handlerFunc)
	router.HandleFunc("/addFeed", handleAdd)
	router.HandleFunc("/deleteFeed/{id}", handleDelete)
	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	fmt.Println("Server listening...")
	server.ListenAndServe()

	// GET the RSS feed (in this case Lorem RSS)
	resp, err := http.Get(baseUrl)
	if err != nil {
		fmt.Println("Error getting site", baseUrl)
	}
	defer resp.Body.Close()

	// Read response body into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
	}

	// parse the RSS
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		fmt.Println("Failed to parse RSS feed")
	}
	fmt.Println(feed)
}
