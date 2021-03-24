package model

import "html/template"

type ZenItem struct {
	Title        string `json:"title"`
	Image        string `json:"image"`
	Link         string `json:"link"`
	CreationTime string `json:"creation_time"`
}

type ZenChannel struct {
	Items []ZenItem `json:"items"`
}

type ChannelData struct {
	Items []ZenItem
}

type PageData struct {
	Detail template.HTML
}
