package model

import "html/template"

type Item struct {
	Id           string `json:"id"`
	Title        string `json:"title"`
	Image        string `json:"image"`
	Link         string `json:"link"`
	CreationTime string `json:"creation_time"`
	Text         string `json:"text"`
	DomainTitle  string `json:"domain_title"`
}

type More struct {
	Link string `json:"link"`
}
type Channel struct {
	Items []Item `json:"items"`
	More  More   `json:"more"`
}

type ChannelData struct {
	Items []Item
}

type PageData struct {
	Detail template.HTML
}

type ChannelViewData struct {
	Items      []Item
	ChannelUrl string
}
