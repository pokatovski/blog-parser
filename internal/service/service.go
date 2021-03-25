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

func ProcessChannel(ch string, isNamed bool) (model.ZenChannel, error) {
	var url string
	var zenCh model.ZenChannel
	var channelData model.ZenChannel
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

		err = json.Unmarshal(bodyBytes, &zenCh)
		if err != nil {
			return channelData, err
		}
		channelData = model.ZenChannel(model.ChannelData{Items: zenCh.Items})

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
