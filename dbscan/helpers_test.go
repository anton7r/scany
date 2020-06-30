package dbscan_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbscan2 "github.com/georgysavva/dbscan/dbscan"
	"github.com/georgysavva/dbscan/pgxscan"
)

func makeStrPtr(v string) *string { return &v }

func queryRows(t *testing.T, query string) dbscan2.Rows {
	t.Helper()
	pgxRows, err := testDB.Query(ctx, query)
	require.NoError(t, err)
	rows := pgxscan.NewRowsAdapter(pgxRows)
	return rows
}

func scan(t *testing.T, dst interface{}, rows dbscan2.Rows) error {
	defer rows.Close()
	rs := dbscan2.NewRowScanner(rows)
	rows.Next()
	if err := rs.Scan(dst); err != nil {
		return err
	}
	requireNoRowsErrorsAndClose(t, rows)
	return nil
}

func requireNoRowsErrorsAndClose(t *testing.T, rows dbscan2.Rows) {
	t.Helper()
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
}

func allocateDestination(v interface{}) interface{} {
	dstType := reflect.TypeOf(v)
	dst := reflect.New(dstType).Interface()
	return dst
}

func assertDestinationEqual(t *testing.T, expected interface{}, dst interface{}) {
	t.Helper()
	got := reflect.ValueOf(dst).Elem().Interface()
	assert.Equal(t, expected, got)
}