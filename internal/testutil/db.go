package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	"github.com/shuvo-paul/uptimebot/internal/database/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	var container testcontainers.Container
	var err error

	once.Do(func() {
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        "postgres:15",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "secret",
				"POSTGRES_USER":     "test",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		}

		container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			t.Fatalf("failed to start container: %v", err)
		}

		host, _ := container.Host(ctx)
		port, _ := container.MappedPort(ctx, "5432")

		dsn := fmt.Sprintf("host=%s port=%s user=test password=secret dbname=testdb sslmode=disable", host, port.Port())
		dbInstance, err = sql.Open("postgres", dsn)
		if err != nil {
			t.Fatalf("failed to open DB: %v", err)
		}

		if err = dbInstance.Ping(); err != nil {
			t.Fatalf("failed to ping DB: %v", err)
		}

		err = migrations.SetupMigration(dbInstance)
		if err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}
	})

	// Register cleanup for each test (but don't close the shared DB or container)
	t.Cleanup(func() {
		// Optionally clean up DB state here, e.g.:
		// dbInstance.Exec("TRUNCATE TABLE users, sessions RESTART IDENTITY CASCADE")

	})

	return dbInstance
}

// GetTestTx returns a new transaction for testing
func GetTestTx(t *testing.T) *sql.Tx {
	db := SetupTestDB(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
