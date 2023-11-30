package common

func HandleString(value *string, defaultValue string) {
	if *value == "" {
		*value = defaultValue
	}
}

func HandleInt(value *int, defaultValue int) {
	if value == nil || *value == 0 {
		*value = defaultValue
	}
}

func HandleInt64(value *int64, defaultValue int64) {
	if value == nil || *value == 0 {
		*value = defaultValue
	}
}
