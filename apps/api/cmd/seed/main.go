package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"exam-results-platform/api/internal/config"
	"exam-results-platform/api/internal/seed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	total := flag.Int("total", 10_000_000, "number of synthetic results to insert")
	batchSize := flag.Int("batch-size", 10_000, "rows per COPY batch")
	truncate := flag.Bool("truncate", false, "delete existing synthetic rows before loading")
	flag.Parse()

	if *total < 1 || *total > 2_176_782_336 || *batchSize < 1 {
		fmt.Fprintln(os.Stderr, "invalid seed arguments")
		os.Exit(2)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("seed configuration is invalid")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Hour)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("seed database configuration is invalid")
		os.Exit(1)
	}
	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("seed database is unavailable")
		os.Exit(1)
	}
	defer pool.Close()

	if *truncate {
		if _, err := pool.Exec(ctx, "TRUNCATE exam_results"); err != nil {
			logger.Error("could not clear synthetic data")
			os.Exit(1)
		}
	}

	columns := seed.ColumnNames()
	for start := 0; start < *total; start += *batchSize {
		end := min(start+*batchSize, *total)
		rows := make([][]any, 0, end-start)
		for index := start; index < end; index++ {
			rows = append(rows, seed.RowForIndex(index))
		}
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"exam_results"}, columns, pgx.CopyFromRows(rows)); err != nil {
			logger.Error("synthetic batch could not be inserted", "from", start, "to", end)
			os.Exit(1)
		}
		logger.Info("synthetic batch inserted", "inserted", end, "total", *total)
	}

	if _, err := pool.Exec(ctx, "ANALYZE exam_results"); err != nil {
		logger.Error("table statistics could not be refreshed")
		os.Exit(1)
	}
	logger.Info("synthetic seed completed", "total", *total)
}
