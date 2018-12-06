package main

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//check if a user is a regular or higher in a channel
func IsRegular(channelID string, db *sql.DB, tags map[string]string) bool {
	var (
		userLevel          string
		subscriberRegulars string
	)
	subscriber := strings.Contains(tags["badges"], "subscriber/") || strings.Contains(tags["badges"], "vip/1")
	user_ID := tags["user-id"]
	err := db.QueryRow("SELECT userlevel FROM users WHERE channel_ID=? AND user_ID=?", channelID, user_ID).Scan(&userLevel)
	if err != nil {
		userLevel = "not found"
	}
	err = db.QueryRow("SELECT subscriberRegulars FROM channel WHERE channel_ID=?", channelID).Scan(&subscriberRegulars)
	check(err)
	return userLevel == "regular" || (boolCheck(subscriberRegulars) && subscriber) || IsModerator(channelID, db, tags)
}

//check if a user is ignored in a channel
func IsIgnored(channelID string, db *sql.DB, tags map[string]string) bool {
	var userLevel string
	user_ID := tags["user-id"]
	err := db.QueryRow("SELECT userlevel FROM users WHERE channel_ID=? AND user_ID=?", channelID, user_ID).Scan(&userLevel)
	if err != nil {
		userLevel = "not found"
	}
	return userLevel == "ignored"

}

//check if a user is an owner in a channel
func IsOwner(channelID string, db *sql.DB, tags map[string]string) bool {
	var userLevel string
	user_ID := tags["user-id"]
	badges := tags["badges"]
	err := db.QueryRow("SELECT userlevel FROM users WHERE channel_ID=? AND user_ID=?", channelID, user_ID).Scan(&userLevel)
	if err != nil {
		userLevel = "not found"
	}
	return userLevel == "owner" || strings.Contains(badges, "broadcaster/1")
}

//check if a user is a moderator or higher in a channel
func IsModerator(channelID string, db *sql.DB, tags map[string]string) bool {
	var userLevel string
	mod := tags["mod"]
	user_ID := tags["user-id"]
	err := db.QueryRow("SELECT userlevel FROM users WHERE channel_ID=? AND user_ID=?", channelID, user_ID).Scan(&userLevel)
	if err != nil {
		userLevel = "not found"
	}
	return userLevel == "moderator" || boolCheck(mod) || IsOwner(channelID, db, tags)
}

func UserLevel(channelID string, db *sql.DB, tags map[string]string) int {

	accessLevel := 0

	if IsRegular(channelID, db, tags) {
		accessLevel = 1
	}
	if IsModerator(channelID, db, tags) {
		accessLevel = 2
	}
	if IsOwner(channelID, db, tags) {
		accessLevel = 3
	}

	return accessLevel
}
