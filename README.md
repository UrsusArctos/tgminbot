<img src="https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white" />&nbsp;<img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
# tgminbot
Minimalistic Telegram Bot package written in Go. Get your bot running in minutes.

## Current features

    * Sending text-only replies with quoting support
	* Sending audio files

## Planned features

    * Sending arbitrary files

## Demo

```go
package main

import (
	"fmt"
	"io/ioutil"
	mbot "projects/tgminbot/minbotcore"
)

var tgb mbot.TGMinBotCore

func ActualHandler(tgmsg mbot.JSONStruct) {
	// Show received message
	fmt.Printf("%s [%d]: %s \n", mbot.TGMSGGetFromUsername(tgmsg), mbot.TGMSGGetFromID(tgmsg), mbot.TGMSGGetText(tgmsg))
	// Send quoted reply
	_, err := tgb.SendMessage_AsReplyTo(fmt.Sprintf("Hello, %s!", mbot.TGMSGGetFromUsername(tgmsg)), tgmsg)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	// Send mp3 file
	afile := mbot.AttachedFileData{LocalFile: "sample.mp3",
		Caption: "Downloaded using @" + tgb.UserName, Performer: "Demo", Title: "Sample Sound",
	}
	_, err = tgb.SendMessage_Audio(afile, mbot.TGMSGGetFromID(tgmsg))
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}

func main() {
	// Read Bot API token from file
	token, _ := ioutil.ReadFile("token.txt")
	// Initialize bot
	tgb = mbot.NewInstance(string(token))
	fmt.Println("Started as @" + tgb.UserName)
	// Set message handler
	tgb.MSGHandler = ActualHandler
	// Run message loop
	for tgb.LoadMessages() {
	}
	// All done
	fmt.Printf("Stopped at message ID = %d\n", tgb.LastUpdateID)
}
