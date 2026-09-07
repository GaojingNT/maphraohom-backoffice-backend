package database

import (
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
)

func NewPostgresConnection() gorm.Dialector {
	return postgres.New(
		postgres.Config{
			// Example DSN: "user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
			DSN: config.Global.Database.DatabaseDSN,
		},
	)
}

func NewSqlServerConnection() gorm.Dialector {
	return sqlserver.New(sqlserver.Config{
		// Example DSN: "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
		DSN: config.Global.Database.DatabaseDSN,
	})
}

func NewMysqlConnection() gorm.Dialector {
	return mysql.New(mysql.Config{
		// Example DSN: "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local"
		DSN: config.Global.Database.DatabaseDSN,
	})
}

func NewSqliteConnection() gorm.Dialector {
	return sqlite.Open("local.db")
}

// MOCK DATABASE CONNECTION ---------------------------------------------------
// Note: for more information about mocking database connection, please visit: https://www.codingexplorations.com/blog/testing-gorm-with-sqlmock

func NewPostgresMockConnection() gorm.Dialector {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	// Assign the mock to the SqlMock variable
	SqlMock = mock

	return postgres.New(postgres.Config{
		Conn: db,
	})
}

func NewMysqlMockConnection() gorm.Dialector {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	// Assign the mock to the SqlMock variable
	SqlMock = mock

	return mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})
}

func NewSqlServerMockConnection() gorm.Dialector {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	// Assign the mock to the SqlMock variable
	SqlMock = mock

	return sqlserver.New(sqlserver.Config{
		Conn: db,
	})
}
