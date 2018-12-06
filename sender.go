package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
)

func SendCommand(conn *irc.Conn, target, message string) {
	conn.Privmsg(target, message)
}
func Send(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID, message string) {
	channel := line.Args[0]
	messageSlice := strings.Split(line.Text(), " ")
	fused := strings.Join(messageSlice[1:], " ")
	parameters := strings.Split(fused, ";")
	var bullet string
	err := db.QueryRow("SELECT bullet FROM channel WHERE channel_ID=?", channelID).Scan(&bullet)
	check(err)

	if strings.Contains(message, "(_ONLINE_CHECK_)") {
		if channelIsLive(channelID) {
			message = strings.Replace(message, "(_ONLINE_CHECK_)", "", -1)
		} else {
			message = ""
		}

	}
	if strings.Contains(message, "(_SUBMODE_ON_)") {
		SendCommand(conn, channel, ".subscribers")
		message = strings.Replace(message, "(_SUBMODE_ON_)", "", -1)
	}
	if strings.Contains(message, "(_SUBMODE_OFF_)") {
		SendCommand(conn, channel, ".subscribersoff")
		message = strings.Replace(message, "(_SUBMODE_OFF_)", "", -1)
	}
	if strings.Contains(message, "(_VIEWERS_)") {
		message = strings.Replace(message, "(_VIEWERS_)", twitchViewers(channelID), -1)
	}
	if strings.Contains(message, "(_GAME_)") {
		message = strings.Replace(message, "(_GAME_)", twitchGame(channel[1:]), -1)
	}
	if strings.Contains(message, "(_STATUS_)") {
		message = strings.Replace(message, "(_STATUS_)", twitchStatus(channel[1:]), -1)
	}
	//TODO make this work for more than the exact number of parameters

	for index, _ := range parameters {
		rep := len(parameters)
		if rep > 1 && strings.Contains(message, "(_PARAMETER_)") {
			message = strings.Replace(message, "(_PARAMETER_)", parameters[index], 1)
			break
		} else if strings.Contains(message, "(_PARAMETER_)") {
			message = strings.Replace(message, "(_PARAMETER_)", parameters[index], -1)
			break
		}
		if rep > 1 && strings.Contains(message, "(_PARAMETER_CAPS_)") {
			message = strings.Replace(message, "(_PARAMETER_CAPS_)", strings.ToUpper(parameters[index]), 1)
			break
		} else if strings.Contains(message, "(_PARAMETER_CAPS_)") {
			message = strings.Replace(message, "(_PARAMETER_CAPS_)", strings.ToUpper(parameters[index]), -1)
			break
		}
	}

	if strings.Contains(message, "(_USER_)") {
		message = strings.Replace(message, "(_USER_)", line.Nick, -1)
	}
	if strings.Contains(message, "(_CHANNEL_URL_)") {
		message = strings.Replace(message, "(_CHANNEL_URL_)", "http://twitch.tv/"+channel[1:], -1)
	}
	if strings.Contains(message, "(_QUOTE_)") {
		var count int
		err := db.QueryRow("SELECT COUNT(quote) FROM quotes WHERE channel_ID=?", channelID).Scan(&count)
		check(err)

		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		randIndex := r1.Intn(count)
		fmt.Println(randIndex)
		var quote string
		err = db.QueryRow("SELECT quote FROM quotes WHERE channel_ID=? ORDER BY created_at LIMIT 1 OFFSET ?", channelID, randIndex).Scan(&quote)
		message = strings.Replace(message, "(_QUOTE_)", quote, -1)
	}

	if strings.Contains(message, "(_LIST_") && strings.Contains(message, "_RANDOM_)") {
		listStart := strings.Index(message, "(_LIST_") + 7
		listEnd := strings.Index(message, "_RANDOM_)")
		listName := strings.ToLower(message[listStart:listEnd])
		fmt.Println(listName)
		var count int
		err = db.QueryRow("SELECT COUNT(item) FROM list_items WHERE channel_ID=? AND list_name=?", channelID, listName).Scan(&count)
		if count > 0 {
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)
			randIndex := r1.Intn(count)
			var listItem string
			err = db.QueryRow("SELECT item FROM list_items WHERE channel_ID=? AND list_name=? ORDER BY created_at LIMIT 1 OFFSET ?", channelID, listName, randIndex).Scan(&listItem)
			if err == nil {
				message = strings.Replace(message, "(_LIST_"+listName+"_RANDOM_)", listItem, -1)
			}
		}

	}
	if message != "" {
		fmt.Println("SEND > " + channel + ": " + bullet + " " + message)
		conn.Privmsg(channel, bullet+" "+message)

	}
}
