package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jinzhu/gorm"

	"github.com/maxp36/photogalery/internal/photogalery"
)

func main() {

	os.MkdirAll("./tmp", os.ModePerm)
	db, err := gorm.Open("sqlite3", "./tmp/test.db")
	if err != nil {
		log.Fatalf("failed to connect database: %s\n", err)
	}
	defer db.Close()

	phs := photogalery.NewService(db)

	router := photogalery.MakeHandler(phs)

	srv := &http.Server{
		Addr:    ":8181",
		Handler: router,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}
