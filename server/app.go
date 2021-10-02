package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kosipov/students/auth"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	authhttp "github.com/kosipov/students/auth/delivery/http"
	authgorm "github.com/kosipov/students/auth/repository/gorm"
	"github.com/kosipov/students/auth/usecase"
	educationalhttp "github.com/kosipov/students/educational/delivery/http"
	educationalgorm "github.com/kosipov/students/educational/repository/gorm"
	educationalusecase "github.com/kosipov/students/educational/usecase"
)

type App struct {
	httpServer *http.Server

	authUC    auth.UseCase
	groupUC   educational.CommonGroupUseCase
	subjectUC educational.CommonSubjectUseCase
}

func NewApp() *App {
	db := initDB()

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
		subjectUC: educationalusecase.NewSubjectUseCase(subjectRepo),
	}
}

func (a *App) Run(port string) error {
	// Init gin handler
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)
	router.LoadHTMLGlob("templates/**/*.html")
	router.Static("/dist", "templates/dist")

	// Set up http handlers
	// SignUp/SignIn endpoints
	authhttp.RegisterHTTPEndpoints(router, a.authUC)
	educationalhttp.RegisterHTTPEndpoints(router, a.subjectUC, a.groupUC)

	/*	// API endpoints
		authMiddleware := authhttp.NewAuthMiddleware(a.authUC)
		api := router.Group("/api", authMiddleware)

		bmhttp.RegisterHTTPEndpoints(api, a.bookmarkUC)*/

	// HTTP Server
	a.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to listen and serve: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)
}

func initDB() *gorm.DB {
	user := viper.GetString("mysql_prod.user")
	pass := viper.GetString("mysql_prod.password")
	host := viper.GetString("mysql_prod.uri")
	dbname := viper.GetString("mysql_prod.name")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, dbname)

	client, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error occured while establishing connection to gorm")
	}

	client.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Subject{},
		&models.SubjectObject{},
	)

	return client
}
