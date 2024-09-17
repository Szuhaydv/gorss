package main

import (
	"encoding/xml"
	"fmt"
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

func printFeed(feed Feed) {
    fmt.Println("Title:", feed.Channel.Title)
    fmt.Println("Description:", feed.Channel.Description)
    fmt.Println()
    fmt.Println("Items:")
    fmt.Println("----------------------")
    for i, item := range feed.Channel.Item {
        fmt.Printf("Item %d:\n", i+1)
        fmt.Println("  Title:", item.Title)
        fmt.Println("  Description:", item.Description)
        fmt.Println("  Link:", item.Link)
        fmt.Println()
    }
}

func main() {
  baseUrl := "https://lorem-rss.herokuapp.com/feed?unit=second&interval=30"

  // GET the RSS feed (in this case Lorem RSS)
  resp, err := http.Get(baseUrl); 
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

  printFeed(feed)
  
}
