package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	//spew is a helpful debugging tool
	"io/ioutil"
	"net/http"
	"strings"

	//"github.com/davecgh/go-spew/spew"

	_ "github.com/go-sql-driver/mysql"
)

//configuration file for the converter
type Config struct {
	Channels string `json:"channelList"`
	DBHost   string `json:"host"`
	DBName   string `json:"databaseName"`
	DBUser   string `json:"user"`
	DBPass   string `json:"password"`
}

//used for converting between twitch name and twitch channel ID
type Users struct {
	Members []User `json:"users"`
}

type User struct {
	Id string `json:"_id"`
}

type Autoreply struct {
	Response string `json:"response"`
	Trigger  string `json:"trigger"`
}

type Command struct {
	Editor      string `json:"editor"`
	Count       int    `json:"count"`
	Value       string `json:"value"`
	Key         string `json:"key"`
	Restriction int    `json:"restriction"`
}

type RepeatedCommand struct {
	MessageDifference int    `json:"messageDifference"`
	Name              string `json:"name"`
	Active            bool   `json:"active"`
	Delay             int    `json:"delay"`
}

type ScheduledCommand struct {
	Pattern           string `json:"pattern"`
	MessageDifference int    `json:"messageDifference"`
	Active            bool   `json:"active"`
	Name              string `json:"name"`
}

type List struct {
	Restriction int      `json:"restriction"`
	Items       []string `json:"items"`
}

type Quote struct {
	Timestamp int64  `json:"timestamp"`
	Editor    string `json:"editor"`
	Text      string `json:"quote"`
}

//overall struct for the channel's config
type Channel struct {
	Commands                []Command          `json:"commands"`
	CommandPrefix           string             `json:"commandPrefix"`
	Lists                   map[string]List    `json:"lists"`
	Quotes                  []Quote            `json:"quotes"`
	RollLevel               string             `json:"rollLevel"`
	FilterLinks             bool               `json:"filterLinks"`
	FilterOffensive         bool               `json:"filterOffensive"`
	ChannelID               string             `json:"channelID"`
	FilterSymbolsPercent    int                `json:"filterSymbolsPercent"`
	FilterCaps              bool               `json:"filterCaps"`
	Cooldown                int                `json:"cooldown"`
	SubscriberAlert         bool               `json:"subscriberAlert"`
	UrbanEnabled            bool               `json:"urbanEnabled"`
	FilterCapsPercent       int                `json:"FilterCapsPercent"`
	ParseYoutube            bool               `json:"parseYoutube"`
	ExtraLifeID             int                `json:"extraLifeID"`
	TimeoutDuration         int                `json:"timeoutDuration"`
	SteamID                 string             `json:"steamID"`
	FilterCapsMinCapitals   int                `json:"filterCapsMinCapitals"`
	SignKicks               bool               `json:"signKicks"`
	ShouldModerate          bool               `json:"shouldModerate"`
	FilterCapsMinCharacters int                `json:"filterCapsMinCharacters"`
	Owners                  []string           `json:"owners"`
	FilterEmotesMax         int                `json:"filterEmotesMax"`
	FilterSymbols           bool               `json:"filterSymbols"`
	Mode                    int                `json:"mode"`
	SubscriberRegulars      bool               `json:subscriberRegulars"`
	OffensiveWords          []string           `json:"offensiveWords"`
	RollCooldown            int                `json:"rollCooldown"`
	SubMessage              string             `json:"subMessage"`
	RollDefault             int                `json:"rollDefault"`
	UseFilters              bool               `json:"useFilters"`
	CommercialLength        int                `json":commercialLength"`
	ClickToTweetFormat      string             `json:"clickToTweetFormat"`
	Regulars                []string           `json:"regulars"`
	EnableWarnings          bool               `json:"enableWarnings"`
	PermittedDomains        []string           `json:"permittedDomains"`
	Moderators              []string           `json:"moderators"`
	RollTimeout             bool               `json:"rollTimeout"`
	FilterEmotes            bool               `json:"filterEmotes"`
	FilterSymbolsMin        int                `json:"filterSymbolsMin"`
	FilterMaxLength         int                `json:"filterMaxLength"`
	LastFM                  string             `json:"lastfm"`
	IgnoredUsers            []string           `json:"ignoredUsers"`
	Bullet                  string             `json:"bullet"`
	RepeatedCommands        []RepeatedCommand  `json:"repeatedCommands"`
	ScheduledCommands       []ScheduledCommand `json:"scheduledCommands"`
	Autoreplies             []Autoreply        `json:"autoReplies"`
}

