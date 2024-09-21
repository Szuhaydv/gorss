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
	Link  string    `xml:"channel>link"`
	ID    uuid.UUID
}

type RSSItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func fetchFeed(url string, myFeed *RSSFeed) {

	// GET the RSS feed (in this case Lorem RSS)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting site")
	}
	defer resp.Body.Close()

	// Read response body into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
	}

	// parse the RSS
	var newFeed RSSFeed
	if err := xml.Unmarshal(body, &newFeed); err != nil {
		fmt.Println("Failed to parse RSS feed")
	}
	*myFeed = newFeed
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	for i, feed := range myRSSReader.Feeds {
		fetchFeed(feed.Link, &myRSSReader.Feeds[i])
	}

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
	urlString := requestURL.String()
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "li", RSSFeed{Link: urlString, ID: uuid.New()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	myRSSReader.Feeds = append(myRSSReader.Feeds, RSSFeed{Link: urlString, ID: uuid.New()})
}

type RSSReader struct {
	Feeds []RSSFeed
}

func createReader() (RSSReader, error) {
	return RSSReader{Feeds: []RSSFeed{
		{Link: "https://lorem-rss.herokuapp.com/feed?unit=second&interval=30"},
	}}, nil
}

var myRSSReader, _ = createReader()

func main() {
	router := http.NewServeMux()
	router.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("POST /addFeed", handleAdd)
	router.HandleFunc("DELETE /deleteFeed/{id}", handleDelete)
	router.HandleFunc("GET /{$}", handlerFunc)
	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	fmt.Println("Server listening...")
	server.ListenAndServe()
}
