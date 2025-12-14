package models

import (
	"fmt"
	"node-herder/utils"
	"time"
)

type TimestampedKeyGenerator interface {
	CreateKeyFromTimestamp(id string, timestamp time.Time) []byte
	CreateKey(id string) []byte
	GetTimestampFromkey(bytes []byte) (time.Time, error)
	CreateKeyPrefixFromTimestamp(timestamp time.Time) []byte
	CreateMaxKeyFromTimestamp(timestamp time.Time) []byte
	GetIdFromKey(bytes []byte) (string, error)
}

type TimestampedKeyGeneratorImp struct {
	timeFormat string
	clock      utils.Clock
}

func NewTimestampedKeyGenerator(clock utils.Clock) TimestampedKeyGenerator {
	return &TimestampedKeyGeneratorImp{
		timeFormat: "2006-01-02T15:04:05.000000000Z",
		clock:      clock,
	}
}

func (k *TimestampedKeyGeneratorImp) CreateKey(id string) []byte {

	timestampStr := k.clock.Now().Format(k.timeFormat)
	key := fmt.Sprintf("%s_%s", timestampStr, id)

	return []byte(key)
}

func (k *TimestampedKeyGeneratorImp) CreateKeyFromTimestamp(id string, timestamp time.Time) []byte {

	timestampStr := timestamp.Format(k.timeFormat)
	key := fmt.Sprintf("%s_%s", timestampStr, id)

	return []byte(key)
}

func (k *TimestampedKeyGeneratorImp) GetTimestampFromkey(bytes []byte) (time.Time, error) {
	timestamp, err := time.Parse(k.timeFormat, string(bytes[:30]))
	if err != nil {
		return k.clock.Now(), err
	}
	return timestamp, nil
}

func (k *TimestampedKeyGeneratorImp) CreateKeyPrefixFromTimestamp(timestamp time.Time) []byte {
	timestampStr := timestamp.Format(k.timeFormat)
	key := fmt.Sprintf("%s_", timestampStr)
	return []byte(key)
}

func (k *TimestampedKeyGeneratorImp) CreateMaxKeyFromTimestamp(timestamp time.Time) []byte {
	timestampStr := timestamp.Format(k.timeFormat)
	// 0xFF as a max byte to include all possible suffixes for the timestamp
	key := append([]byte(fmt.Sprintf("%s_", timestampStr)), 0xFF)
	return key
}

func (k *TimestampedKeyGeneratorImp) GetIdFromKey(b []byte) (string, error) {
	// Keys are formatted as: <timestamp>_<id>
	for i := 0; i < len(b); i++ {
		if b[i] == '_' {
			if i+1 < len(b) {
				return string(b[i+1:]), nil
			}
			return "", fmt.Errorf("key missing id suffix")
		}
	}
	return "", fmt.Errorf("invalid key format: underscore not found")
}
