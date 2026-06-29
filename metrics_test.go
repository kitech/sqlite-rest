package main

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestMetricsServer_monitorDatabaseSize(t *testing.T) {
	t.Parallel()

	tc := createTestContextWithHMACTokenAuth(t)
	defer tc.CleanUp(t)

	tc.ExecuteSQL(t, "CREATE TABLE test (id int, s text)")
	tc.ExecuteSQL(t, `INSERT INTO test (id, s) VALUES (1, "a"), (1, "a"), (1, "a")`)

	metricsServer, err := NewMetricsServer(MetricsServerOptions{
		Logger:   createTestLogger(t).WithName("test"),
		Addr:     ":8081",
		Registry: NewDatabaseRegistry(map[string]*sqlx.DB{testDBName: tc.DB()}),
	})
	assert.NoError(t, err)

	done := make(chan struct{})

	go func() {
		metricsServer.monitorDatabaseSizes(done)
	}()

	time.Sleep(100 * time.Millisecond)
	close(done)
}
