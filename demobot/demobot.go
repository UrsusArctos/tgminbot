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
