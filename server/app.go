package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kosipov/students/auth"
	"github.com/kosipov/students/group"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	authhttp "github.com/kosipov/students/auth/delivery/http"
	authgorm "github.com/kosipov/students/auth/repository/gorm"
	"github.com/kosipov/students/auth/usecase"
	grouphttp "github.com/kosipov/students/group/http"
	groupgorm "github.com/kosipov/students/group/repository/gorm"
	groupusecase "github.com/kosipov/students/group/usecase"
)

type App struct {
	httpServer *http.Server

	authUC  auth.UseCase
	groupUC group.UseCase
}

func NewApp() *App {
	db := initDB()

	userRepo := authgorm.NewUserRepository(db)
	groupRepo := groupgorm.NewGroupRepository(db)

	return &App{
		authUC: usecase.NewAuthUseCase(
			userRepo,
			viper.GetString("auth.hash_salt"),
			[]byte(viper.GetString("auth.signing_key")),
			viper.GetDuration("auth.token_ttl"),
		),
		groupUC: groupusecase.NewGroupUseCase(
			groupRepo),
	}
}

func (a *App) Run(port string) error {
	// Init gin handler
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	// Set up http handlers
	// SignUp/SignIn endpoints
	authhttp.RegisterHTTPEndpoints(router, a.authUC)
	grouphttp.RegisterHTTPEndpoints(router, a.groupUC)

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
	user := viper.GetString("mysql.user")
	pass := viper.GetString("mysql.password")
	host := viper.GetString("mysql.uri")
	dbname := viper.GetString("mysql.name")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, dbname)

	client, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error occured while establishing connection to gorm")
	}

	return client
}
