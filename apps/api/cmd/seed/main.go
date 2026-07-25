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
	total := flag.Int("total", 10_000_000, "target number of synthetic results in the table")
	batchSize := flag.Int("batch-size", 10_000, "rows per COPY batch")
	truncate := flag.Bool("truncate", false, "DANGER: truncate the entire exam_results table before loading")
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

	conn, err := pool.Acquire(ctx)
	if err != nil {
		logger.Error("seed database connection is unavailable")
		os.Exit(1)
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtext('exam_results_synthetic_seed'))").Scan(&locked); err != nil || !locked {
		logger.Error("another synthetic seed process is already running")
		os.Exit(1)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtext('exam_results_synthetic_seed'))")
	}()

	var tableName *string
	if err := conn.QueryRow(ctx, "SELECT to_regclass('public.exam_results')::text").Scan(&tableName); err != nil || tableName == nil {
		logger.Error("exam_results table is missing; apply apps/api/migrations/0001_exam_results.sql first")
		os.Exit(1)
	}

	if *truncate {
		if _, err := conn.Exec(ctx, "TRUNCATE exam_results"); err != nil {
			logger.Error("could not truncate exam_results")
			os.Exit(1)
		}
	}

	var existing int
	var firstFIN string
	var lastFIN string
	if err := conn.QueryRow(ctx, `
		SELECT count(*), COALESCE(min(fin_code), ''), COALESCE(max(fin_code), '')
		FROM exam_results
		WHERE fin_code >= 'T000000' AND fin_code < 'U000000'
	`).Scan(&existing, &firstFIN, &lastFIN); err != nil {
		logger.Error("could not inspect existing synthetic results")
		os.Exit(1)
	}
	if existing > 0 && (firstFIN != seed.FINForIndex(0) || lastFIN != seed.FINForIndex(existing-1)) {
		logger.Error("existing synthetic results are not a resumable sequence; do not use --truncate unless a full wipe is intentional")
		os.Exit(1)
	}

	if existing >= *total {
		if _, err := conn.Exec(ctx, "ANALYZE exam_results"); err != nil {
			logger.Error("table statistics could not be refreshed")
			os.Exit(1)
		}
		logger.Info("synthetic seed already complete", "existing", existing, "target", *total)
		return
	}

	columns := seed.ColumnNames()
	logger.Info("synthetic seed started", "existing", existing, "target", *total, "batch_size", *batchSize)
	for start := existing; start < *total; start += *batchSize {
		end := min(start+*batchSize, *total)
		rows := make([][]any, 0, end-start)
		for index := start; index < end; index++ {
			rows = append(rows, seed.RowForIndex(index))
		}
		if _, err := conn.CopyFrom(ctx, pgx.Identifier{"exam_results"}, columns, pgx.CopyFromRows(rows)); err != nil {
			logger.Error("synthetic batch could not be inserted", "from", start, "to", end)
			os.Exit(1)
		}
		logger.Info("synthetic batch inserted", "inserted", end, "total", *total)
	}

	if _, err := conn.Exec(ctx, "ANALYZE exam_results"); err != nil {
		logger.Error("table statistics could not be refreshed")
		os.Exit(1)
	}
	logger.Info("synthetic seed completed", "total", *total)
}