func main() {
	//load the config for the converter itself, and add listed channels to a map for convenience of checking if they exist
	config := Config{}
	configFile, err := ioutil.ReadFile("converter.json.secret")
	check(err)
	enabledChannels := make(map[string]interface{})

	if err := json.Unmarshal(configFile, &config); err != nil {
		panic(err)
	}
	for _, u := range strings.Split(config.Channels, ",") {
		fmt.Println(strings.TrimPrefix(u, "#"))
		enabledChannels[strings.TrimPrefix(u, "#")] = ""
	}
	//Connect to the database
	db, err := sql.Open("mysql", config.DBUser+":"+config.DBPass+"@tcp("+config.DBHost+")/"+config.DBName)
	if err != nil {
		panic(err)
	}

	//find all the files in the current directory, and iterate through them if they start with # and end with .json
	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "#") {
			fmt.Println(f.Name())
			channel := Channel{}
			channelFile, _ := ioutil.ReadFile(f.Name())

			if err := json.Unmarshal(channelFile, &channel); err != nil {
				fmt.Println(err.Error())
				break
			}
			//trim off everything except the name of the channel, and get the channel ID from twitch
			channelName := strings.TrimSuffix(f.Name(), ".json")
			channelName = strings.TrimPrefix(channelName, "#")
			tempUser := Users{}
			err := getJson("https://api.twitch.tv/kraken/users?login="+channelName+"&api_version=5", &tempUser)
			check(err)
			var channelID string
			if len(tempUser.Members) > 0 {
				channelID = tempUser.Members[0].Id
			} else {
				continue
			}

			//prepare the insert statement
			stmt, err := db.Prepare("INSERT INTO channel(channel_name,channel_ID,enabled,commandPrefix,bullet," +
				"cooldown,mode,lastfm,extraLifeID,timeoutDuration,commercialLength,signKicks,enableWarnings," +
				"useFilters,steamID,shouldModerate,subscriberAlert,subscriberRegulars,subMessage,clickToTweetFormat," +
				"parseYoutube,urbanEnabled,rollTimeout,rollLevel,rollCooldown,rollDefault) " +
				"VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")

			//since mysql doesnt support literal booleans, we're going to use values of 'Y' and 'N' instead, this converts it
			enabled, signKicks, enableWarnings, useFilters, shouldModerate, subscriberAlert, subscriberRegulars, parseYoutube, urbanEnabled, rollTimeout := "N", "N", "N", "N", "N", "N", "N", "N", "N", "N"
			_, isEnabled := enabledChannels[channelName]
			if isEnabled {
				enabled = "Y"
			}
			if channel.SignKicks {
				signKicks = "Y"
			}
			if channel.EnableWarnings {
				enableWarnings = "Y"
			}
			if channel.UseFilters {
				useFilters = "Y"
			}
			if channel.ShouldModerate {
				shouldModerate = "Y"
			}
			if channel.SubscriberAlert {
				subscriberAlert = "Y"
			}
			if channel.SubscriberRegulars {
				subscriberRegulars = "Y"
			}
			if channel.ParseYoutube {
				parseYoutube = "Y"
			}
			if channel.UrbanEnabled {
				urbanEnabled = "Y"
			}
			if channel.RollTimeout {
				rollTimeout = "Y"
			}

			//execute the statement, if there's an error with the statement, we'll continue to the next user
			_, err = stmt.Exec(channelName, channelID, enabled, channel.CommandPrefix, channel.Bullet, channel.Cooldown,
				channel.Mode, channel.LastFM, channel.ExtraLifeID, channel.TimeoutDuration, channel.CommercialLength,
				signKicks, enableWarnings, useFilters, channel.SteamID, shouldModerate, subscriberAlert, subscriberRegulars,
				channel.SubMessage, channel.ClickToTweetFormat, parseYoutube, urbanEnabled, rollTimeout, channel.RollLevel,
				channel.RollCooldown, channel.RollDefault)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			//again converting between booleans
			filterLinks, filterOffensive, filterCaps, filterSymbols, filterEmotes := "N", "N", "N", "N", "N"

			if channel.FilterLinks {
				filterLinks = "Y"
			}
			if channel.FilterOffensive {
				filterOffensive = "Y"
			}
			if channel.FilterCaps {
				filterCaps = "Y"
			}

			if channel.FilterSymbols {
				filterSymbols = "Y"
			}
			if channel.FilterEmotes {
				filterEmotes = "Y"
			}

			//inserting filters
			stmt, err = db.Prepare("INSERT INTO filters(channel_ID,filterLinks,filterOffensive,filterCaps,filterSymbols," +
				"filterEmotes,filterSymbolsPercent,filterCapsPercent,filterCapsMinCapitals,filterCapsMinCharacters,filterSymbolsMin," +
				"filterMaxLength, filterEmotesMax) " +
				"VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)")
			check(err)

			_, err = stmt.Exec(channelID, filterLinks, filterOffensive, filterCaps, filterSymbols, filterEmotes, channel.FilterSymbolsPercent,
				channel.FilterCapsPercent, channel.FilterCapsMinCapitals, channel.FilterCapsMinCharacters, channel.FilterSymbolsMin, channel.FilterMaxLength, channel.FilterEmotesMax)
			check(err)

			//inserting commands
			for i := 0; i < len(channel.Commands); i++ {
				stmt, err = db.Prepare("INSERT INTO commands(channel_ID,editor,count,restriction,`key`,value) VALUES(?,?,?,?,?,?)")
				check(err)

				_, err = stmt.Exec(channelID, channel.Commands[i].Editor, channel.Commands[i].Count, channel.Commands[i].Restriction,
					channel.Commands[i].Key, channel.Commands[i].Value)
				check(err)
				if err != nil {
					fmt.Println(err.Error())
					//fmt.Println(channel.Commands[i].Value)
				}
			}
			//inserting quotes
			stmt, err = db.Prepare("INSERT INTO lists(channel_ID,list_name) VALUES (?,?)")
			_, err = stmt.Exec(channelID, "quote")
			for i := 0; i < len(channel.Quotes); i++ {
				stmt, err = db.Prepare("INSERT INTO list_items(channel_ID,list_name,item, `index`) VALUES(?,?,?,?)")
				check(err)

				_, err = stmt.Exec(channelID, "quote", channel.Quotes[i].Text, i)
				check(err)

			}
			userpermissions := make(map[string]string)
			//inserting regulars
			for i := 0; i < len(channel.Regulars); i++ {
				userpermissions[channel.Regulars[i]] = "regular"
				//userID := getChannelID(channel.Regulars[i])
				//stmt, err = db.Prepare("INSERT INTO users(channel_ID,user_ID) VALUES(?,?)")
				//check(err)
				//_, err = stmt.Exec(channelID, channel.Regulars[i])
				//check(err)
			}
			//inserting moderators
			for i := 0; i < len(channel.Moderators); i++ {
				userpermissions[channel.Moderators[i]] = "moderator"
				//stmt, err = db.Prepare("INSERT INTO moderators(channel_ID,name) VALUES(?,?)")
				//check(err)
				//_, err = stmt.Exec(channelID, channel.Moderators[i])
				//check(err)
			}
			//inserting owners
			for i := 0; i < len(channel.Owners); i++ {
				userpermissions[channel.Owners[i]] = "owner"
				//stmt, err = db.Prepare("INSERT INTO owners(channel_ID,name) VALUES(?,?)")
				//check(err)
				//_, err = stmt.Exec(channelID, channel.Owners[i])
				//check(err)
			}
			//inserting ignored_users
			for i := 0; i < len(channel.IgnoredUsers); i++ {
				userpermissions[channel.IgnoredUsers[i]] = "ignored"
				//stmt, err = db.Prepare("INSERT INTO ignored_users(channel_ID,name) VALUES(?,?)")
				//check(err)
				//_, err = stmt.Exec(channelID, channel.IgnoredUsers[i])
				//check(err)
			}
			for k, v := range userpermissions {
				userID := getChannelID(k)
				if userID != "" {

					stmt, err = db.Prepare("INSERT INTO users(channel_ID,user_ID,userlevel) VALUES(?,?,?)")
					check(err)
					_, err = stmt.Exec(channelID, userID, v)
					check(err)
				}
			}

			//inserting permitted domains
			for i := 0; i < len(channel.PermittedDomains); i++ {
				stmt, err = db.Prepare("INSERT INTO permitteddomains(channel_ID,domain) VALUES(?,?)")
				check(err)
				_, err = stmt.Exec(channelID, channel.PermittedDomains[i])
				check(err)
			}
			//inserting autoreplies
			for i := 0; i < len(channel.Autoreplies); i++ {
				stmt, err = db.Prepare("INSERT INTO autoreplies(channel_ID,`trigger`,`response`,`index`) VALUES(?,?,?,?)")
				check(err)
				_, err = stmt.Exec(channelID, channel.Autoreplies[i].Trigger, channel.Autoreplies[i].Response, i)
				check(err)
			}
			//inserting repeated commands
			for i := 0; i < len(channel.RepeatedCommands); i++ {
				stmt, err = db.Prepare("INSERT INTO repeated_commands(channel_ID,name,active,delay,messageDifference) VALUES(?,?,?,?,?)")
				check(err)
				active := "N"
				if channel.RepeatedCommands[i].Active {
					active = "Y"
				}
				_, err = stmt.Exec(channelID, channel.RepeatedCommands[i].Name, active, channel.RepeatedCommands[i].Delay, channel.RepeatedCommands[i].MessageDifference)
				check(err)
			}
			//inserting scheduled commands
			for i := 0; i < len(channel.ScheduledCommands); i++ {
				stmt, err = db.Prepare("INSERT INTO scheduled_commands(channel_ID,name,active,pattern,messageDifference) VALUES(?,?,?,?,?)")
				check(err)
				active := "N"
				if channel.ScheduledCommands[i].Active {
					active = "Y"
				}
				_, err = stmt.Exec(channelID, channel.ScheduledCommands[i].Name, active, channel.ScheduledCommands[i].Pattern, channel.ScheduledCommands[i].MessageDifference)
				check(err)
			}
			//insert offensive words
			for i := 0; i < len(channel.OffensiveWords); i++ {
				stmt, err = db.Prepare("INSERT INTO offensivewords(channel_ID,phrase) VALUES(?,?)")
				check(err)
				_, err = stmt.Exec(channelID, channel.OffensiveWords[i])
				check(err)
			}
			//inserting lists
			for k := range channel.Lists {
				stmt, err = db.Prepare("INSERT INTO lists(channel_ID,list_name,restriction) VALUES(?,?,?)")
				check(err)
				_, err = stmt.Exec(channelID, k, channel.Lists[k].Restriction)
				check(err)

				for i := 0; i < len(channel.Lists[k].Items); i++ {
					stmt, err = db.Prepare("INSERT INTO list_items(channel_ID,list_name,item,`index`) VALUES(?,?,?,?)")
					check(err)
					_, err = stmt.Exec(channelID, k, channel.Lists[k].Items[i], i)
					check(err)
				}
			}

		}

	}

}

//check errors
func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
	}
}

//json downloading helper function
func getJson(url string, target interface{}) error {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Client-ID", "q6batx0epp608isickayubi39itsckt")
	req.Header.Set("Accept", "application/vnd.twitchtv.v5.json")
	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getChannelID(name string) string {
	tempUser := Users{}
	err := getJson("https://api.twitch.tv/kraken/users?login="+name+"&api_version=5", &tempUser)
	check(err)
	var channelID string
	if len(tempUser.Members) > 0 {
		channelID = tempUser.Members[0].Id
	} else {
		channelID = ""
	}
	return channelID
}
