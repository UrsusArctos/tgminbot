package minbotcore

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	apiBaseURL     = "https://api.telegram.org/bot"
	apiGetMe       = "getMe"
	apiGetUpdates  = "getUpdates"
	apiSendMessage = "sendMessage"
	// Misc
	apiWaitTime = 30
	apiMIMEType = "application/json"
)

type (
	TGMinBotCore struct {
		DisplayName  string
		UserName     string
		APIToken     string
		BotID        int64
		LastUpdateID int64
		MSGHandler   TGMessageHandler
	}

	JSONStruct = map[string]interface{}

	TGMessageHandler func(tgmsg JSONStruct)
)

func TGMSGGetText(jsmsg JSONStruct) string {
	return jsmsg["message"].(JSONStruct)["text"].(string)
}

func TGMSGGetMessageID(jsmsg JSONStruct) int64 {
	idval, _ := jsmsg["message"].(JSONStruct)["message_id"].(json.Number).Int64()
	return idval
}

func TGMSGGetFromID(jsmsg JSONStruct) int64 {
	idval, _ := jsmsg["message"].(JSONStruct)["from"].(JSONStruct)["id"].(json.Number).Int64()
	return idval
}

func TGMSGGetFromUsername(jsmsg JSONStruct) string {
	return jsmsg["message"].(JSONStruct)["from"].(JSONStruct)["username"].(string)
}

func NewInstance(BOTToken string) (tgbc TGMinBotCore) {
	resp, err := http.Post(apiBaseURL+BOTToken+"/"+apiGetMe, apiMIMEType, nil)
	if err == nil {
		defer resp.Body.Close()
		var bodydata JSONStruct
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		err := decoder.Decode(&bodydata)
		if err == nil {
			if bodydata["ok"].(bool) {
				result := bodydata["result"].(JSONStruct)
				tgbc.DisplayName = result["first_name"].(string)
				tgbc.UserName = result["username"].(string)
				tgbc.APIToken = BOTToken
				tgbc.BotID, _ = result["id"].(json.Number).Int64()
				tgbc.LastUpdateID = 0
				tgbc.MSGHandler = nil
			}
		}
	}
	return tgbc
}

func (tgbc TGMinBotCore) jsonRPC(instruct JSONStruct, apiMethod string) (outstruct JSONStruct, err error) {
	jsonval, _ := json.Marshal(instruct)
	resp, err := http.Post(apiBaseURL+tgbc.APIToken+"/"+apiMethod, apiMIMEType, bytes.NewBuffer(jsonval))
	if err == nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		err := decoder.Decode(&outstruct)
		if err == nil {
			return outstruct, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (tgbc *TGMinBotCore) LoadMessages() bool {
	APIReq := JSONStruct{"offset": tgbc.LastUpdateID + 1, "timeout": apiWaitTime}
	APIResp, err := tgbc.jsonRPC(APIReq, apiGetUpdates)
	if err == nil {
		if APIResp["ok"].(bool) {
			Results := APIResp["result"].([]interface{})
			if len(Results) > 0 {
				for _, msgstruct := range Results {
					if tgbc.MSGHandler != nil {
						tgbc.MSGHandler(msgstruct.(JSONStruct))
						newuid, _ := msgstruct.(JSONStruct)["update_id"].(json.Number).Int64()
						if newuid > tgbc.LastUpdateID {
							tgbc.LastUpdateID = newuid
						}
					}
				}
			}
			return true
		}
	}
	return false
}

func (tgbc TGMinBotCore) SendMessage_PlainText(msgtext string, chatid int64, replyto int64) bool {
	APIReq := JSONStruct{"chat_id": chatid, "text": msgtext}
	if replyto != 0 {
		APIReq["reply_to_message_id"] = replyto
	}
	APIResp, err := tgbc.jsonRPC(APIReq, apiSendMessage)
	if err == nil {
		return APIResp["ok"].(bool)
	}
	return false
}
