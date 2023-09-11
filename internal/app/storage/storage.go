package storage

import "errors"

var serverStorage map[string]string

func Init() error {
	serverStorage = make(map[string]string)
	return nil
}

func GetStorage() (*map[string]string, error) {
	if serverStorage == nil {
		return nil, errors.New("nil pointer exception call Init before")
	}
	return &serverStorage, nil
}
