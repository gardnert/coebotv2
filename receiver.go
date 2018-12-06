package main

import (
	"database/sql"
	"fmt"

	//"github.com/davecgh/go-spew/spew"
	"strings"

	"github.com/davecgh/go-spew/spew"
	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
)

//create a new connection to twitch, sending the correct CAP requests for proper bot functionality
//also register the callbacks for specific IRC events
func NewReceiver(cfg *irc.Config, channel_name string, db *sql.DB) *irc.Conn {
	c := irc.Client(cfg)
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			conn.Cap("REQ", "twitch.tv/commands")
			conn.Cap("REQ", "twitch.tv/tags")
			conn.Join("#" + channel_name)
			fmt.Println("Joined #" + channel_name)
		})
	c.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			ParseMessage(conn, line, db)
		})
	c.HandleFunc(irc.NOTICE,
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Println(line.Nick + ": " + line.Raw)
		})
	c.Connect()
	return c
}

func ParseMessage(conn *irc.Conn, line *irc.Line, db *sql.DB) {
	//ignore messages from the bot
	//	if strings.EqualFold(line.Nick, conn.Me().Nick) {
	//		return
	//	}

	var (
		prefix         string
		shouldModerate string
		filtersEnabled string
		channelID      string
		cooldown       int
		mode           int
		lastfm         string
		parseYoutube   string
		urbanEnabled   string
		rollLevel      string
	)

	channel := line.Args[0]

	//retrieve several variables from the database for the channel
	err := db.QueryRow("SELECT commandPrefix, shouldModerate, useFilters, channel_ID, cooldown, mode, lastfm, parseYoutube, urbanEnabled,"+
		" rollLevel FROM channel WHERE channel_name=?",
		strings.TrimPrefix(channel, "#")).Scan(&prefix, &shouldModerate, &filtersEnabled, &channelID, &cooldown, &mode, &lastfm,
		&parseYoutube, &urbanEnabled, &rollLevel)
	check(err)
	actioned := false
	//if boolCheck(filtersEnabled) && boolCheck(shouldModerate) && !IsRegular(channelID, db, line.Tags) {
	if boolCheck(filtersEnabled) && boolCheck(shouldModerate) {
		actioned = ParseFilters(conn, line, db, channelID)
		//check filters before breaking for ignored users
	}

	if IsIgnored(channelID, db, line.Tags) {
		return
	}

	//check if it matches any commands
	if strings.HasPrefix(line.Text(), prefix) && !actioned {
		actioned = ParseCommand(conn, line, db, channelID)
	}
	//check if it matches any autoreplies
	if !actioned {
		ParseAutoReply(conn, line, db, channelID)
	}
	//TODO write the rest of the things

	//print the channel, sender, and message
	spew.Dump(line.Tags)
	fmt.Println(line.Args[0] + ": " + line.Nick + ": " + line.Text())

}
