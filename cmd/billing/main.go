package main

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/app/billing"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/cache"
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
		"billing",
		&cfg,
	)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbMasterSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Master.Host, cfg.DB.Master.Port, cfg.DB.Master.User, cfg.DB.Master.Password, cfg.DB.Master.Name, cfg.DB.Master.SSLMode,
	)

	dbMaster, err := util.OpenDB(dbMasterSource)
	if err != nil {
		log.Fatalf("failed to open dbMaster: %v", err)
	}

	dbReplicaSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Replica.Host, cfg.DB.Replica.Port, cfg.DB.Replica.User, cfg.DB.Replica.Password, cfg.DB.Replica.Name, cfg.DB.Replica.SSLMode,
	)

	dbReplica, err := util.OpenDB(dbReplicaSource)
	if err != nil {
		log.Fatalf("failed to open dbReplica: %v", err)
	}

	repo := billing.NewPgRepo(dbMaster, dbReplica)

	pendingPaymentsProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "pending_payments")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	paidPaymentsProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "paid_payments")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	resetProducer, err := kafka.NewSaramaProducer(cfg.Brokers, "reset")
	if err != nil {
		log.Fatalf("create sarama producer: %v", err)
	}

	kafkaClient := billing.NewKafkaClient(pendingPaymentsProducer, paidPaymentsProducer, resetProducer)

	cch := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)

	svc := billing.NewService(repo, kafkaClient, cch)

	hdl := billing.NewKafkaHandler(svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := kafka.NewSaramaConsumer(
		ctx,
		cfg.Brokers,
		[]string{"reserved_orders", "receipts", "reset", "cancel"},
		"billing",
		hdl,
	)
	if err != nil {
		log.Fatalf("init consumer err: %v", err)
	}

	<-ctx.Done()
	consumer.Close()
}
