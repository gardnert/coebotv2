package main

import (
	//"github.com/davecgh/go-spew/spew"
	"strconv"
)

type Stream struct {
	Game      string `json:"game"`
	Viewers   int    `json:"viewers"`
	CreatedAt string `json:"created_at"`
}
type Channel struct {
	Status string `json:"status"`
	Game   string `json:"game"`
}

type StreamChannel struct {
	MStream  Stream  `json:"stream"`
	MChannel Channel `json:"channel"`
}

func channelIsLive(channelID string) bool {
	mStreamChannel := StreamChannel{}
	getTwitchJson("https://api.twitch.tv/kraken/streams/"+channelID, &mStreamChannel)
	//spew.Dump(mStreamChannel)
	return mStreamChannel.MStream.CreatedAt != ""
}

func twitchGame(channelName string) string {
	mChannel := Channel{}
	getTwitchJson("https://api.twitch.tv/kraken/channels/"+channelName, &mChannel)
	return mChannel.Game
}

func twitchStatus(channelName string) string {
	mChannel := Channel{}
	getTwitchJson("https://api.twitch.tv/kraken/channels/"+channelName, &mChannel)
	return mChannel.Status
}

func twitchViewers(channelID string) string {
	mStreamChannel := StreamChannel{}
	getTwitchJson("https://api.twitch.tv/kraken/streams/"+channelID, &mStreamChannel)

	if mStreamChannel.MStream.CreatedAt != "" {
		return strconv.Itoa(mStreamChannel.MStream.Viewers)
	} else {
		return "0"
	}
}
