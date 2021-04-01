package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pokatovski/blog-parser/internal/model"
)

const zenApi = "https://zen.yandex.ru/api/v3/launcher/"
const articleCount = 100
const articleOffset = 20

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
