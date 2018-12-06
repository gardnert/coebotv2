package main

import (
	"database/sql"
	"fmt"
	"regexp"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
)

func ParseAutoReply(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	channel := line.Args[0]
	rows, err := db.Query("SELECT `trigger`, `response` FROM autoreplies WHERE channel_ID=?", channelID)
	check(err)

	for rows.Next() {
		defer rows.Close()
		var trigger, response string
		rows.Scan(&trigger, &response)
		matched, err := regexp.MatchString(trigger, line.Text())
		if err != nil {
			check(err)
		}
		if matched {
			Send(conn, line, db, channelID, response)
			fmt.Println("Matched: " + channel + ": " + trigger + " " + response)
			break
		}
	}
}
