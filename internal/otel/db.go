package otel

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

// otelMySQLDriverName is the database/sql driver name registered for
// instrumented MySQL connections. gorm opens DSNs with this prefix to obtain
// wrapped connections.
const otelMySQLDriverName = "otel-mysql"

// Raw DB semantic-convention attribute keys. semconv v1.43.0 removed the
// classic db.* helpers, so they are defined inline here.
const (
	dbSystemKey    = attribute.Key("db.system")
	dbNameKey      = attribute.Key("db.name")
	dbStatementKey = attribute.Key("db.statement")
)

// ctxStartKey carries the query start time so record() can measure duration.
type ctxStartKey struct{}

func withStart(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, ctxStartKey{}, t)
}

func startOf(ctx context.Context) time.Time {
	if t, ok := ctx.Value(ctxStartKey{}).(time.Time); ok {
		return t
	}
	return time.Now()
}

// otelDSN prefixes a MySQL DSN so database/sql selects the registered
// instrumented driver.
func otelDSN(dsn string) string {
	return otelMySQLDriverName + "://" + dsn
}

// underlyingConn is the set of database/sql interfaces the MySQL connection
// implements. Embedding it lets instrumentedConn forward every method except
// the two we wrap. driver.ConnContext is intentionally excluded: this Go build
// omits it, so we rely on QueryerContext/ExecerContext instead.
type underlyingConn interface {
	driver.Conn
	driver.QueryerContext
	driver.ExecerContext
	driver.NamedValueChecker
}

// instrumentedDriver is a database/sql driver.DriverContext that wraps MySQL
// connections to emit per-query spans and metrics.
type instrumentedDriver struct {
	name   string
	dbName string
	system string
	tracer trace.Tracer
	meter  meterHandle

	queries  metric.Int64Counter
	queryDur metric.Int64Histogram
}

// OpenConnector implements driver.DriverContext. It delegates to the real MySQL
// connector and wraps the resulting connections.
func (d *instrumentedDriver) OpenConnector(dsn string) (driver.Connector, error) {
	dsn = strings.TrimPrefix(dsn, d.name+"://")
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	inner, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, err
	}
	return &instrumentedConnector{inner: inner, driver: d}, nil
}

// Open implements driver.Driver. database/sql calls this when the driver does
// not expose OpenConnector; it delegates to OpenConnector for consistency.
func (d *instrumentedDriver) Open(name string) (driver.Conn, error) {
	connector, err := d.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return connector.Connect(context.Background())
}

// instrumentedConnector wraps a driver.Connector so every connection it hands
// out is instrumented.
type instrumentedConnector struct {
	inner  driver.Connector
	driver *instrumentedDriver
}

func (c *instrumentedConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.inner.Connect(ctx)
	if err != nil {
		return nil, err
	}
	uc, ok := conn.(underlyingConn)
	if !ok {
		return nil, fmt.Errorf("driver %q does not implement the required connection interfaces", c.driver.name)
	}
	return &instrumentedConn{driver: c.driver, underlyingConn: uc}, nil
}

func (c *instrumentedConnector) Driver() driver.Driver { return c.driver }

// instrumentedConn wraps a MySQL connection, tracing and metering queries.
type instrumentedConn struct {
	driver *instrumentedDriver
	underlyingConn
}

func (c *instrumentedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	ctx, span := c.driver.span(ctx, query)
	rows, err := c.underlyingConn.QueryContext(ctx, query, args)
	c.driver.record(ctx, query, err)
	span.End()
	return rows, err
}

func (c *instrumentedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ctx, span := c.driver.span(ctx, query)
	res, err := c.underlyingConn.ExecContext(ctx, query, args)
	c.driver.record(ctx, query, err)
	span.End()
	return res, err
}

// span starts a DB span with the standard attributes and marks it successful.
// It returns the (possibly new) context carrying the query start time.
func (d *instrumentedDriver) span(ctx context.Context, query string) (context.Context, trace.Span) {
	ctx = withStart(ctx, time.Now())
	_, span := d.tracer.Start(ctx, query)
	span.SetAttributes(
		dbSystemKey.String(d.system),
		dbNameKey.String(d.dbName),
		dbStatementKey.String(query),
		semconv.OTelStatusCodeOk,
	)
	return ctx, span
}

// record updates the span status and increments the query counter.
func (d *instrumentedDriver) record(ctx context.Context, query string, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil && err != io.EOF {
		span.SetAttributes(semconv.OTelStatusCodeError)
		span.RecordError(err)
		span.SetStatus(1, err.Error())
	}
	d.queries.Add(ctx, 1)
	d.queryDur.Record(ctx, time.Since(startOf(ctx)).Milliseconds())
}

// registerOnce ensures the instrumented driver is registered exactly once per
// process (database/sql.Register panics on double registration).
var registerOnce sync.Once

// RegisterInstrumentedMySQL registers the instrumented MySQL driver under
// otelMySQLDriverName. Safe to call multiple times; only the first takes effect.
func RegisterInstrumentedMySQL(tracer trace.Tracer, meter meterHandle, dbName string) {
	registerOnce.Do(func() {
		queries, err := meter.Int64Counter("db.client.queries")
		if err != nil {
			panic(err)
		}
		queryDur, err := meter.Int64Histogram("db.client.query.duration")
		if err != nil {
			panic(err)
		}
		sql.Register(otelMySQLDriverName, &instrumentedDriver{
			name:     otelMySQLDriverName,
			dbName:   dbName,
			system:   "mysql",
			tracer:   tracer,
			meter:    meter,
			queries:  queries,
			queryDur: queryDur,
		})
	})
}
