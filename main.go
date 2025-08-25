package main

import (
	"database/sql"
	handlerFiles "fm/handler/files"
	handlerFolders "fm/handler/folders"
	svcFiles "fm/service/files"
	svcFolders "fm/service/folders"
	"fm/store"
	"fm/store/buckets"
	"fm/store/files"
	"fm/store/folders"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/syntaxLabz/configManager/pkg/configManager"
	"github.com/syntaxLabz/errors/pkg/codes"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

func main() {
	configs := configManager.New()

	db := initializeDB(configs, "")
	log.Println(configs)
	if db == nil {
		log.Fatal("DB connection failed")
	}
	runMigrations(configs)
	r := fiber.New()
	bucket := buckets.New(
		configs.GetConfig("S3_ENDPOINT"),
		configs.GetConfig("S3_BUCKET"),
		configs.GetConfig("S3_TOKEN"),
	)
	initializeFolderRoutes(r, db, bucket)
	initializeFileRoutes(r, db, bucket)

	r.Listen(":" + configs.GetConfig("HTTP_PORT"))
}
func initializeFolderRoutes(app *fiber.App, db *sql.DB, bucket store.Bucket) {
	folderStore := folders.New(db)
	foldersvc := svcFolders.New(folderStore, bucket)
	folderHanlde := handlerFolders.New(foldersvc)

	app.Post("/folder", folderHanlde.Create)
	app.Get("/folder", folderHanlde.GetALL)
	app.Get("/folder/:id", folderHanlde.GetById)
	app.Get("/folder/:id/subfolders", folderHanlde.GetSubFolders)
}

func initializeFileRoutes(app *fiber.App, db *sql.DB, bucket store.Bucket) {
	fileStore := files.New(db)
	folderStore := folders.New(db)
	filesvc := svcFiles.New(fileStore, folderStore, bucket)
	fileHandler := handlerFiles.New(filesvc)

	app.Post("/file", fileHandler.Create)
	app.Get("/file/:id", fileHandler.GetById)
	app.Get("/folder/:folderId/files", fileHandler.GetFiles)
}

func runMigrations(configs *configManager.Config) {
	dbConfig := intializeDBConfigs(configs, "")
	connStr := generateConnectionString(dbConfig)

	m, err := migrate.New(
		"file://migrations",
		connStr,
	)

	if err != nil {
		log.Fatal("Migration setup failed:", err)
	}

	// Apply migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Migrations applied successfully")
}

type dbConfig struct {
	host                  string
	password              string
	user                  string
	port                  string
	dialect               string
	dbName                string
	sslMode               string
	maxOpenConns          int
	maxIdleConns          int
	connMaxLifeTime       int
	idleConnectionTimeout int
	monitoringEnable      bool
}

func intializeDBConfigs(c *configManager.Config, prefix string) dbConfig {
	var (
		maxConnections, maxIdleConnections, connectionMaxLifeTime, idleConnectionTimeout int
		monitoring                                                                       bool
		err                                                                              error
	)

	maxIdleConnections, err = strconv.Atoi(c.GetConfig(prefix + "DB_MAX_IDLE_CONNECTIONS"))
	if err != nil {
		maxIdleConnections = 5
	}

	maxConnections, err = strconv.Atoi(c.GetConfig(prefix + "DB_MAX_CONNECTIONS"))
	if err != nil {
		maxConnections = 20
	}

	connectionMaxLifeTime, err = strconv.Atoi(c.GetConfig(prefix + "DB_CONNECTIONS_MAX_LIFETIME"))
	if err != nil {
		connectionMaxLifeTime = 15
	}

	idleConnectionTimeout, err = strconv.Atoi(c.GetConfig(prefix + "DB_IDLE_CONNECTIONS_TIMEOUT"))
	if err != nil {
		idleConnectionTimeout = 10
	}

	monitoring, err = strconv.ParseBool(c.GetConfig(prefix + "DB_MONITORING"))
	if err != nil {
		monitoring = false
	}

	dbConfig := dbConfig{
		host:                  c.GetConfig(prefix + "DB_HOST"),
		password:              c.GetConfig(prefix + "DB_PASSWORD"),
		user:                  c.GetConfig(prefix + "DB_USER"),
		port:                  c.GetConfig(prefix + "DB_PORT"),
		dialect:               c.GetConfig(prefix + "DB_DIALECT"),
		dbName:                c.GetConfig(prefix + "DB_NAME"),
		sslMode:               c.GetConfig(prefix + "DB_SSL"),
		maxOpenConns:          maxConnections,
		maxIdleConns:          maxIdleConnections,
		connMaxLifeTime:       connectionMaxLifeTime,
		idleConnectionTimeout: idleConnectionTimeout,
		monitoringEnable:      monitoring,
	}

	return dbConfig
}

func initializeDB(c *configManager.Config, prefix string) *sql.DB {
	dbConfig := intializeDBConfigs(c, prefix)

	if dbConfig.host != "" && dbConfig.port != "" && dbConfig.user != "" && dbConfig.password != "" && dbConfig.dialect != "" {
		if dbConfig.sslMode == "" {
			dbConfig.sslMode = "disable"
		}

		db, err := establishDBConnection(dbConfig)
		if err == nil {
			db.SetMaxOpenConns(dbConfig.maxOpenConns)
			db.SetMaxIdleConns(dbConfig.maxIdleConns)
			db.SetConnMaxLifetime(time.Minute * time.Duration(dbConfig.connMaxLifeTime))
			db.SetConnMaxIdleTime(time.Minute * time.Duration(dbConfig.idleConnectionTimeout))

			return db
		}
	}

	return nil
}

func establishDBConnection(c dbConfig) (*sql.DB, error) {
	connectionString := generateConnectionString(c)
	if connectionString == "" {
		return nil, httperrors.New(codes.InternalServerError, "Invalid dialect")
	}

	db, err := sql.Open(c.dialect, connectionString)
	if err != nil {
		log.Println("Error while connecting to database", err)
		return db, err
	}

	err = db.Ping()
	if err != nil {
		log.Println("Error while pinging to database", err)
		return db, err
	}

	return db, nil
}

func generateConnectionString(c dbConfig) string {
	switch c.dialect {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", c.user, c.password, c.host, c.port, c.dbName)
	case "postgres":
		return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", c.user, c.password, c.host, c.port, c.dbName, c.sslMode)
	}

	return ""
}
