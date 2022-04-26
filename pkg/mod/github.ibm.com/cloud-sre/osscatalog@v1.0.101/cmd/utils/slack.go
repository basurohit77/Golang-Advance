package utils

import (
	"github.com/pkg/errors"
	"github.ibm.com/cloud-sre/osscatalog/notification"
)

type MessageType int32

const (
	ErrorType   MessageType = 0
	WarningType MessageType = 1
	InfoType    MessageType = 2
	FatalType   MessageType = 3
)

func PostSlackMessage(appName, channel, title, body string, messageType MessageType, email string) error {
	if channel != "" {
		var color string
		switch messageType {
		case ErrorType:
			color = "#FF0000"
		case WarningType:
			color = "#FFFF00"
		case FatalType:
			color = "#000000"
		case InfoType:
			color = "#36a64f"
		}
		pretext := "*" + appName + "* " + title
		attachments := make([]map[string]interface{}, 1)
		attachment := make(map[string]interface{})
		attachment["pretext"] = pretext
		attachment["text"] = body
		attachment["color"] = color
		attachments[0] = attachment
		return notification.PostMessage(channel, attachments, email)
	}
	return errors.New("slack channel is required parameter.")
}
func ExitHandler(appName, channel, email, recoveredErrorMsg string) {
	if channel == "" {
		return
	}
	if recoveredErrorMsg != "" {
		PostSlackMessage(appName, channel, "Abnormal system exit", recoveredErrorMsg, FatalType, email)
	}
}
