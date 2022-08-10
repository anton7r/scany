package pgxscan

import (
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/anton7r/scany/dbscan"
)

// Querier is something that pgxscan can query and get the pgx.Rows from.
// For example, it can be: *pgxpool.Pool, *pgx.Conn or pgx.Tx.
type Querier interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

var (
	_ Querier = &pgxpool.Pool{}
	_ Querier = &pgx.Conn{}
	_ Querier = pgx.Tx(nil)
)

// Select is a package-level helper function that uses the DefaultAPI object.
// See API.Select for details.
func Select(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	return errors.WithStack(DefaultAPI.Select(ctx, db, dst, query, args...))
}

// Get is a package-level helper function that uses the DefaultAPI object.
// See API.Get for details.
func Get(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	return errors.WithStack(DefaultAPI.Get(ctx, db, dst, query, args...))
}

// Exec is a package-level helper function that uses the DefaultAPI object.
// See API.Exec for details.
func Exec(ctx context.Context, db Querier, query string, args ...interface{}) (pgconn.CommandTag, error) {
	commandTag, err := DefaultAPI.Exec(ctx, db, query, args...)
	return commandTag, errors.WithStack(err)
}

// ScanAll is a package-level helper function that uses the DefaultAPI object.
// See API.ScanAll for details.
func ScanAll(dst interface{}, rows pgx.Rows) error {
	return errors.WithStack(DefaultAPI.ScanAll(dst, rows))
}

// ScanOne is a package-level helper function that uses the DefaultAPI object.
// See API.ScanOne for details.
func ScanOne(dst interface{}, rows pgx.Rows) error {
	return errors.WithStack(DefaultAPI.ScanOne(dst, rows))
}

// RowScanner is a wrapper around the dbscan.RowScanner type.
// See dbscan.RowScanner for details.
type RowScanner struct {
	*dbscan.RowScanner
}

// NewRowScanner is a package-level helper function that uses the DefaultAPI object.
// See API.NewRowScanner for details.
func NewRowScanner(rows pgx.Rows) *RowScanner {
	return DefaultAPI.NewRowScanner(rows)
}

// ScanRow is a package-level helper function that uses the DefaultAPI object.
// See API.ScanRow for details.
func ScanRow(dst interface{}, rows pgx.Rows) error {
	return DefaultAPI.ScanRow(dst, rows)
}

// NewDBScanAPI creates a new dbscan API object with default configuration settings for pgxscan.
func NewDBScanAPI(opts ...dbscan.APIOption) (*dbscan.API, error) {
	defaultOpts := []dbscan.APIOption{
		dbscan.WithScannableTypes(
			(*sql.Scanner)(nil),
			(*pgtype.BinaryDecoder)(nil),
			(*pgtype.TextDecoder)(nil),
		),
	}
	opts = append(defaultOpts, opts...)
	api, err := dbscan.NewAPI(opts...)
	return api, errors.WithStack(err)
}

// API is a wrapper around the dbscan.API type.
// See dbscan.API for details.
type API struct {
	dbscanAPI *dbscan.API
}

// NewAPI creates new API instance from dbscan.API instance.
func NewAPI(dbscanAPI *dbscan.API) (*API, error) {
	api := &API{
		dbscanAPI: dbscanAPI,
	}
	return api, nil
}

// Select is a high-level function that queries rows from Querier and calls the ScanAll function.
// See ScanAll for details.
func (api *API) Select(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "scany: query multiple result rows")
	}
	err = api.ScanAll(dst, rows)
	return errors.WithStack(err)
}

// SelectNamed is a high-level function that queries rows from Querier and calls the ScanAll function.
// See ScanAll for details.
func (api *API) SelectNamed(ctx context.Context, db Querier, dst interface{}, query string, arg interface{}) error {
	compiledQuery, args, err := api.dbscanAPI.NamedQueryParams(query, arg)
	if err != nil {
		return err
	}

	return api.Select(ctx, db, dst, compiledQuery, args)
}

// Get is a high-level function that queries rows from Querier and calls the ScanOne function.
// See ScanOne for details.
func (api *API) Get(ctx context.Context, db Querier, dst interface{}, query string, args ...interface{}) error {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "scany: query one result row")
	}
	err = api.ScanOne(dst, rows)
	return errors.WithStack(err)
}

// GetNamed is a high-level function that queries rows from Querier and calls the ScanOne function.
// See ScanOne for details.
func (api *API) GetNamed(ctx context.Context, db Querier, dst interface{}, query string, arg interface{}) error {
	compiledQuery, args, err := api.dbscanAPI.NamedQueryParams(query, arg)
	if err != nil {
		return err
	}

	return api.Get(ctx, db, dst, compiledQuery, args)
}

// Exec is a high-level function that sends an executable action to the database
func (api *API) Exec(ctx context.Context, db Querier, query string, args ...interface{}) (pgconn.CommandTag, error) {
	tag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "scany: exec")
	}

	return tag, nil
}

// ExecNamed is a high-level function that sends an executable action to the database with named parameters
func (api *API) ExecNamed(ctx context.Context, db Querier, query string, arg interface{}) (pgconn.CommandTag, error) {
	compiledQuery, args, err := api.dbscanAPI.NamedQueryParams(query, arg)
	if err != nil {
		return nil, err
	}

	return api.Exec(ctx, db, compiledQuery, args)
}

