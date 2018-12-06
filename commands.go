package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"
)

func ParseCommand(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) bool {
	msgs := strings.Split(line.Text(), " ")
	first := msgs[0][1:]
	userLevel := UserLevel(channelID, db, line.Tags)
	fmt.Println(userLevel)
	switch strings.ToLower(first) {
	case "command":
		if len(msgs) < 3 || userLevel < 2 {
			break
		}
		ModifyCommand(conn, line, db, channelID)
		return true

	case "alias":
		if len(msgs) < 3 || userLevel < 2 {
			break
		}
		ModifyAlias(conn, line, db, channelID)
		return true
	case "autoreply":
		if len(msgs) < 3 || userLevel < 2 {
			break
		}
		ModifyAutoreply(conn, line, db, channelID)
		return true
	case "filter":
		if len(msgs) < 2 || userLevel < 2 {
			break
		}
		ModifyFilter(conn, line, db, channelID)
		return true
	case "set":
		if len(msgs) < 3 || userLevel < 2 {
			break
		}
		ModifySetting(conn, line, db, channelID)
		return true
	default:
		//check if it's a custom command
		fmt.Println("Matched command: " + first)
		var value string
		var restriction int
		err := db.QueryRow("SELECT value,restriction FROM commands WHERE channel_ID=? AND `key`=? AND enabled='Y'", channelID, first).Scan(&value, &restriction)
		if err != nil {
			//if it's not, see if it's an alias for a command
			err = db.QueryRow("SELECT value, restriction FROM commands WHERE channel_ID=? AND `key`=(SELECT `key` FROM aliases WHERE channel_ID=? AND alias=?)",
				channelID, channelID, first).Scan(&value, &restriction)

			if err != nil {
				//if it's not, see if it's a list we should pull from
				err = db.QueryRow("SELECT restriction FROM lists WHERE channel_ID=? AND `list_name`=?", channelID, first).Scan(&restriction)
				if err == nil {
					if len(msgs) > 1 {
						switch msgs[1] {
						//if we need to modify the list
						case "add", "create", "delete", "remove":
							if len(msgs) < 3 || userLevel < 2 {
								break
							}
							ModifyListItem(conn, line, db, channelID)
							return true
						default:
							if len(msgs) > 1 {
								if isInteger(msgs[1]) {
									index, _ := strconv.Atoi(msgs[1])
									db.QueryRow("SELECT item FROM list_items WHERE channel_ID=? AND list_name=? LIMIT ?,1", channelID, first, index-1).Scan(&value)

								} else {
									db.QueryRow("Select item FROM list_items WHERE channel_ID=? AND list_name=? AND item LIKE ? ORDER BY RAND() LIMIT 1", channelID, first, "%"+strings.Join(msgs[1:], " ")+"%").Scan(&value)

								}
							}
						}
					}

				}
			}
		}
		if userLevel >= restriction {
			Send(conn, line, db, channelID, value)
			return true
		}
	}
	return false
}
func ModifyListItem(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "add", "create":
		if len(msgs) < 3 {
			return
		}
		list := msgs[0][1:]
		item := strings.Join(msgs[2:], " ")
		biggestIndex := -1
		err := db.QueryRow("SELECT `index` FROM list_items WHERE channel_ID=? AND list_name=? ORDER BY `index` DESC LIMIT 1", channelID, list).Scan(&biggestIndex)
		if err != nil {
			check(err)
		}
		_, err = db.Exec("INSERT INTO list_items(channel_ID, `list_name`,item,`index`) VALUES(?,?,?,?)", channelID, list, item, biggestIndex+1)
		if err != nil {
			check(err)
			return
		}
		Send(conn, line, db, channelID, "List item added.")

	case "delete", "remove":
		if len(msgs) < 3 || !isInteger(msgs[2]) {
			return
		}
		list := msgs[0][1:]
		index, _ := strconv.Atoi(msgs[2])
		_, err := db.Exec("DELETE FROM list_items WHERE channel_ID=? AND list_name=? AND `index`=?", channelID, list, index-1)
		if err != nil {
			check(err)
			return
		}
		Send(conn, line, db, channelID, "List item deleted.")
	}
}
func ModifySetting(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "lastfm":
		_, err := db.Exec("UPDATE channel SET lastfm=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			check(err)
			Send(conn, line, db, channelID, "Error setting LastFM username.")
			return
		}
		Send(conn, line, db, channelID, "LastFM username set to: "+msgs[2]+".")
	case "steam":
		_, err := db.Exec("UPDATE channel SET steamID=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting Steam ID.")
			return
		}
		Send(conn, line, db, channelID, "Steam ID set to: "+msgs[2]+".")

	case "mode":
		if !isInteger(msgs[2]) {
			Send(conn, line, db, channelID, "Invalid mode setting.")
			return
		}
		_, err := db.Exec("UPDATE channel SET mode=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting channel mode.")
			return
		}
		Send(conn, line, db, channelID, "Channel mode set to: "+msgs[2]+".")
	case "commerciallength":
		length := msgs[2]
		if length != "30" && length != "60" && length != "90" && length != "120" && length != "150" && length != "180" {
			Send(conn, line, db, channelID, "Invalid commercial length. Valid lengths are 30, 60, 90, 120, 150, and 180.")
			return
		}
		_, err := db.Exec("UPDATE channel SET lastfm=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting commercial length.")
			return
		}
		Send(conn, line, db, channelID, "Commercial length set to: "+msgs[2]+".")

	case "prefix":
		if len(msgs[2]) > 1 {
			Send(conn, line, db, channelID, "Command prefix must be a single character.")
			return
		} else if msgs[2] == "." || msgs[2] == "/" {
			Send(conn, line, db, channelID, "Command prefix cannot be '.' or '/'.")
			return
		}
		_, err := db.Exec("UPDATE channel SET commandPrefix=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting command prefix.")
			return
		}
		Send(conn, line, db, channelID, "Command prefix set to: '"+msgs[2]+"'.")

	case "bullet":
		_, err := db.Exec("UPDATE channel SET bullet=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting bullet.")
			return
		}
		Send(conn, line, db, channelID, "Bullet set to: "+msgs[2]+".")

	case "subscriberregulars":
		if strings.ToLower(msgs[2]) == "on" || strings.ToLower(msgs[2]) == "enable" {
			_, err := db.Exec("UPDATE channel SET subscriberRegulars=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Subscribers will be treated as regulars.")

		} else if strings.ToLower(msgs[2]) == "off" || strings.ToLower(msgs[2]) == "disable" {
			_, err := db.Exec("UPDATE channel SET subscriberRegulars=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Subscribers will not be treated as regulars.")

		}
	case "subscriberalerts":
		if strings.ToLower(msgs[2]) == "on" || strings.ToLower(msgs[2]) == "enable" {
			_, err := db.Exec("UPDATE channel SET subscriberAlert=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Subscribers will be announced.")
		} else if strings.ToLower(msgs[2]) == "off" || strings.ToLower(msgs[2]) == "disable" {
			_, err := db.Exec("UPDATE channel SET subscriberAlert=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Subscribers will not be announced.")

		} else if strings.ToLower(msgs[2]) == "message" {
			if len(msgs) < 4 {
				return
			}
			_, err := db.Exec("UPDATE channel SET subMessage=? WHERE channel_ID=?", strings.Join(msgs[3:], " "), channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Subscriber message has been set to: "+strings.Join(msgs[3:], " "))
			return
		}

	case "cooldown":
		if !isInteger(msgs[2]) {
			Send(conn, line, db, channelID, "Invalid cooldown time.")
			return
		}
		_, err := db.Exec("UPDATE channel SET cooldown=? WHERE channel_ID=?", msgs[2], channelID)
		if err != nil {
			fmt.Println(err.Error())
			Send(conn, line, db, channelID, "Error setting cooldown.")
			return
		}
		Send(conn, line, db, channelID, "Cooldown for custom commands set to: "+msgs[2]+".")

	case "urban":
		if strings.ToLower(msgs[2]) == "on" || strings.ToLower(msgs[2]) == "enable" {
			_, err := db.Exec("UPDATE channel SET urbanEnabled=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Urban dictionary word lookup is now enabled.")

		} else if strings.ToLower(msgs[2]) == "off" || strings.ToLower(msgs[2]) == "disable" {
			_, err := db.Exec("UPDATE channel SET urbanEnabled=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Urban dictionary word lookup is now disabled.")

		}
	}

}

