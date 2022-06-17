package main

import (
	"context"
	"fmt"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/util/random"
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
		"seed",
		&cfg,
	)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err = seedWarehouses(context.Background(), cfg); err != nil {
		log.Fatalf("seed warehouses: %v", err)
	}
}

func seedWarehouses(ctx context.Context, cfg Config) error {
	dbSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.WarehousesDB.Host, cfg.WarehousesDB.Port, cfg.WarehousesDB.User, cfg.WarehousesDB.Password, cfg.WarehousesDB.Name, cfg.WarehousesDB.SSLMode,
	)

	db, err := util.OpenDB(dbSource)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}

	query := "INSERT INTO quantities (product_id, quantity) VALUES ($1, $2) ON CONFLICT (product_id) DO NOTHING"

	ids := make(map[int64]struct{}, 50)

	for i := 0; i < 50; i++ {
		var id int64

		for {
			id = random.From1To1000()
			if _, ok := ids[id]; !ok {
				ids[id] = struct{}{}
				break
			}
		}

		qnt := random.From0To1000()
		_, err := db.Exec(ctx, query, id, qnt)
		if err != nil {
			return fmt.Errorf("db exec: %w", err)
		}
	}

	return nil
}
