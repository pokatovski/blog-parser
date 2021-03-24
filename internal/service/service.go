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
