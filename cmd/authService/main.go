package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"nightcore-team/nightcore-auth-service-go/config"
	http_handlers "nightcore-team/nightcore-auth-service-go/internal/api/http/handlers"
	http_middlewares "nightcore-team/nightcore-auth-service-go/internal/api/http/middlewares"
	http_routes "nightcore-team/nightcore-auth-service-go/internal/api/http/routes"
	"nightcore-team/nightcore-auth-service-go/internal/infra/discord"
	rds "nightcore-team/nightcore-auth-service-go/internal/infra/redis"
	"nightcore-team/nightcore-auth-service-go/internal/repository"
	"nightcore-team/nightcore-auth-service-go/internal/services"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	app := gin.Default()
	app.Use(http_middlewares.CORSMiddleware())

	redisClient := rds.NewRedisClient()

	sessionRepo := repository.NewSessionRepository(redisClient)
	tokenService := services.NewTokenService()
	oauthProvider := discord.NewOauthProvider()

	oicService := services.NewOICService(sessionRepo, oauthProvider, tokenService)

	handler := http_handlers.NewAuthHandler(oicService)

	http_routes.AddAuthRoutes(&app.RouterGroup, handler)

	addr := fmt.Sprintf("%s:%d", config.API.API_HOST, config.API.API_PORT)
	srv := &http.Server{
		Addr:    addr,
		Handler: app,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}