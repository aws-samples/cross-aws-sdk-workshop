package main

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}
type Item struct {
	Title         string    `xml:"title"`
	Description   string    `xml:"description"`
	Guid          string    `xml:"guid"`
	Enclosure     Enclosure `xml:"enclosure"`
	PublishedDate string    `xml:"pubDate"`
}
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}