func ModifyAutoreply(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "add", "create":
		if len(msgs) < 4 {
			return
		}
		trigger := strings.Replace(msgs[2], "_", " ", -1)
		response := strings.Join(msgs[3:], " ")
		rebuiltTrigger := ""
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(trigger, "*"), "*"), "*")

		if !strings.HasPrefix(trigger, "REGEX:") {
			if strings.HasPrefix(trigger, "*") {
				rebuiltTrigger += ".*"
			}
			for i, part := range parts {
				if len(part) < 1 {
					continue
				}
				rebuiltTrigger += regexp.QuoteMeta(part)
				if i < len(parts)-1 {
					rebuiltTrigger += ".*"
				}
			}
			if strings.HasSuffix(trigger, "*") {
				rebuiltTrigger += ".*"
			}
		} else {
			rebuiltTrigger = strings.Replace(trigger, "REGEX:", "", -1)
		}
		biggestIndex := -1
		err := db.QueryRow("SELECT `index` FROM autoreplies WHERE channel_ID=? ORDER BY `index` DESC LIMIT 1", channelID).Scan(&biggestIndex)
		if err != nil {
			check(err)

		}
		_, err = db.Exec("INSERT INTO autoreplies(channel_ID, `trigger`,response,`index`) VALUES(?,?,?,?)", channelID, rebuiltTrigger, response, biggestIndex+1)
		if err != nil {
			check(err)
			return
		}
		Send(conn, line, db, channelID, "Autoreply added.")
		return

	case "delete", "remove":
		if len(msgs) < 3 || !isInteger(msgs[2]) {
			return
		}
		index, _ := strconv.Atoi(msgs[2])
		_, err := db.Exec("DELETE FROM autoreplies WHERE channel_ID=? AND `index`=?", channelID, index-1)
		if err != nil {
			check(err)
			return
		}
		Send(conn, line, db, channelID, "Autoreply deleted.")

	}
}
func ModifyFilter(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "on", "enable":
		_, err := db.Exec("UPDATE channel SET useFilters=? WHERE channel_ID=?", "Y", channelID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		Send(conn, line, db, channelID, "Filters enabled.")

	case "off", "disable":
		_, err := db.Exec("UPDATE channel SET useFilters=? WHERE channel_ID=?", "N", channelID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		Send(conn, line, db, channelID, "Filters disabled.")
	case "enablewarnings":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE channel SET enableWarnings=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Warnings enabled.")

		case "off", "disable":
			_, err := db.Exec("UPDATE channel SET enableWarnings=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Warnings disabled.")
		}
	case "displaywarnings":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE channel SET signKicks=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Timeouts from filters will now be displayed.")

		case "off", "disable":
			_, err := db.Exec("UPDATE channel SET signKicks=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Timeouts from filters will now be silent.")
		}
	case "timeoutduration":
		if len(msgs) < 3 {
			return
		}
		if !isInteger(msgs[2]) {
			Send(conn, line, db, channelID, msgs[2]+" is not a valid timeout duration.")
			return
		} else {
			_, err := db.Exec("UPDATE channel SET timeoutDuration=? WHERE channel_ID=?", msgs[2], channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Timeout duration has been set to "+msgs[2]+"seconds.")
			return
		}
	case "messagelength":
		if len(msgs) < 3 {
			return
		}
		if !isInteger(msgs[2]) {
			Send(conn, line, db, channelID, msgs[2]+" is not a valid max message length.")
			return
		} else {
			_, err := db.Exec("UPDATE filters SET filterMaxLength=? WHERE channel_ID=?", msgs[2], channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Maximum message length has been set to "+msgs[2]+" characters.")
			return
		}

	case "links":
		if len(msgs) < 3 {
			return
		}
		if strings.ToLower(msgs[2]) == "on" {
			_, err := db.Exec("UPDATE filters SET filterLinks=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Link filter enabled.")
		} else if strings.ToLower(msgs[2]) == "off" {
			_, err := db.Exec("UPDATE filters SET filterLinks=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Link filter disabled.")
		}
	case "pd", "permitteddomain":
		if len(msgs) < 4 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "add":

			_, err := db.Exec("INSERT INTO permitteddomains(channel_ID, domain) VALUES(?,?)", channelID, msgs[3])
			if err != nil {
				fmt.Println(err.Error())
				Send(conn, line, db, channelID, "Unable to add permitted domain.")
				return
			}
			Send(conn, line, db, channelID, "Added permitted domain: "+msgs[3]+".")
		case "delete", "remove":

			var pdExists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM permitteddomains WHERE channel_ID=? AND domain=?)",
				channelID, msgs[3]).Scan(&pdExists)
			check(err)
			if !pdExists {
				Send(conn, line, db, channelID, msgs[3]+" doesn't exist.")
				return
			} else {
				_, err := db.Exec("DELETE FROM permitteddomains WHERE channel_ID=? AND domain=?", channelID, msgs[3])
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Unable to delete permitted domain.")
					return
				} else {
					Send(conn, line, db, channelID, msgs[3]+" removed as a permitted domain.")
					return
				}
			}

		case "list", "show":
			sendStr := "Permitted domains: "
			rows, err := db.Query("SELECT domain from permitteddomains WHERE channel_ID=?", channelID)
			check(err)
			for rows.Next() {
				var domain string
				err := rows.Scan(&domain)
				check(err)
				sendStr += domain + ", "
			}
			sendStr = sendStr[:len(sendStr)-2] + "."
			rows.Close()
			Send(conn, line, db, channelID, sendStr)
		}
	case "caps", "captials":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE filters SET filterCaps=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Caps filter enabled.")

		case "off", "disable":
			_, err := db.Exec("UPDATE filters SET filterCaps=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Caps filter disabled.")
		case "percent":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid caps percent.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterCapsPercent=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting caps percent.")
					return
				}
				Send(conn, line, db, channelID, "Caps percent has been set to "+msgs[3]+"%.")
				return
			}
		case "mincaps":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid number for min caps.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterCapsMinCapitals=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting minimum number of caps.")
					return
				}
				Send(conn, line, db, channelID, "Minimum caps has been set to "+msgs[3]+".")
				return
			}
		case "minchars":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid number for minimum characters.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterCapsMinCharacters=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting minimum number of characters.")
					return
				}
				Send(conn, line, db, channelID, "Minimum characters has been set to "+msgs[3]+".")
				return
			}
		}
	case "banphrase":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE filters SET filterOffensive=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Banphrase filter enabled.")

		case "off", "disable":
			_, err := db.Exec("UPDATE filters SET filterOffensive=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Banphrase filter disabled.")

		case "list", "show":
			sendStr := "Banned phrases: "
			rows, err := db.Query("SELECT phrase from offensivewords WHERE channel_ID=?", channelID)
			check(err)
			for rows.Next() {
				var phrase string
				err := rows.Scan(&phrase)
				check(err)
				sendStr += phrase + ", "
			}
			sendStr = sendStr[:len(sendStr)-2] + "."
			rows.Close()
			Send(conn, line, db, channelID, sendStr)

		case "add", "new":
			if len(msgs) < 4 {
				return
			}
			_, err := db.Exec("INSERT INTO offensivewords(channel_ID, phrase) VALUES(?,?)", channelID, msgs[3])
			if err != nil {
				fmt.Println(err.Error())
				Send(conn, line, db, channelID, "Unable to add banned phrase.")
				return
			}
			Send(conn, line, db, channelID, "Added banned phrase: "+msgs[3]+".")
		case "delete", "remove":
			if len(msgs) < 4 {
				return
			}
			var phraseExists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM offensivewords WHERE channel_ID=? AND phrase=?)",
				channelID, msgs[3]).Scan(&phraseExists)
			check(err)
			if !phraseExists {
				Send(conn, line, db, channelID, "Banned phrase: "+msgs[3]+" doesn't exist.")
				return
			} else {
				_, err := db.Exec("DELETE FROM offensivewords WHERE channel_ID=? AND phrase=?", channelID, msgs[3])
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Unable to delete banned phrase.")
					return
				} else {
					Send(conn, line, db, channelID, msgs[3]+" removed as a banned phrase.")
					return
				}
			}
		}
	case "symbols":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE filters SET filterSymbols=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Symbols filter enabled.")

		case "off", "disable":
			_, err := db.Exec("UPDATE filters SET filterSymbols=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Symbols filter disabled.")

		case "min":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid number for min symbols.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterSymbolsMin=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting minimum number of symbols.")
					return
				}
				Send(conn, line, db, channelID, "Minimum symbols has been set to "+msgs[3]+".")
				return
			}
		case "percent":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid number for symbols percent.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterSymbolsPercent=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting symbols percent.")
					return
				}
				Send(conn, line, db, channelID, "Symbols percent has been set to "+msgs[3]+"%.")
				return
			}
		}
	case "emotes":
		if len(msgs) < 3 {
			return
		}
		sub := strings.ToLower(msgs[2])
		switch sub {
		case "on", "enable":
			_, err := db.Exec("UPDATE filters SET filterEmotes=? WHERE channel_ID=?", "Y", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Emote filter enabled.")

		case "off", "disable":
			_, err := db.Exec("UPDATE filters SET filterEmotes=? WHERE channel_ID=?", "N", channelID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Emote filter disabled.")

		case "min":
			if len(msgs) < 4 {
				return
			}
			if !isInteger(msgs[3]) {
				Send(conn, line, db, channelID, msgs[3]+" is not a valid number for max emotes.")
				return
			} else {
				_, err := db.Exec("UPDATE filters SET filterEmotesMax=? WHERE channel_ID=?", msgs[3], channelID)
				if err != nil {
					fmt.Println(err.Error())
					Send(conn, line, db, channelID, "Error setting maximum number of emotes.")
					return
				}
				Send(conn, line, db, channelID, "Maximum emotes has been set to "+msgs[3]+".")
				return
			}
		}
	}
}
func ModifyAlias(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {
	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "add", "create":
		if len(msgs) < 4 {
			return
		}
		alias := msgs[2]
		key := msgs[3]

		var commandExists, aliasExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM commands WHERE channel_ID=? AND `key`=?)", channelID, key).Scan(&commandExists)
		check(err)
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM aliases WHERE channel_ID=? AND alias=?)", channelID, alias).Scan(&aliasExists)
		check(err)
		if commandExists && !aliasExists {
			_, err = db.Exec("INSERT INTO aliases(channel_ID,`key`,alias) VALUES(?,?,?)",
				channelID, key, alias)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Alias "+alias+" added for command "+key+".")
		} else if !commandExists {
			Send(conn, line, db, channelID, "Command "+key+" doesn't exist.")
		} else if aliasExists {
			Send(conn, line, db, channelID, "Alias "+alias+" already exists.")
		}

	case "delete", "remove":
		if len(msgs) < 3 {
			return
		}
		alias := msgs[2]
		res, err := db.Exec("DELETE FROM aliases WHERE channel_ID=? AND alias=?", channelID, alias)
		check(err)
		if updated, _ := res.RowsAffected(); updated < 1 {
			Send(conn, line, db, channelID, "Alias "+alias+" doesn't exist.")
		} else {
			Send(conn, line, db, channelID, "Alias "+alias+" removed.")
		}

	}
}
func ModifyCommand(conn *irc.Conn, line *irc.Line, db *sql.DB, channelID string) {

	msgs := strings.Split(line.Text(), " ")

	subcommand := strings.ToLower(msgs[1])
	switch subcommand {
	case "add", "create":
		if len(msgs) < 4 {
			return
		}
		key := msgs[2]
		value := strings.Join(msgs[3:], " ")
		var commandExists, aliasExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM commands WHERE channel_ID=? AND `key`=?)", channelID, key).Scan(&commandExists)
		check(err)
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM aliases WHERE channel_ID=? AND alias=?)", channelID, key).Scan(&aliasExists)
		check(err)
		if !commandExists && !aliasExists {
			restriction := 1
			if strings.Contains(value, "(_BAN_)") || strings.Contains(value, "(_PURGE_)") || strings.Contains(value, "(_TIMEOUT_)") ||
				strings.Contains(value, "(_SUBMODE_ON_)") || strings.Contains(value, "(_SUBMODE_OFF_)") || strings.Contains(value, "(_VARS_)") {
				restriction = 2
			}
			_, err = db.Exec("INSERT INTO commands(channel_ID,editor,restriction,`key`,value) VALUES(?,?,?,?,?)",
				channelID, line.Nick, restriction, key, value)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			Send(conn, line, db, channelID, "Command "+key+" added.")
		} else if commandExists {
			Send(conn, line, db, channelID, "Command "+key+" already exists.")
		} else if aliasExists {
			Send(conn, line, db, channelID, "Command Alias "+key+" already exists.")
		}

	case "delete", "remove":
		if len(msgs) < 3 {
			return
		}
		key := msgs[2]
		res, err := db.Exec("DELETE FROM commands WHERE channel_ID=? AND `key`=?", channelID, key)
		check(err)
		if updated, _ := res.RowsAffected(); updated < 1 {
			Send(conn, line, db, channelID, "Command "+key+" doesn't exist.")
		} else {
			Send(conn, line, db, channelID, "Command "+key+" removed.")
		}

	case "edit", "update":
		if len(msgs) < 4 {
			return
		}
		key := msgs[2]
		value := strings.Join(msgs[3:], " ")
		res, err := db.Exec("UPDATE commands SET value=? WHERE channel_ID=? and `key`=?", value, channelID, key)
		check(err)
		if updated, _ := res.RowsAffected(); updated < 1 {
			Send(conn, line, db, channelID, "Command "+key+" doesn't exist.")
		} else {
			Send(conn, line, db, channelID, "Command "+key+" updated successfully.")
		}
	case "rename":
		if len(msgs) < 4 {
			return
		}
		key := msgs[2]
		newkey := msgs[2]
		var commandExists, aliasExists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM commands WHERE channel_ID=? AND `key`=?)", channelID, key).Scan(&commandExists)
		check(err)
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM aliases WHERE channel_ID=? AND alias=?)", channelID, key).Scan(&aliasExists)
		check(err)
		if commandExists && !aliasExists {
			_, err := db.Exec("UPDATE commands SET `key`=? WHERE channel_ID=? and `key`=?", newkey, channelID, key)
			check(err)
			Send(conn, line, db, channelID, "Command "+key+" renamed to "+newkey+".")
		} else if !commandExists {
			Send(conn, line, db, channelID, "Command "+key+" doesn't exist.")
		} else if aliasExists {
			Send(conn, line, db, channelID, "New command name "+newkey+" already exists as an alias.")
		}
	}

}
