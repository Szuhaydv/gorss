package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Item        []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, exampleServers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("li.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	err = tmpl.Execute(w, r.FormValue("inputField"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type RSSReader struct {
	Servers []string
	Items   []Item
}

var exampleServers = RSSReader{
	Servers: []string{"https://www.idk.com", "https://www.classfm.com"},
	Items:   []Item{},
}

func main() {
	baseUrl := "https://lorem-rss.herokuapp.com/feed?unit=second&interval=30"

	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", handlerFunc)
	router.HandleFunc("/addServer", handleAdd)
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
	var feed Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		fmt.Println("Failed to parse RSS feed")
	}
}
