package main

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/app/notification"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util"
	"log"
	"path"
	"runtime"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := path.Join(path.Dir(filename), "../..")

	var cfg Config

	err := util.LoadConfig(
		path.Join(rootDir, "configs"),
		"notifications",
		&cfg,
	)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	db, err := util.OpenDB(dbSource)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	repo := notification.NewPgRepo(db)

	emailNotificationsProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "email_notifications")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	kafkaClient := notification.NewKafkaClient(emailNotificationsProducer)

	svc := notification.NewService(repo, kafkaClient)

	hdl := notification.NewKafkaHandler(svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := kafka.NewSaramaConsumer(
		ctx,
		cfg.Brokers,
		[]string{"paid_orders", "check"},
		"orders",
		hdl,
	)
	if err != nil {
		log.Fatalf("init consumer err: %v", err)
	}

	<-ctx.Done()
	consumer.Close()
}
