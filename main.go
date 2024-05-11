// main function of the example application
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	// "os"
	// "os/signal"
	// "syscall"
	"time"

	gconfig "github.com/tinkerbaj/gintemp/config"
	gdatabase "github.com/tinkerbaj/gintemp/database"

	"github.com/tinkerbaj/gintemp/database/migrate"
	"github.com/tinkerbaj/gintemp/router"
)

func main() {
	// set configs
	err := gconfig.Config()
	if err != nil {
		fmt.Println(err)
		return
	}

	// read configs
	configure := gconfig.GetConfig()

	if gconfig.IsRDBMS() {
		// Initialize RDBMS client
		if err := gdatabase.InitDB().Error; err != nil {
			fmt.Println(err)
			return
		}

		// Drop all tables from DB
		/*
			if err := migrate.DropAllTables(); err != nil {
				fmt.Println(err)
				return
			}
		*/

		// Start DB migration
		if err := migrate.StartMigration(*configure); err != nil {
			fmt.Println(err)
			return
		}

		// Manually set foreign key for MySQL and PostgreSQL
		if err := migrate.SetPkFk(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if gconfig.IsRedis() {
		// Initialize REDIS client
		if _, err := gdatabase.InitRedis(); err != nil {
			fmt.Println(err)
			return
		}
	}

	r, err := router.SetupRouter(configure)
	if err != nil {
		fmt.Println(err)
		return
	}

	// srv := &http.Server{
	// 	Addr:    configure.Server.ServerHost + ":" + configure.Server.ServerPort,
	// 	Handler: r,
	// }

	// srvErrs := make(chan error, 1)
	// go func() {
	// 	srvErrs <- srv.ListenAndServe()
	// }()

	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// shutdown := gracefulShutdown(srv)

	// select {
	// case err := <-srvErrs:
	// 	shutdown(err)
	// case sig := <-quit:
	// 	shutdown(sig)
	// }
	// Attaches the router to a http.Server and starts listening and serving HTTP requests
	err = r.Run(configure.Server.ServerHost + ":" + configure.Server.ServerPort)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func gracefulShutdown(srv *http.Server) func(reason interface{}) {
	return func(reason interface{}) {
		log.Println("Server Shutdown:", reason)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Error Gracefully Shutting Down API:", err)
		}
	}
}
