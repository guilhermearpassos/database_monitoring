package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	config2 "github.com/guilhermearpassos/database-monitoring/internal/common/config"
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	_ "github.com/guilhermearpassos/database-monitoring/internal/services/agent/adapters/metrics"
	"github.com/jmoiron/sqlx"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"sync"
)

var (
	LoadGenCmd = &cobra.Command{
		Use:     "load_gen",
		Short:   "run dbm load_gen",
		Long:    "run dbm load_gen",
		Aliases: []string{},
		Example: "dbm load_gen --config=local/agent.toml",
		RunE:    StartLoadGen,
	}
)

func init() {
	LoadGenCmd.Flags().StringVar(&configFileName, "config", "local/agent.toml", "--config=local/agent.toml")
}

func StartLoadGen(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	var config config2.AgentConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		panic(fmt.Errorf("config file does not exist: %s", configFileName))
	}
	if _, err := toml.DecodeFile(configFileName, &config); err != nil {
		panic(fmt.Errorf("failed to parse config file: %s", err))
	}
	err := telemetry.InitTelemetryFromConfig(config.Telemetry)
	if err != nil {
		panic(fmt.Errorf("failed to init telemetry: %v", err))
	}
	wg := sync.WaitGroup{}
	for _, tgt := range config.TargetHosts {
		db, err := telemetry.OpenInstrumentedDB(tgt.Driver, tgt.ConnString)
		if err != nil {
			panic(fmt.Errorf("error connecting to %s: %w", tgt.Alias, err))
		}
		wg.Add(1)
		go generateLoad(ctx, db, &wg)
	}
	wg.Wait()
	return nil
}

func ensureTables(ctx context.Context, db *sqlx.DB) {
	q := `if not exists(select 1
              from sys.tables with (rowlock, updlock )
              where name = 'trades')
    begin
        create table trades
        (
            id       int identity primary key,
            strategy varchar(50)    not null,
            qty      decimal(24, 8) not null,
            price    decimal(24, 8) not null,
            account  varchar(50),
            asset    varchar(100)
        );
        create index trades_group_01_idx on trades (strategy, asset, account) include (qty);

    end`
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		panic(err)
	}
}

type Trade struct {
	Id       int
	Strategy string
	Qty      float64
	Price    float64
	Account  string
	Asset    string
}

func NewTrade() Trade {
	strategies := []string{"DayTrading", "SwingTrading", "ScalpTrading", "PositionTrading", "AlgoTrading"}
	accounts := []string{"ACC001", "ACC002", "ACC003", "ACC004", "ACC005"}
	assets := []string{"BTC", "ETH", "SOL", "ADA", "DOT", "AVAX", "MATIC"}

	return Trade{
		Strategy: strategies[rand.Intn(len(strategies))],
		Qty:      float64(rand.Intn(1000) + 1),
		Price:    float64(rand.Intn(990) + 10),
		Account:  accounts[rand.Intn(len(accounts))],
		Asset:    assets[rand.Intn(len(assets))],
	}
}

func generateReadLoad(ctx context.Context, db *sqlx.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}
		q := `select strategy, account, asset, sum(qty) as qty from trades
group by strategy, account, asset
having sum(qty) <> 0`
		_ = db.QueryRowContext(ctx, q)
	}
}

func generateWriteLoad(ctx context.Context, db *sqlx.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}
		_ = bulkInsert(ctx, db)

	}
}

func bulkInsert(ctx context.Context, db *sqlx.DB) (err error) {
	txx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = txx.Rollback()
		}
		err = txx.Commit()
	}()
	q := mssql.CopyIn("trades", mssql.BulkOptions{}, "strategy", "qty", "price", "account", "asset")
	stmt, err := txx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for i := 0; i < 1000; i++ {
		trade := NewTrade()
		_, err = stmt.Exec(trade.Strategy, trade.Qty, trade.Price, trade.Account, trade.Asset)
		if err != nil {
			return err
		}
	}
	_, err = stmt.Exec()
	return err
}

func generateLoad(ctx context.Context, db *sqlx.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	ensureTables(ctx, db)
	wg.Add(20)
	for i := 0; i < 10; i++ {
		go generateReadLoad(ctx, db, wg)
		go generateWriteLoad(ctx, db, wg)
	}
}
