package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/pokatovski/blog-parser/internal/model"
	"github.com/recoilme/clean"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const zenApi = "https://zen.yandex.ru/api/v3/launcher/"
const articleCount = 100
const articleOffset = 20
const maxGoroutines = 3

var feedItems = make(map[string]*feeds.Item)

func ProcessChannel(ch string, isNamed bool) (model.ChannelData, error) {
	var url string
	var Channel model.Channel
	var channelData model.ChannelData
	if isNamed {
		url = fmt.Sprintf("%smore?channel_name=%s", zenApi, ch)
	} else {
		url = fmt.Sprintf("%smore?channel_id=%s", zenApi, ch)
	}

	resp, err := http.Get(url)
	if err != nil {
		return channelData, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return channelData, err
		}
		err = json.Unmarshal(bodyBytes, &Channel)
		if err != nil {
			return channelData, err
		}
		moreLink := Channel.More.Link
		if len(Channel.Items) == articleOffset {
			var tmpChanItems, mergedItems []model.Item
			for i := 0; i < articleCount/articleOffset-1; i++ {
				tmpMoreChan, err := loadMore(moreLink)
				if err != nil {
					return channelData, err
				}
				tmpChanItems = append(tmpChanItems, tmpMoreChan.Items...)
				moreLink = tmpMoreChan.More.Link
				if moreLink == "" {
					break
				}
			}
			mergedItems = append(Channel.Items, tmpChanItems...)
			channelData = model.ChannelData{Items: mergedItems}
		} else {
			channelData = model.ChannelData{Items: Channel.Items}
		}
	} else {
		err := errors.New("status code from blog:" + strconv.Itoa(resp.StatusCode))
		return channelData, err
	}

	return channelData, nil
}

func GetChannel(splitted []string) (string, bool, error) {
	//zen channels has two types: named and unnamed(https://zen.yandex.ru/channel_name or https://zen.yandex.ru/id/1)
	var isNamed bool
	var ch string
	if len(splitted) == 4 {
		isNamed = true
		ch = splitted[3]
	} else if len(splitted) == 5 {
		ch = splitted[4]
	} else {
		err := errors.New("bad path")
		return ch, isNamed, err
	}
	return ch, isNamed, nil
}

func loadMore(url string) (model.Channel, error) {
	var channel model.Channel
	resp, err := http.Get(url)
	if err != nil {
		return channel, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return channel, err
		}
		err = json.Unmarshal(bodyBytes, &channel)
		if err != nil {
			return channel, err
		}
		moreLink := channel.More.Link
		channel = model.Channel{Items: channel.Items, More: model.More{Link: moreLink}}
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return channel, err
		}
		emptyMore := "{}"
		if string(bodyBytes) == emptyMore {
			return channel, nil
		}
		err = errors.New("status code from more blog:" + strconv.Itoa(resp.StatusCode))
		return channel, err
	}
	return channel, nil
}

func MakeRss(data model.ChannelData, chUrl, host string) (string, error) {
	start := time.Now()
	created := time.Now()
	var sortedFeed []*feeds.Item
	feedLink := fmt.Sprintf("https://%s/parse?url=%s", host, chUrl)
	title := data.Items[0].DomainTitle
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: feedLink},
		Description: title,
		Created:     created,
	}

	maxGoroutines := maxGoroutines
	jobs := make(chan struct{}, maxGoroutines)
	wg := sync.WaitGroup{}
	for _, item := range data.Items {
		wg.Add(1)
		jobs <- struct{}{}
		go process(item, jobs, host, &wg)
		durationClear := time.Since(start).Seconds()
		fmt.Println("process clear duration", durationClear)
	}
	wg.Wait()
	close(jobs)
	duration := time.Since(start).Seconds()
	fmt.Println("process items duration", duration)
	for _, dataItem := range data.Items {
		if v, ok := feedItems[dataItem.Id]; ok {
			sortedFeed = append(sortedFeed, v)
		}
	}

	feed.Items = sortedFeed

	rss, err := feed.ToRss()
	if err != nil {
		return "", nil
	}

	return rss, nil
}

func makeLink(url, host string) (string, error) {
	splitUrl := strings.Split(url, "?")
	if len(splitUrl) == 1 {
		err := errors.New(fmt.Sprintf("bad url for splitting:%s", url))
		return "", err
	}
	resLink := fmt.Sprintf("https://%s/parse?url=%s", host, splitUrl[0])
	return resLink, nil
}

func process(item model.Item, jobs <-chan struct{}, host string, wg *sync.WaitGroup) {
	result, err := clean.URI(item.Link, false)
	defer wg.Done()
	if err != nil {
		fmt.Println("failed for get response from url: ", item.Link)
		fmt.Println("failed for get response from url err: ", err.Error())
		<-jobs
		return
	}
	emptyImg := `<img src=""/>`
	result = strings.ReplaceAll(result, emptyImg, "")
	link, err := makeLink(item.Link, host)
	if err != nil {
		fmt.Println("err make link ", err)
		<-jobs
		return
	}
	created := time.Now()
	newItem := &feeds.Item{
		Title:       item.Title,
		Link:        &feeds.Link{Href: link},
		Description: item.Text,
		Content:     html.UnescapeString(result),
		Created:     created,
	}

	if item.Image != "" {
		newItem.Enclosure = &feeds.Enclosure{Url: item.Image, Type: "image/jpeg", Length: "1"}
	}

	feedItems[item.Id] = newItem

	<-jobs
}
