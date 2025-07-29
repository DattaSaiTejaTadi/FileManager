package main

import (
	"database/sql"
	"fm/handler"
	"fm/models"
	"fm/service"
	"fm/store"
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
	s3config := models.S3Config{
		Endpoint:  configs.GetConfig("S3_ENDPOINT"),
		Region:    "auto",
		AccessKey: configs.GetConfig("S3_ACCESS_KEY"),
		SecretKey: configs.GetConfig("S3_SECRET_KEY"),
		Bucket:    configs.GetConfig("S3_BUCKET"),
	}
	S3client, err := store.NewS3Client(s3config)
	if S3client == nil {
		log.Fatal(" S3 coocnetion failed")
	}
	if err != nil {
		log.Println("Issue in initializing s3 client", err.Error())
	}
	db := initializeDB(configs, "")
	log.Println(configs)
	if db == nil {
		log.Fatal("DB connection failed")
	}
	runMigrations(configs)

	storeDB := store.New(db)
	serviceLayer := service.New(storeDB, *S3client)
	handlerLayer := handler.New(serviceLayer)
	r := fiber.New()
	r.Get("/getobjects", func(ctx fiber.Ctx) error {
		fileinfo, err := S3client.ListFolder("")
		if err != nil {
			ctx.Status(400).JSON(map[string]any{
				"Code":    400,
				"Message": "some issue in fetching s3 objects",
			})
			return nil
		}
		ctx.Status(200).JSON(
			map[string]any{
				"Code":    200,
				"Message": "Successfully fetched objects",
				"Data":    fileinfo,
			})
		return nil
	})
	r.Post("/createFolder", handlerLayer.CreateFolder)
	r.Get("/getFolder", handlerLayer.GetAllFolders)
	r.Listen(":" + configs.GetConfig("PORT"))
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
