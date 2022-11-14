package main

import "time"

// Category is a category of an Atom entry.
type Category struct {
	Term string `xml:"term,attr"`
}

// Entry is a single entry in an Atom feed.
type Entry struct {
	Title     string        `xml:"title"`
	Summary   string        `xml:"summary,omitempty"`
	Content   *EntryContent `xml:"content"`
	Published time.Time     `xml:"published"`
	Updated   time.Time     `xml:"updated"`
	Link      *Link         `xml:"link"`
	ID        string        `xml:"id"`

	AuthorName string `xml:"author>name,omitempty"`
	AuthorURI  string `xml:"author>uri,omitempty"`

	Categories []*Category `xml:"category"`
}

// EntryContent is a simple helper class that allows us to wrap an entry's
// content in an XML CDATA tag.
type EntryContent struct {
	Content string `xml:",cdata"`
	Type    string `xml:"type,attr,omitempty"`
}

// Feed represents an Atom feed that with be marshaled to XML.
//
// Note that XMLName is a Golang XML "magic" attribute.
type Feed struct {
	XMLName struct{} `xml:"feed"`

	XMLLang string `xml:"xml:lang,attr"`
	XMLNS   string `xml:"xmlns,attr"`

	Title   string    `xml:"title"`
	ID      string    `xml:"id"`
	Updated time.Time `xml:"updated"`

	Links   []*Link  `xml:"link"`
	Entries []*Entry `xml:"entry"`
}

// Link is a link embedded in the header of an Atom feed.
type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Href string `xml:"href,attr"`
}

func sortEntriesDesc(a, b *Entry) bool {
	return b.Published.Before(a.Published)
}
