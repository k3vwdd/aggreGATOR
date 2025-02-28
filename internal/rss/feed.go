package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedUrl, nil)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error creating request")
    }

    req.Header.Set("User-Agent", "gator")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error performing request")
    }

    defer resp.Body.Close()
    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error reading data")
    }

    var rssData RSSFeed
    err = xml.Unmarshal(data, &rssData)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error decoding xml: %w", err)
    }

    // clean un-escaped html
    //    Outside the loop:
    //        Unescapes and updates the channel's Title
    //        Unescapes and updates the channel's Description
    //    Inside the loop for each item:
    //        Unescapes and updates THAT item's Title (Item[i].Title)
    //        Unescapes and updates THAT item's Description (Item[i].Description)
    //        Each field gets properly unescaped and saved back to its correct location.
    //        The channel fields stay unchanged while each item's fields get updated.

    //rssData.Channel.Title = html.UnescapeString(rssData.Channel.Title)
    //rssData.Channel.Description = html.UnescapeString(rssData.Channel.Description)
    //for i := range rssData.Channel.Item {
    //    rssData.Channel.Item[i].Title = html.UnescapeString(rssData.Channel.Item[i].Title)
    //    rssData.Channel.Item[i].Description = html.UnescapeString(rssData.Channel.Item[i].Description)
    //}
    rssData.Channel.Title = html.UnescapeString(rssData.Channel.Title)
    rssData.Channel.Description = html.UnescapeString(rssData.Channel.Description)
    for i, item := range rssData.Channel.Item {
        item.Title = html.UnescapeString(item.Title)
        item.Description = html.UnescapeString(item.Description)
        rssData.Channel.Item[i] = item
    }

    return &rssData, nil
}
