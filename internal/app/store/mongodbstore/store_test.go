package mongodbstore_test

import (
	"os"
	"testing"
)

var (
	databaseUrl string
)

func TestMain(m *testing.M) {
	// ...
	databaseUrl = os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = "mongodb+srv://psihachina:cSTJ9Ia2nHLDYZZq@cluster0.u1rzt.gcp.mongodb.net/test_database?retryWrites=true&w=majority"
	}

	os.Exit(m.Run())
}
