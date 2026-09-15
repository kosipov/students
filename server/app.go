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
	"github.com/kosipov/students/campus"
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
	schedulehttp "github.com/kosipov/students/schedule/delivery/http"
	schedulegorm "github.com/kosipov/students/schedule/repository/gorm"
	scheduleusecase "github.com/kosipov/students/schedule/usecase"
)

import _ "github.com/go-sql-driver/mysql"

type App struct {
	httpServer *http.Server

	authUC     auth.UseCase
	groupUC    educational.CommonGroupUseCase
	subjectUC  educational.CommonSubjectUseCase
	scheduleUC *scheduleusecase.UseCase
	// syncSchedule is off when no teacher is configured.
	syncSchedule         bool
	scheduleSyncInterval time.Duration
}

func NewApp() *App {
	db := InitDB()

	userRepo := authgorm.NewUserRepository(db)
	groupRepo := educationalgorm.NewGroupRepository(db)
	subjectRepo := educationalgorm.NewSubjectRepository(db)

	teacher := viper.GetString("campus.teacher")
	syncInterval := viper.GetDuration("campus.sync_interval")
	if syncInterval < 5*time.Minute {
		// Protects the university's site from a typo like "30s".
		syncInterval = 30 * time.Minute
	}
	scheduleSource := campus.Source{Client: campus.NewClient(teacher, campusClientOptions()...)}

	return &App{
		scheduleUC:           scheduleusecase.NewUseCase(schedulegorm.NewRepository(db), scheduleSource),
		syncSchedule:         teacher != "",
		scheduleSyncInterval: syncInterval,
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
	schedulehttp.RegisterHTTPEndpoints(api, admin, a.scheduleUC)

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

	backgroundCtx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	if a.syncSchedule {
		go a.scheduleUC.Run(backgroundCtx, a.scheduleSyncInterval)
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
	stopBackground()

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}

// campusClientOptions reads how to reach the university's site when it doesn't answer the server's address:
// through a relay (CAMPUS_RELAY_URL, CAMPUS_RELAY_SECRET) or a proxy (CAMPUS_PROXY_URL). Env variables,
// not config.yml, because they hold secrets.
func campusClientOptions() []campus.Option {
	var options []campus.Option

	if relayURL := os.Getenv("CAMPUS_RELAY_URL"); relayURL != "" {
		secret := os.Getenv("CAMPUS_RELAY_SECRET")
		if secret == "" {
			// Without the secret the relay would answer nobody, and the schedule would never load.
			log.Fatalf("CAMPUS_RELAY_URL is set without CAMPUS_RELAY_SECRET")
		}
		options = append(options, campus.WithRelay(relayURL, secret))
		log.Printf("Campus schedule is requested through relay %s", relayURL)
	}

	if raw := os.Getenv("CAMPUS_PROXY_URL"); raw != "" {
		proxy, err := campus.ParseProxyURL(raw)
		if err != nil {
			// Failing on start is better than silently going around the proxy.
			log.Fatalf("Invalid CAMPUS_PROXY_URL: %s", err)
		}
		options = append(options, campus.WithProxy(proxy))
		log.Printf("Campus schedule is requested through proxy %s", proxy.Redacted())
	}

	return options
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
		&models.ScheduleLesson{},
		&models.ScheduleSync{},
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
