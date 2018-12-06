package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
	xurls "github.com/mvdan/xurls"
)

func ParseFilters(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) bool {
	var filterLinks, filterOffensive, filterCaps, filterSymbols, filterEmotes, signKicks string
	var filterCapsPercent, filterCapsMinCaps, filterCapsMinChars, filterSymbolsMin, filterMaxLength, filterEmotesMax, filterSymbolsPercent, timeoutDuration int

	err := db.QueryRow("SELECT filterLinks, filterOffensive, filterCaps, filterSymbols, filterEmotes, filterCapsPercent, filterCapsMinCapitals, "+
		"filterCapsMinCharacters, filterSymbolsMin, filterMaxLength, filterEmotesMax, filterSymbolsPercent FROM filters WHERE channel_ID=?",
		channelID).Scan(&filterLinks, &filterOffensive, &filterCaps, &filterSymbols, &filterEmotes, &filterCapsPercent, &filterCapsMinCaps,
		&filterCapsMinChars, &filterSymbolsMin, &filterMaxLength, &filterEmotesMax, &filterSymbolsPercent)
	check(err)

	err = db.QueryRow("SELECT timeoutDuration,signKicks FROM channel WHERE channel_ID=?", channelID).Scan(&timeoutDuration, &signKicks)
	check(err)
	if boolCheck(filterLinks) {
		urls := xurls.Relaxed().FindAllString(line.Text(), -1)

		allowedLinks := true
		if urls != nil {
			rows, err := db.Query("SELECT domain FROM permitteddomains WHERE channel_ID=?", channelID)
			check(err)
			for _, url := range urls {
				if !checkAllowedURL(url, rows) {
					allowedLinks = false
				}
			}
		}
		if !allowedLinks {

			SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
			if boolCheck(signKicks) {
				Send(conn, line, db, channelID, line.Nick+", please ask a moderator before posting links. - Timeout")
				return true
			}
		}

	}
	if len(line.Text()) > filterMaxLength {
		fmt.Println(filterMaxLength)
		SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
		if boolCheck(signKicks) {
			Send(conn, line, db, channelID, line.Nick+", please limit the length of your messages. - Timeout")
			return true
		}
	}
	if boolCheck(filterEmotes) {
		emotes := line.Tags["emotes"]
		if len(emotes) > 0 {
			numEmotes := len(strings.Split(emotes, ","))
			if numEmotes > filterEmotesMax {
				SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
				if boolCheck(signKicks) {
					Send(conn, line, db, channelID, line.Nick+", please limit the number of emotes you're sending. - Timeout")
					return true
				}
			}
		}
	}
	if boolCheck(filterOffensive) {
		rows, err := db.Query("SELECT phrase FROM offensivewords WHERE channel_ID=?", channelID)
		check(err)
		var phrase string
		for rows.Next() {
			defer rows.Close()
			rows.Scan(&phrase)
			matched, err := regexp.MatchString(phrase, line.Text())
			check(err)
			if matched {
				SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
				if boolCheck(signKicks) {
					Send(conn, line, db, channelID, line.Nick+", disallowed word or phrase. - Timeout")
					return true
				}
				break
			}
		}
	}

	if boolCheck(filterCaps) {
		length := len(strings.Replace(line.Text(), " ", "", -1))
		numCaps := countCaps(line.Text())
		percentCaps := (numCaps / float32(length)) * 100
		if int(percentCaps) > filterCapsPercent && filterCapsMinChars <= length {
			SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
			if boolCheck(signKicks) {
				Send(conn, line, db, channelID, line.Nick+", please limit the number of caps in your messages. - Timeout")
				return true
			}
		}
	}
	if boolCheck(filterSymbols) {
		length := len(strings.Replace(line.Text(), " ", "", -1))
		numSymbols := countSymbols(strings.Replace(line.Text(), " ", "", -1))
		percentSymbols := (numSymbols / float32(length)) * 100
		if int(percentSymbols) > filterSymbolsPercent && float32(filterSymbolsMin) < numSymbols {
			SendCommand(conn, line.Args[0], ".timeout "+line.Nick+" "+strconv.Itoa(timeoutDuration))
			if boolCheck(signKicks) {
				Send(conn, line, db, channelID, line.Nick+", please limit the number of symbols in your messages. - Timeout")
				return true
			}
		}

	}
	return false
}
func checkAllowedURL(url string, rows *sql.Rows) bool {
	permitted := false
	var domain string
	for rows.Next() {
		defer rows.Close()
		rows.Scan(&domain)
		matched, err := regexp.MatchString(domain, url)
		check(err)
		if matched {
			permitted = true
		}
	}
	return permitted
}

func countCaps(line string) float32 {
	var numCaps float32 = 0
	for _, runeValue := range line {
		if unicode.IsUpper(runeValue) {
			numCaps++
		}
	}

	return numCaps
}

func countSymbols(line string) float32 {
	var numSymbols float32 = 0
	for _, runeValue := range line {
		//if the rune isn't uppercase, lowercase, or a 1-9
		if (runeValue < 97 || runeValue > 122) && (runeValue < 65 || runeValue > 90) && (runeValue < 48 || runeValue > 57) {
			numSymbols++
		}
	}
	return numSymbols
}
