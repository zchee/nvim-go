package main

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
)

type RpluginFormatter struct {
}

func (f *RpluginFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+3)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	prefixFieldClashes(data)

	// data["time"] = entry.Time.Format(timestampFormat)
	data["cmd"] = entry.Data["cmd"]
	data["msg"] = entry.Message
	data["level"] = entry.Level.String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func prefixFieldClashes(data logrus.Fields) {
	_, ok := data["cmd"]
	if ok {
		data["fields.cmd"] = data["cmd"]
	}

	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}

	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}
}
