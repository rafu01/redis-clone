package main

import "time"

type ValueStore struct {
	value      string
	expiryTime int64
}

func setExpiryTime(valueStore *ValueStore, durationInMilli int64) {
	if durationInMilli == -1 {
		valueStore.expiryTime = -1
		return
	}
	valueStore.expiryTime = time.Now().UnixMilli() + durationInMilli
}

func isExpired(valueStore *ValueStore) bool {
	if valueStore.expiryTime == -1 {
		return false
	}
	return time.Now().UnixMilli() > valueStore.expiryTime
}
