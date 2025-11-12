package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
)

type Config struct {
	Token string `env:"MAX_TOKEN"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := godotenv.Load("../.env"); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to get .env %v", err)
		}
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to get cfg %v", err)
	}

	api, err := maxbot.New(cfg.Token)
	if err != nil {
		log.Fatalf("Failed to create api %v", err)
	}

	_, err = api.Bots.GetBot(ctx)
	if err != nil {
		log.Fatalf("Failed to create bot %v", err)
	}

	go func() {
		for upd := range api.GetUpdates(ctx) {

			_, err := api.Messages.Send(ctx, maxbot.NewMessage().
				SetUser(upd.GetUserID()).
				SetText(fmt.Sprintf("Ваш user_id **%d**", upd.GetUserID())).SetFormat("markdown"))
			if err != nil && err.Error() != "" {
				log.Printf("Failed to send message %v", err)
			}

		}
	}()
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGTERM, os.Interrupt)
		select {
		case <-exit:
			cancel()
		case <-ctx.Done():
			return
		}
	}()
	<-ctx.Done()
}
