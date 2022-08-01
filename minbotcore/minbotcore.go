package minbotcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"reflect"
)

const (
	apiBaseURL     = "https://api.telegram.org/bot"
	apiGetMe       = "getMe"
	apiGetUpdates  = "getUpdates"
	apiSendMessage = "sendMessage"
	apiSendAudio   = "sendAudio"
	// Misc
	apiWaitTime = 30
	apiMIMEType = "application/json"
	// Parse modes
	PMPlainText  = ""
	PMMarkdown   = "Markdown"
	PMMarkdownV2 = "MarkdownV2"
	PMHTML       = "HTML"
)

type (
	TGMinBotCore struct {
		DisplayName  string
		UserName     string
		APIToken     string
		BotID        int64
		LastUpdateID int64
		MSGParseMode string
		MSGHandler   TGMessageHandler
	}

	JSONStruct = map[string]interface{}

	TGMessageHandler func(tgmsg JSONStruct)

	AttachedFileData struct {
		LocalFile  string
		RemoteName string
		FieldName  string
		MimeType   string
		Caption    string
		Performer  string
		Title      string
	}
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

func Sent(RPCResponse JSONStruct) bool {
	return RPCResponse["ok"].(bool)
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
				tgbc.MSGParseMode = PMPlainText
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

func (tgbc TGMinBotCore) formRPC(instruct JSONStruct, apiMethod string, attFile AttachedFileData) (outstruct JSONStruct, err error) {
	body := &bytes.Buffer{}
	// Set message parameters as multipart form data
	writer := multipart.NewWriter(body)
	for key, value := range instruct {
		vt := reflect.TypeOf(value)
		switch vt.Kind() {
		case reflect.String:
			writer.WriteField(key, value.(string))
		case reflect.Int64:
			writer.WriteField(key, fmt.Sprintf("%d", value.(int64)))
		}
	}
	// Attach a file
	afile, err := os.Open(attFile.LocalFile)
	if err == nil {
		defer afile.Close()
		aheader := make(textproto.MIMEHeader)
		aheader.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, attFile.FieldName, attFile.RemoteName))
		aheader.Set("Content-Type", attFile.MimeType)
		afilepart, err := writer.CreatePart(aheader)
		if err == nil {
			io.Copy(afilepart, afile)
			// NB! writer.Close() is not deferred here because
			// the multipart closing boundary must be written before issuing HTTPS request
			writer.Close()
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
	req, err := http.NewRequest("POST", apiBaseURL+tgbc.APIToken+"/"+apiMethod, body)
	if err == nil {
		req.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
		hclient := &http.Client{}
		resp, err := hclient.Do(req)
		if err == nil {
			decoder := json.NewDecoder(resp.Body)
			decoder.UseNumber()
			err := decoder.Decode(&outstruct)
			resp.Body.Close()
			if err == nil {
				return outstruct, nil
			} else {
				return nil, err
			}
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

func (tgbc TGMinBotCore) SendMessage_PlainText(msgtext string, chatid int64, replyto int64) (outstruct JSONStruct, err error) {
	APIReq := JSONStruct{"chat_id": chatid, "text": msgtext, "parse_mode": tgbc.MSGParseMode}
	if replyto != 0 {
		APIReq["reply_to_message_id"] = replyto
		APIReq["allow_sending_without_reply"] = true
	}
	return tgbc.jsonRPC(APIReq, apiSendMessage)
}

func (tgbc TGMinBotCore) SendMessage_AsReplyTo(msgtext string, quotedmsg JSONStruct) (outstruct JSONStruct, err error) {
	return tgbc.SendMessage_PlainText(msgtext, TGMSGGetFromID(quotedmsg), TGMSGGetMessageID(quotedmsg))
}

func (tgbc TGMinBotCore) SendMessage_Audio(audiofile AttachedFileData, chatid int64) (outstruct JSONStruct, err error) {
	APIReq := JSONStruct{"chat_id": chatid,
		"caption":   audiofile.Caption,
		"performer": audiofile.Performer,
		"title":     audiofile.Title}
	audiofile.RemoteName = audiofile.Performer + " - " + audiofile.Title + ".mp3"
	audiofile.FieldName = "audio"
	audiofile.MimeType = "audio/mpeg"
	return tgbc.formRPC(APIReq, apiSendAudio, audiofile)
}
