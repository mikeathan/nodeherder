package metrics

import (
	"fmt"
	"node-herder/utils"
	"time"
)

type TimestampedKeyGenerator interface {
	CreateKeyFromTimestamp(id string, timestamp time.Time) []byte
	CreateKey(id string) []byte
	GetTimestampFromkey(bytes []byte) (time.Time, error)
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
