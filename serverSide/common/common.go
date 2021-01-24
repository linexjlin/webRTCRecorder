package common

import (
	"encoding/base64"
	"encoding/json"
	"github.com/linexjlin/simple-log"
)

func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
