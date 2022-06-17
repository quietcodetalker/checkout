package main

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/app/stock"
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
		"stock",
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

	repo := stock.NewPgRepo(db)

	reservedOrdersProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "reserved_orders")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	resetProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "reset")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	kafkaClient := stock.NewKafkaClient(reservedOrdersProducer, resetProducer)

	svc := stock.NewService(repo, kafkaClient)

	hdl := stock.NewKafkaHandler(svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := kafka.NewSaramaConsumer(
		ctx,
		cfg.Brokers,
		[]string{"saved_orders", "reset", "cancel", "paid_orders"},
		"stock",
		hdl,
	)
	if err != nil {
		log.Fatalf("init consumer err: %v", err)
	}

	<-ctx.Done()
	consumer.Close()
}
