package database

import (
	"log"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

var (
	DbClient *gorm.DB
	SqlMock  sqlmock.Sqlmock
)

func CurrentDatabase() *gorm.DB {
	return DbClient
}

func CurrentSqlMock() sqlmock.Sqlmock {
	return SqlMock
}

func Initialize() *gorm.DB {
	if !fiber.IsChild() {
		log.Println("[App] Database connecting...")
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  true,          // Disable color
		},
	)

	dbConn, err := gorm.Open(
		// Select database connection
		func() gorm.Dialector {
			switch config.Global.Database.DatabaseDriver {
			case "mysql":
				return NewMysqlConnection()
			case "sqlserver":
				return NewSqlServerConnection()
			case "sqlite":
				return NewSqliteConnection()

			// MOCK DATABASE CONNECTION ---------------------------------------------------
			case "postgresql_mock":
				return NewPostgresMockConnection()
			case "mysql_mock":
				return NewMysqlMockConnection()
			case "sqlserver_mock":
				return NewSqlServerMockConnection()

			// Default database driver is PostgreSQL
			default:
				return NewPostgresConnection()
			}
		}(),
		&gorm.Config{
			SkipDefaultTransaction: false, // Disable automatic transaction (gain performance 30% increase)
			PrepareStmt:            true,
			Logger:                 gormLogger,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: "tbl_", // table name prefix, table for `User` would be `tbl_users`
			},
		},
	)
	dbConn.Use(
		dbresolver.Register(dbresolver.Config{
			Sources:           []gorm.Dialector{},
			Replicas:          []gorm.Dialector{},
			Policy:            nil,
			TraceResolverMode: false,
		}).
			SetConnMaxIdleTime(time.Hour).
			SetConnMaxLifetime(24 * time.Hour).
			SetMaxIdleConns(config.Global.Database.DatabaseMaxIdleConns).
			SetMaxOpenConns(config.Global.Database.DatabaseMaxOpenConns),
	)

	// Check error
	if err != nil {
		log.Printf("[App] Cannot connect to database!")
		log.Fatal("[App] DatabaseError:", err)
	}

	// Post processing
	if !fiber.IsChild() {
		log.Println("[App] Database connected", color.Format(color.GREEN, "successfully!"))

		// Migrate the database schema by GORM
		// AutoMigrate(dbConn)

		// For manual migration, uncomment the line below
		// ManualMigrate()
	}

	return dbConn
}

func AutoMigrate(dbConn *gorm.DB) {
	log.Println("[App] Auto migrating the schema...")

	// Auto migrate the schema by GORM
	if err := dbConn.AutoMigrate(
		// Add your models here...
		models.GroupPermission{},
		models.Permission{},
		models.RoleGroup{},
		models.RoleGroupRole{},
		models.Role{},
		models.Menu{},
		models.User{},
	); err != nil {
		log.Fatal(err)
	}

	log.Println("[App] Auto migration completed")
}

func ManualMigrate() {
	// Migrate the schema
	migrator, err := migrate.New(
		"file://database/migrations",
		config.Global.Database.DatabaseURL)

	// Check error when create new migrate instance
	if err != nil {
		log.Fatal(err)
	}

	// Run the migrations
	log.Println("Migrating the schema...")
	if err := migrator.Up(); err != nil {
		if err != migrate.ErrNoChange {
			Rollback()
			log.Fatal(err)
		}
	}

	log.Println("[App] Migration completed, Close the database connection...")

	// Close the database connection
	migrator.Close()
}

func Rollback() {
	// Migrate the schema
	migrator, err := migrate.New(
		"file://database/migrations",
		config.Global.Database.DatabaseURL)

	// Check error when create new migrate instance
	if err != nil {
		log.Fatal(err)
	}

	// Run the migrations
	log.Println("[App] Rollback the schema...")
	if err := migrator.Steps(-1); err != nil {
		log.Fatal(err)
	}

	log.Println("[App] Rollback completed, Close the database connection...")

	// Close the database connection
	migrator.Close()
}
