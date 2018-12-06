package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*
checks for any errors and logs them
*/
func check(e error) {
	if e != nil {
		fmt.Println("error: ", e.Error())
	}
}

/*
Utility Method to convert from mysql boolean ('Y' or 'N') to bool
returns true if "Y" or false if anything else
*/
func boolCheck(s string) bool {
	return strings.EqualFold("Y", s) || strings.EqualFold(s, "1")
}

//json downloading helper function
func getTwitchJson(url string, target interface{}) error {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Client-ID", MConfig.TwitchClientID)
	req.Header.Set("Accept", "application/vnd.twitchtv.v5.json")
	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func isInteger(str string) bool {
	if _, err := strconv.Atoi(str); err == nil {
		return true
	} else {
		return false
	}
}
