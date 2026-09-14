package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kosipov/students/auth"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/onedrive"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authhttp "github.com/kosipov/students/auth/delivery/http"
	authgorm "github.com/kosipov/students/auth/repository/gorm"
	"github.com/kosipov/students/auth/usecase"
	educationalhttp "github.com/kosipov/students/educational/delivery/http"
	educationalgorm "github.com/kosipov/students/educational/repository/gorm"
	educationalusecase "github.com/kosipov/students/educational/usecase"
)

import _ "github.com/go-sql-driver/mysql"

type App struct {
	httpServer *http.Server

	authUC    auth.UseCase
	groupUC   educational.CommonGroupUseCase
	subjectUC educational.CommonSubjectUseCase
}

func NewApp() *App {
	db := InitDB()

	userRepo := authgorm.NewUserRepository(db)
	groupRepo := educationalgorm.NewGroupRepository(db)
	subjectRepo := educationalgorm.NewSubjectRepository(db)

	return &App{
		authUC: usecase.NewAuthUseCase(
			userRepo,
			viper.GetString("auth.hash_salt"),
			[]byte(viper.GetString("auth.signing_key")),
			viper.GetDuration("auth.token_ttl"),
		),
		groupUC: educationalusecase.NewGroupUseCase(
			groupRepo),
		subjectUC: educationalusecase.NewSubjectUseCase(subjectRepo, onedrive.NewClient()),
	}
}

func (a *App) Run(port string) error {
	// Init gin handler
	gin.SetMode(viperEnvVariable("GIN_MODE"))
	router := gin.New()
	router.Use(
		gin.Recovery(),
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/healthz"}}),
	)
	// Liveness probe for container healthcheck. It does not touch the DB on purpose:
	// a DB outage should not make the orchestrator restart the app in a loop.
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api",
		sessions.Sessions("student-session", newCookieStore([]byte("secret"))),
		limitRequestBody,
		requireCSRFHeader,
	)
	authhttp.RegisterHTTPEndpoints(api, a.authUC)
	admin := api.Group("/admin", authhttp.NewAuthMiddleware())
	educationalhttp.RegisterHTTPEndpoints(api, admin, a.subjectUC, a.groupUC)

	registerSPA(router, "web/dist")

	// HTTP Server
	a.httpServer = &http.Server{
		Addr:        ":" + port,
		Handler:     router,
		ReadTimeout: 10 * time.Second,
		// A task that has never been downloaded waits for OneDrive for up to 15 seconds.
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	// Docker (and Dokploy on top of Swarm) stops containers with SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}

// InitDB connects to MySQL using MYSQL_* env variables and migrates the schema.
func InitDB() *gorm.DB {
	user := viperEnvVariable("MYSQL_USER")
	pass := viperEnvVariable("MYSQL_PASSWORD")
	host := viperEnvVariable("MYSQL_HOST")
	dbname := viperEnvVariable("MYSQL_DBNAME")
	// parseTime is required to scan DATETIME columns into time.Time.
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbname)

	client, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error occured while establishing connection to gorm %s", err.Error())
	}

	client.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Subject{},
		&models.SubjectObject{},
		&models.SubjectObjectCategory{},
	)

	return client
}

func viperEnvVariable(key string) string {
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	value, ok := viper.Get(key).(string)

	if !ok {
		value = os.Getenv(key)
		if value == "" {
			log.Fatalf("Invalid type assertion with key %s", key)
		}
	}

	return value
}
