package environment

import "os"

func GetInitSQLPath() string {
	switch os.Getenv("ENVIRONMENT") {
	case "development":
		return os.Getenv("INIT_SQL_PATH_DEV")
	case "production":
		return os.Getenv("INIT_SQL_PATH_PROD")
	default:
		return ""
	}
}

func GetSecretKey() []byte {
	return []byte(os.Getenv("SECRET_KEY"))
}
