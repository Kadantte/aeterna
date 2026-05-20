package services

import (
	"fmt"
	"os"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

type DatabaseModule struct{}

func (DatabaseModule) Name() string { return "DatabaseModule" }
func (DatabaseModule) Section() string {
	return "database"
}

func init() {
	common.Register(DatabaseModule{})
}

type DatabaseSection struct {
	Path      string
	PathIsSet bool

	DBHost       string
	PostgresHost string
	DatabaseURL  string

	EncryptionEnabled        bool
	EncryptionAutoMigrate    bool
	EncryptionKDFContextFile string
}

func (DatabaseModule) LoadAndValidate() (DatabaseSection, error) {
	rawPath := os.Getenv("DATABASE_PATH")
	section := DatabaseSection{
		Path:         common.WithDefault(common.GetenvTrim("DATABASE_PATH"), common.DefaultDatabasePath),
		PathIsSet:    rawPath != "",
		DBHost:       common.GetenvTrim("DB_HOST"),
		PostgresHost: common.GetenvTrim("POSTGRES_HOST"),
		DatabaseURL:  common.GetenvTrim("DATABASE_URL"),

		EncryptionEnabled:        common.GetBool("DB_ENCRYPTION_ENABLED", common.DefaultDBEncryptionEnabled),
		EncryptionAutoMigrate:    common.GetBool("DB_ENCRYPTION_AUTO_MIGRATE", common.DefaultDBEncryptionAutoMigrate),
		EncryptionKDFContextFile: common.DefaultDBEncryptionKDFContextFile,
	}
	if common.GetenvTrim("ENV") == "production" && !section.PathIsSet {
		return DatabaseSection{}, fmt.Errorf("DATABASE_PATH must be set in production")
	}
	return section, nil
}
