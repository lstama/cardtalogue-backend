package otel

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// InstrumentMySQLConn registers the instrumented MySQL driver (idempotent per
// process) and opens an instrumented *sql.DB suitable for use as a gorm.ConnPool.
// The returned pool routes every query through the span-recording wrapper,
// giving DB spans and db.client.calls metrics without changing gorm usage.
func InstrumentMySQLConn(tracer trace.Tracer, meter meterHandle, dsn string) (gorm.ConnPool, error) {
	dbName := mysqlDBName(dsn)
	RegisterInstrumentedMySQL(tracer, meter, dbName)
	sqlDB, err := sql.Open(otelMySQLDriverName, otelDSN(dsn))
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}

// mysqlDBName extracts the database name from a MySQL DSN for span attributes,
// falling back to "mysql" when the DSN cannot be parsed.
func mysqlDBName(dsn string) string {
	if cfg, err := mysql.ParseDSN(dsn); err == nil && cfg.DBName != "" {
		return cfg.DBName
	}
	return "mysql"
}
