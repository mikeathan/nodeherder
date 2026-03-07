package services


// func (s *MetricsRepo) runPruningTask() {
// 	go func() {
// 		for {

//
// 			// maybe pass duration in configuration
// 			s.clock.Sleep(time.Minute) // Change it hour or day !!!!!!
// 			utils.LogInfof("Start pruning bucket %v", metricsBucketName)

// 			err := s.pruneEntries(s.db, metricsBucketName)
// 			if err != nil {
// 				utils.LogErrorf("Error pruning entries: %v", err)
// 			}
// 			utils.LogInfof("End pruning bucket %v", metricsBucketName)

// 		}
// 	}()
// }
// func (s *MetricsRepo) pruneEntries(db *bolt.DB, bucketName string) error {
// 	return db.Update(func(tx *bolt.Tx) error {
// 		bucket := tx.Bucket([]byte(bucketName))
// 		if bucket == nil {
// 			return fmt.Errorf("bucket not found: %s", bucketName)
// 		}

// 		c := bucket.Cursor()
// 		for key, _ := c.First(); key != nil; key, _ = c.Next() {

// 			fmt.Println("device id", string(key))

// 			bucket.Bucket(key).ForEach(func(k, _ []byte) error {
// 				fmt.Println("device key", string(k))
// 				// voc2024-08-25T18:02:52.455198804Z
// 				return nil
// 			})

// 			// timestamp, err := s.readTimestampFromKey(expose.Name, key)
// 			// timestamp, err := s.readTimestampFromKey(expose.Name, key)

// 			// expiresAt, err := strconv.ParseInt(strings.Split(k, "-")[0], 10, 64)
// 			// if err != nil {
// 			// 	return fmt.Errorf("error parsing expiration timestamp: %w", err)
// 			// }

// 			// if time.Now().Unix() > expiresAt {
// 			// 	if err := bucket.Delete(k); err != nil {
// 			// 		return fmt.Errorf("error deleting expired entry: %w", err)
// 			// 	}
// 			// }
// 		}

// 		return nil
// 	})
// }