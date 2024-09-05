package utils_test

import (
	"fmt"
	"node-herder/mocks"
	"node-herder/models/metrics"
	"node-herder/repository"
	"node-herder/utils/storage"
	"time"
)

func CreateMetricsRepo(filename string) (metrics.Repository, *mocks.MockClock, error) {
	kvdb, err := storage.NewBoltKeyValueDatabase(filename, "metrics")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialise keyvalue db: %s", err.Error())
	}

	mockClock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})
	repo, err := repository.NewMetricsRepoFromDatabase(kvdb, mockClock)

	if err != nil {
		return nil, nil, err
	}

	return repo, mockClock, nil
}
