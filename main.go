package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	connectDb()

	e := echo.New()
	e.Use(addBasicAuthMiddleware())
	registerRoutes(e)
	startServer(e)
}

func connectDb() {
	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal("Cannot connect to database.", err)
	}

	defer conn.Close()
}

func addBasicAuthMiddleware() echo.MiddlewareFunc {
	authFn := func(username, password string, ctx echo.Context) (bool, error) {
		adminUser, adminPassword := prepareAdminUserPas(os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD"))

		return username == adminUser && password == adminPassword, nil
	}

	return middleware.BasicAuth(authFn)
}

func prepareAdminUserPas(username, password string) (adminUsername, adminPassword string) {
	if username == "" && password == "" {
		return "adminTax", "admin!"
	}

	return username, password
}

func registerRoutes(e *echo.Echo) {
	e.GET("/", handleRoot)
}

func handleRoot(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, Go Bootcamp!")
}

func startServer(e *echo.Echo) {
	go func() {
		if err := e.Start(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown
	log.Printf("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdonw failed: %v", err)
	}

	log.Println("Shutdown completed")
}
