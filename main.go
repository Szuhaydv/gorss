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
	Title string    `xml:"channel>title"`
	ID    uuid.UUID
	Link  string
}

type RSSItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func fetchFeed(url string) RSSFeed {

	// GET the XML from the provided URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting site")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
	}

	// parse the RSS schema
	var newFeed RSSFeed
	if err := xml.Unmarshal(body, &newFeed); err != nil {
		fmt.Println("Failed to parse RSS feed")
	}
	return newFeed
}

func handlePolling(w http.ResponseWriter, r *http.Request) {

	hashMap := make(map[string]bool)
	for i, feed := range myRSSReader.Feeds {
		// collect each new feed
		newFeed := fetchFeed(feed.Link)
		// populate hashmap for efficieny
		for _, oldItem := range feed.Items {
			hashMap[oldItem.Title] = true
		}
		// diff the new items with the old ones and update data accordingly
		for _, newItem := range newFeed.Items {
			if !hashMap[newItem.Title] {
				myRSSReader.Feeds[i].Items = append(myRSSReader.Feeds[i].Items, newItem)
			}
		}
	}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = tmpl.ExecuteTemplate(w, "item-feed", myRSSReader.Feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {

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

	// Parse the ID of the feed requested for deletion
	requestedID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Search the ID in our data
	index := -1
	for i, v := range myRSSReader.Feeds {
		if v.ID == requestedID {
			index = i
			break
		}
	}

	// If found
	if index != -1 {
		// Remove it from our data
		myRSSReader.Feeds = append(myRSSReader.Feeds[:index], myRSSReader.Feeds[index+1:]...)
		// Send back the new HTML feed
		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(w, "index", myRSSReader.Feeds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		// If not found, send back an error
		http.Error(w, "Item with requested ID not found", http.StatusBadRequest)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// parse and validate the url
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

	// check if duplicate
	for _, feed := range myRSSReader.Feeds {
		if feed.Link == urlString {
			err = tmpl.ExecuteTemplate(w, "error", struct{ Error string }{Error: "This feed has already been added"})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// clear the template Error
			return
		}
	}

	// fetch new feed from url
	newFeed := fetchFeed(urlString)

	// attach an ID and the link to our new value
	newFeed.ID = uuid.New()
	newFeed.Link = urlString

	// update state of truth (attach ID & link to newFeed)
	myRSSReader.Feeds = append(myRSSReader.Feeds, newFeed)

	// send back HTML
	err = tmpl.ExecuteTemplate(w, "feed", newFeed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type RSSReader struct {
	Feeds []RSSFeed
}

var myRSSReader = RSSReader{}

type ConnectionOption int

const (
	Polling ConnectionOption = iota
	SSE
	Websocket
)

func main() {
	router := http.NewServeMux()
	router.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("POST /addFeed", handleAdd)
	router.HandleFunc("DELETE /deleteFeed/{id}", handleDelete)
	router.HandleFunc("PATCH /handlePolling", handlePolling)
	router.HandleFunc("GET /{$}", handleHome)
	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	fmt.Println("Server listening...")
	server.ListenAndServe()
}