// QueryNamed is a high-level function that is used to retrieve pgx.Rows from the database with named parameters
func (api *API) QueryNamed(ctx context.Context, db Querier, query string, arg interface{}) (pgx.Rows, error) {
	compiledQuery, args, err := api.dbscanAPI.NamedQueryParams(query, arg)
	if err != nil {
		return nil, err
	}

	return db.Query(ctx, compiledQuery, args)
}

// ScanAll is a wrapper around the dbscan.ScanAll function.
// See dbscan.ScanAll for details.
func (api *API) ScanAll(dst interface{}, rows pgx.Rows) error {
	err := api.dbscanAPI.ScanAll(dst, NewRowsAdapter(rows))
	return errors.WithStack(err)
}

// ScanOne is a wrapper around the dbscan.ScanOne function.
// See dbscan.ScanOne for details. If no rows are found it
// returns a pgx.ErrNoRows error.
func (api *API) ScanOne(dst interface{}, rows pgx.Rows) error {
	err := api.dbscanAPI.ScanOne(dst, NewRowsAdapter(rows))
	if dbscan.NotFound(err) {
		return errors.WithStack(pgx.ErrNoRows)
	}
	return errors.WithStack(err)
}

type PreparedQuery struct {
	api  *API
	prep *dbscan.PreparedQuery
}

func (api *API) PrepareNamed(query string, assertableStruct ...interface{}) (*PreparedQuery, error) {
	dbPrep, err := api.dbscanAPI.PrepareNamed(query, assertableStruct...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &PreparedQuery{api, dbPrep}, nil
}

// SelectNamed is a high-level function that queries rows from Querier and calls the ScanAll function.
// See ScanAll for details.
func (pq *PreparedQuery) SelectNamed(ctx context.Context, db Querier, dst interface{}, arg interface{}) error {
	query, args, err := pq.prep.GetQuery(arg)
	if err != nil {
		return err
	}

	return pq.api.Select(ctx, db, dst, query, args)
}

// GetNamed is a high-level function that queries rows from Querier and calls the ScanOne function.
// See ScanOne for details.
func (pq *PreparedQuery) GetNamed(ctx context.Context, db Querier, dst interface{}, arg interface{}) error {
	query, args, err := pq.prep.GetQuery(arg)
	if err != nil {
		return err
	}

	return pq.api.Get(ctx, db, dst, query, args)
}

// ExecNamed is a high-level function that sends an executable action to the database with named parameters
func (pq *PreparedQuery) ExecNamed(ctx context.Context, db Querier, arg interface{}) (pgconn.CommandTag, error) {
	query, args, err := pq.prep.GetQuery(arg)
	if err != nil {
		return nil, err
	}

	return pq.api.Exec(ctx, db, query, args)
}

// QueryNamed is a high-level function that is used to retrieve *sql.Rows from the database with named parameters
func (pq *PreparedQuery) QueryNamed(ctx context.Context, db Querier, arg interface{}) (pgx.Rows, error) {
	query, args, err := pq.prep.GetQuery(arg)
	if err != nil {
		return nil, err
	}

	return db.Query(ctx, query, args)
}

// NotFound is a helper function to check if an error
// is `pgx.ErrNoRows`.
func NotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// NewRowScanner returns a new RowScanner instance.
func (api *API) NewRowScanner(rows pgx.Rows) *RowScanner {
	ra := NewRowsAdapter(rows)
	return &RowScanner{RowScanner: api.dbscanAPI.NewRowScanner(ra)}
}

// ScanRow is a wrapper around the dbscan.ScanRow function.
// See dbscan.ScanRow for details.
func (api *API) ScanRow(dst interface{}, rows pgx.Rows) error {
	err := api.dbscanAPI.ScanRow(dst, NewRowsAdapter(rows))
	return errors.WithStack(err)
}

// RowsAdapter makes pgx.Rows compliant with the dbscan.Rows interface.
// See dbscan.Rows for details.
type RowsAdapter struct {
	pgx.Rows
}

// NewRowsAdapter returns a new RowsAdapter instance.
func NewRowsAdapter(rows pgx.Rows) *RowsAdapter {
	return &RowsAdapter{Rows: rows}
}

// Columns implements the dbscan.Rows.Columns method.
func (ra RowsAdapter) Columns() ([]string, error) {
	columns := make([]string, len(ra.Rows.FieldDescriptions()))
	for i, fd := range ra.Rows.FieldDescriptions() {
		columns[i] = string(fd.Name)
	}
	return columns, nil
}

// Close implements the dbscan.Rows.Close method.
func (ra RowsAdapter) Close() error {
	ra.Rows.Close()
	return nil
}

func mustNewDBScanAPI(opts ...dbscan.APIOption) *dbscan.API {
	api, err := NewDBScanAPI(opts...)
	if err != nil {
		panic(err)
	}
	return api
}

func mustNewAPI(dbscanAPI *dbscan.API) *API {
	api, err := NewAPI(dbscanAPI)
	if err != nil {
		panic(err)
	}
	return api
}

// DefaultAPI is the default instance of API with all configuration settings set to default.
var DefaultAPI = mustNewAPI(mustNewDBScanAPI())
