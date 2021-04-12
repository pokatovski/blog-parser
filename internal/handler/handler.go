package handler

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pokatovski/blog-parser/internal/model"
	"github.com/pokatovski/blog-parser/internal/service"
	"github.com/recoilme/clean"
)

var templates = template.Must(template.ParseGlob("web/templates/*"))

func Index(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Parse(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	url = html.EscapeString(url)
	if url == "" {
		err := errors.New("bad request: url is required")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	splittedUrl := strings.Split(url, "/")
	if splittedUrl[2] == "zen.yandex.ru" && splittedUrl[3] != "media" {
		ch, isNamed, err := service.GetChannel(splittedUrl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		chData, err := service.ProcessChannel(ch, isNamed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := model.ChannelViewData{Items: chData.Items, ChannelUrl: url}
		err = templates.ExecuteTemplate(w, "channel.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}
	result, err := clean.URI(url, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := model.PageData{Detail: template.HTML(result)}
	err = templates.ExecuteTemplate(w, "detail.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Rss(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	url = html.EscapeString(url)
	if url == "" {
		err := errors.New("bad request: url is required")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	splittedUrl := strings.Split(url, "/")
	ch, isNamed, err := service.GetChannel(splittedUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	now := time.Now()
	chData, err := service.ProcessChannel(ch, isNamed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	duration := time.Since(now).Seconds()
	fmt.Println("process channel duration", duration)
	rss, err := service.MakeRss(chData, url, r.Host)
	fmt.Println("end make rss")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(rss))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RssSingle(w http.ResponseWriter, r *http.Request) {
	xml, err := ioutil.ReadFile("rss.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(xml)
}
