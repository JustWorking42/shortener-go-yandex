package storage

var serverStorage map[string]string

func Init() {
	serverStorage = make(map[string]string)
}

func GetStorage() *map[string]string {
	return &serverStorage
}
