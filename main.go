package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"

	irc "github.com/fluffle/goirc/client"
	_ "github.com/go-sql-driver/mysql"

	//"strings"
	//"fmt"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
)

//bot config
type Config struct {
	Nick              string `json:"nick"`
	Server            string `json:"server"`
	Pass              string `json:"pass"`
	MysqlUser         string `json:"mysqlUser"`
	MysqlPass         string `json:"mysqlPass"`
	MysqlIP           string `json:"mysqlIP"`
	MysqlDatabase     string `json:"mysqlDatabase"`
	TwitchClientID    string `json:"twitchClientID"`
	TwitchKrakenOauth string `json:"twitchKrakenOauth"`
}

var (
	Connections map[string]*irc.Conn
	MConfig     Config
)

func main() {

	//load the config file for the bot
	MConfig = Config{}
	configFile, err := ioutil.ReadFile("mainconfig.json.secret")
	check(err)
	if err := json.Unmarshal(configFile, &MConfig); err != nil {
		panic(err)
	}
	//spew.Dump(MConfig)
	//create the config for the irc client
	cfg := irc.NewConfig(MConfig.Nick)
	cfg.SSL = false
	cfg.Server = MConfig.Server
	cfg.Pass = MConfig.Pass
	//create the map that will store all of the connections (one per channel)
	Connections = make(map[string]*irc.Conn)

	//connect to the database
	db, err := sql.Open("mysql", MConfig.MysqlUser+":"+MConfig.MysqlPass+"@tcp("+MConfig.MysqlIP+":3306)/"+MConfig.MysqlDatabase)
	if err != nil {
		panic(err)
	}
	//only join channels that are 'enabled'
	rows, err := db.Query("SELECT channel_name FROM channel WHERE enabled='Y'")
	check(err)
	defer rows.Close()
	for rows.Next() {
		var channelName string
		err := rows.Scan(&channelName)
		fmt.Println(channelName)
		check(err)
		//create a new connection to twitch, #JOINing channelName
		Connections[channelName] = NewReceiver(cfg, channelName, db)
	}

	//keep the bot running
	quit := make(chan bool)
	<-quit

}
