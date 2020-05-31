package pgxscan_test

import (
	"context"
	"testing"

	"github.com/georgysavva/pgxscan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryAll_Succeeds(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	q := &fakeQuerier{rows: rows}
	type dst struct {
		Foo string
	}
	var got []dst
	err := pgxscan.QueryAll(context.Background(), q, &got, "" /* sql */)
	require.NoError(t, err)
	expected := []dst{{Foo: "foo val"}, {Foo: "foo val 2"}, {Foo: "foo val 3"}}
	assert.Equal(t, expected, got)
}

func TestQueryOne_Succeeds(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
		},
	}
	q := &fakeQuerier{rows: rows}
	type dst struct {
		Foo string
	}
	var got dst
	err := pgxscan.QueryOne(context.Background(), q, &got, "" /* sql */)
	require.NoError(t, err)
	expected := dst{Foo: "foo val"}
	assert.Equal(t, expected, got)
}

func TestScanAll_Succeeds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		rows     *fakeRows
		expected interface{}
	}{
		{
			name: "slice of structs",
			rows: &fakeRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []struct {
				Foo string
				Bar string
			}{
				{Foo: "foo val", Bar: "bar val"},
				{Foo: "foo val 2", Bar: "bar val 2"},
				{Foo: "foo val 3", Bar: "bar val 3"},
			},
		},
		{
			name: "slice of structs by ptr",
			rows: &fakeRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []*struct {
				Foo string
				Bar string
			}{
				{Foo: "foo val", Bar: "bar val"},
				{Foo: "foo val 2", Bar: "bar val 2"},
				{Foo: "foo val 3", Bar: "bar val 3"},
			},
		},
		{
			name: "slice of maps",
			rows: &fakeRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []map[string]interface{}{
				{"foo": "foo val", "bar": "bar val"},
				{"foo": "foo val 2", "bar": "bar val 2"},
				{"foo": "foo val 3", "bar": "bar val 3"},
			},
		},
		{
			name: "slice of maps by ptr",
			rows: &fakeRows{
				columns: []string{"foo", "bar"},
				data: [][]interface{}{
					{"foo val", "bar val"},
					{"foo val 2", "bar val 2"},
					{"foo val 3", "bar val 3"},
				},
			},
			expected: []*map[string]interface{}{
				{"foo": "foo val", "bar": "bar val"},
				{"foo": "foo val 2", "bar": "bar val 2"},
				{"foo": "foo val 3", "bar": "bar val 3"},
			},
		},
		{
			name: "slice of strings",
			rows: &fakeRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{"foo val"},
					{"foo val 2"},
					{"foo val 3"},
				},
			},
			expected: []string{"foo val", "foo val 2", "foo val 3"},
		},
		{
			name: "slice of strings by ptr",
			rows: &fakeRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{makeStrPtr("foo val")},
					{nil},
					{makeStrPtr("foo val 3")},
				},
			},
			expected: []*string{makeStrPtr("foo val"), nil, makeStrPtr("foo val 3")},
		},
		{
			name: "slice of slices",
			rows: &fakeRows{
				columns: []string{"foo"},
				data: [][]interface{}{
					{[]string{"foo val", "foo val 2"}},
					{[]string{"foo val 3", "foo val 4"}},
					{[]string{"foo val 5", "foo val 6"}},
				},
			},
			expected: [][]string{
				{"foo val", "foo val 2"},
				{"foo val 3", "foo val 4"},
				{"foo val 5", "foo val 6"},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dstVal := newDstValue(tc.expected)
			err := pgxscan.ScanAll(dstVal.Addr().Interface(), tc.rows)
			require.NoError(t, err)
			assertDstValueEqual(t, tc.expected, dstVal)
		})
	}
}

func TestScanAll_NonEmptySlice_ResetsDstSlice(t *testing.T) {
	t.Parallel()
	fr := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	expected := []string{"foo val", "foo val 2", "foo val 3"}
	got := []string{"junk data", "junk data 2"}
	err := pgxscan.ScanAll(&got, fr)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestScanAll_NonSliceDestination_ReturnsErr(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	var dst struct {
		Foo string
	}
	err := pgxscan.ScanAll(&dst, rows)
	expectedErr := "destination must be a slice, got: struct { Foo string }"
	assert.EqualError(t, err, expectedErr)
}

func TestScanOne_Succeeds(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
		},
	}
	type dst struct {
		Foo string
	}
	got := dst{}
	err := pgxscan.ScanOne(&got, rows)
	require.NoError(t, err)
	expected := dst{Foo: "foo val"}
	assert.Equal(t, expected, got)
}

func TestScanRow_Succeeds(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
		},
	}
	type dst struct {
		Foo string
	}
	rows.Next()
	var got dst
	err := pgxscan.ScanRow(&got, rows)
	require.NoError(t, err)
	expected := dst{Foo: "foo val"}
	assert.Equal(t, expected, got)
}

func TestScanOne_ZeroRows_ReturnsNotFoundErr(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data:    [][]interface{}{},
	}
	var dst string
	err := pgxscan.ScanOne(&dst, rows)
	assert.True(t, pgxscan.NotFound(err))
}

func TestScanOne_MultipleRows_ReturnsErr(t *testing.T) {
	t.Parallel()
	rows := &fakeRows{
		columns: []string{"foo"},
		data: [][]interface{}{
			{"foo val"},
			{"foo val 2"},
			{"foo val 3"},
		},
	}
	var dst string
	err := pgxscan.ScanOne(&dst, rows)
	expectedErr := "expected 1 row, got: 3"
	assert.EqualError(t, err, expectedErr)
}