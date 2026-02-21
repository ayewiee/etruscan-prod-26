package main

import (
	"context"
	"errors"
	"etruscan/internal/app"
	"etruscan/internal/infrastructure/cache"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sethvargo/go-envconfig"
)

func main() {
	// context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var cfg app.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	appInstance, err := app.NewApiApp(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer appInstance.DBPool.Close()

	// goland wanted me to handle this error
	defer func(RedisClient *cache.Client) {
		err = RedisClient.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(appInstance.RedisClient)

	// run the server
	go func() {
		err = appInstance.Echo.Start(fmt.Sprintf(":%d", cfg.HttpPort))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := appInstance.Echo.Shutdown(ctx); err != nil {
		appInstance.Echo.Logger.Fatal(err)
	}
}
