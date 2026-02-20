package queue

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

type Conf struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type AlertEvent struct {
	ID          uuid.UUID
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Username    string
}

func EncodeAlertEvent(event *AlertEvent) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(event); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DecodeAlertEvent(bodyMsg []byte) (*AlertEvent, error) {
	var event AlertEvent

	buffer := bytes.NewBuffer(bodyMsg)
	dec := gob.NewDecoder(buffer)
	if err := dec.Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}
