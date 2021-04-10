package influxql_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/influxql"
)

// Ensure the parser can parse a multi-statement query.
func TestParser_ParseQuery(t *testing.T) {
	s := `SELECT a FROM b; SELECT c FROM d`
	q, err := influxql.NewParser(strings.NewReader(s)).ParseQuery()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if len(q.Statements) != 2 {
		t.Fatalf("unexpected statement count: %d", len(q.Statements))
	}
}

func TestParser_ParseQuery_TrailingSemicolon(t *testing.T) {
	s := `SELECT value FROM cpu;`
	q, err := influxql.NewParser(strings.NewReader(s)).ParseQuery()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if len(q.Statements) != 1 {
		t.Fatalf("unexpected statement count: %d", len(q.Statements))
	}
}

// Ensure the parser can parse an empty query.
func TestParser_ParseQuery_Empty(t *testing.T) {
	q, err := influxql.NewParser(strings.NewReader(``)).ParseQuery()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if len(q.Statements) != 0 {
		t.Fatalf("unexpected statement count: %d", len(q.Statements))
	}
}

// Ensure the parser will skip comments.
func TestParser_ParseQuery_SkipComments(t *testing.T) {
	q, err := influxql.ParseQuery(`SELECT * FROM cpu; -- read from cpu database

/* create continuous query */
CREATE CONTINUOUS QUERY cq0 ON db0 BEGIN
	SELECT mean(*) INTO db1..:MEASUREMENT FROM cpu GROUP BY time(5m)
END;

/* just a multline comment
what is this doing here?
**/

-- should ignore the trailing multiline comment /*
SELECT mean(value) FROM gpu;
-- trailing comment at the end`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	} else if len(q.Statements) != 3 {
		t.Fatalf("unexpected statement count: %d", len(q.Statements))
	}
}

// Ensure the parser can return an error from an malformed statement.
func TestParser_ParseQuery_ParseError(t *testing.T) {
	_, err := influxql.NewParser(strings.NewReader(`SELECT`)).ParseQuery()
	if err == nil || err.Error() != `found EOF, expected identifier, string, number, bool at line 1, char 8` {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestParser_ParseQuery_NoSemicolon(t *testing.T) {
	_, err := influxql.NewParser(strings.NewReader(`CREATE DATABASE foo CREATE DATABASE bar`)).ParseQuery()
	if err == nil || err.Error() != `found CREATE, expected ; at line 1, char 21` {
		t.Fatalf("unexpected error: %s", err)
	}
}

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	// For use in various tests.
	now := time.Now()

	var tests = []struct {
		skip   bool
		s      string
		params map[string]interface{}
		stmt   influxql.Statement
		err    string
	}{
		// SELECT * statement
		{
			s: `SELECT * FROM myseries`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.Wildcard{}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},
		{
			s: `SELECT * FROM myseries GROUP BY *`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.Wildcard{}},
				},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Wildcard{}}},
			},
		},
		{
			s: `SELECT field1, * FROM myseries GROUP BY *`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.VarRef{Val: "field1"}},
					{Expr: &influxql.Wildcard{}},
				},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Wildcard{}}},
			},
		},
		{
			s: `SELECT *, field1 FROM myseries GROUP BY *`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.Wildcard{}},
					{Expr: &influxql.VarRef{Val: "field1"}},
				},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Wildcard{}}},
			},
		},

		// SELECT statement
		{
			s: fmt.Sprintf(`SELECT mean(field1), sum(field2), count(field3) AS field_x FROM myseries WHERE host = 'hosta.influxdb.org' and time > '%s' GROUP BY time(10h) ORDER BY DESC LIMIT 20 OFFSET 10;`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "mean", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}},
					{Expr: &influxql.Call{Name: "sum", Args: []influxql.Expr{&influxql.VarRef{Val: "field2"}}}},
					{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.VarRef{Val: "field3"}}}, Alias: "field_x"},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Condition: &influxql.BinaryExpr{
					Op: influxql.AND,
					LHS: &influxql.BinaryExpr{
						Op:  influxql.EQ,
						LHS: &influxql.VarRef{Val: "host"},
						RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
					},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.GT,
						LHS: &influxql.VarRef{Val: "time"},
						RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
					},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 10 * time.Hour}}}}},
				SortFields: []*influxql.SortField{
					{Ascending: false},
				},
				Limit:  20,
				Offset: 10,
			},
		},
		{
			s: `SELECT "foo.bar.baz" AS foo FROM myseries`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.VarRef{Val: "foo.bar.baz"}, Alias: "foo"},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},
		{
			s: `SELECT "foo.bar.baz" AS foo FROM foo`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.VarRef{Val: "foo.bar.baz"}, Alias: "foo"},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "foo"}},
			},
		},

		// sample
		{
			s: `SELECT sample(field1, 100) FROM myseries;`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "sample", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.IntegerLiteral{Val: 100}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		// derivative
		{
			s: `SELECT derivative(field1, 1h) FROM myseries;`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "derivative", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.DurationLiteral{Val: time.Hour}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		{
			s: fmt.Sprintf(`SELECT derivative(field1, 1h) FROM myseries WHERE time > '%s'`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "derivative", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.DurationLiteral{Val: time.Hour}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		{
			s: `SELECT derivative(field1, 1h) / derivative(field2, 1h) FROM myseries`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.BinaryExpr{
							LHS: &influxql.Call{
								Name: "derivative",
								Args: []influxql.Expr{
									&influxql.VarRef{Val: "field1"},
									&influxql.DurationLiteral{Val: time.Hour},
								},
							},
							RHS: &influxql.Call{
								Name: "derivative",
								Args: []influxql.Expr{
									&influxql.VarRef{Val: "field2"},
									&influxql.DurationLiteral{Val: time.Hour},
								},
							},
							Op: influxql.DIV,
						},
					},
				},
				Sources: []influxql.Source{
					&influxql.Measurement{Name: "myseries"},
				},
			},
		},

		// difference
		{
			s: `SELECT difference(field1) FROM myseries;`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "difference", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		{
			s: fmt.Sprintf(`SELECT difference(max(field1)) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "difference",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "max",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// non_negative_difference
		{
			s: `SELECT non_negative_difference(field1) FROM myseries;`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "non_negative_difference", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		{
			s: fmt.Sprintf(`SELECT non_negative_difference(max(field1)) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "non_negative_difference",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "max",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// moving_average
		{
			s: `SELECT moving_average(field1, 3) FROM myseries;`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "moving_average", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.IntegerLiteral{Val: 3}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		{
			s: fmt.Sprintf(`SELECT moving_average(max(field1), 3) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "moving_average",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "max",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
								&influxql.IntegerLiteral{Val: 3},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// cumulative_sum
		{
			s: fmt.Sprintf(`SELECT cumulative_sum(field1) FROM myseries WHERE time > '%s'`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "cumulative_sum",
							Args: []influxql.Expr{
								&influxql.VarRef{Val: "field1"},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		{
			s: fmt.Sprintf(`SELECT cumulative_sum(mean(field1)) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "cumulative_sum",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "mean",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// holt_winters
		{
			s: fmt.Sprintf(`SELECT holt_winters(first(field1), 3, 1) FROM myseries WHERE time > '%s' GROUP BY time(1h);`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{
						Name: "holt_winters",
						Args: []influxql.Expr{
							&influxql.Call{
								Name: "first",
								Args: []influxql.Expr{
									&influxql.VarRef{Val: "field1"},
								},
							},
							&influxql.IntegerLiteral{Val: 3},
							&influxql.IntegerLiteral{Val: 1},
						},
					}},
				},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: 1 * time.Hour},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		{
			s: fmt.Sprintf(`SELECT holt_winters_with_fit(first(field1), 3, 1) FROM myseries WHERE time > '%s' GROUP BY time(1h);`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{
						Name: "holt_winters_with_fit",
						Args: []influxql.Expr{
							&influxql.Call{
								Name: "first",
								Args: []influxql.Expr{
									&influxql.VarRef{Val: "field1"},
								},
							},
							&influxql.IntegerLiteral{Val: 3},
							&influxql.IntegerLiteral{Val: 1},
						}}},
				},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: 1 * time.Hour},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},
		{
			s: fmt.Sprintf(`SELECT holt_winters(max(field1), 4, 5) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "holt_winters",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "max",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
								&influxql.IntegerLiteral{Val: 4},
								&influxql.IntegerLiteral{Val: 5},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		{
			s: fmt.Sprintf(`SELECT holt_winters_with_fit(max(field1), 4, 5) FROM myseries WHERE time > '%s' GROUP BY time(1m)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.Call{
							Name: "holt_winters_with_fit",
							Args: []influxql.Expr{
								&influxql.Call{
									Name: "max",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "field1"},
									},
								},
								&influxql.IntegerLiteral{Val: 4},
								&influxql.IntegerLiteral{Val: 5},
							},
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				Dimensions: []*influxql.Dimension{
					{
						Expr: &influxql.Call{
							Name: "time",
							Args: []influxql.Expr{
								&influxql.DurationLiteral{Val: time.Minute},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// SELECT statement (lowercase)
		{
			s: `select my_field from myseries`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.VarRef{Val: "my_field"}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		// SELECT statement (lowercase) with quoted field
		{
			s: `select 'my_field' from myseries`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.StringLiteral{Val: "my_field"}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
			},
		},

		// SELECT statement with multiple ORDER BY fields
		{
			skip: true,
			s:    `SELECT field1 FROM myseries ORDER BY ASC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.VarRef{Val: "field1"}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				SortFields: []*influxql.SortField{
					{Ascending: true},
					{Name: "field1"},
					{Name: "field2"},
				},
				Limit: 10,
			},
		},

		// SELECT statement with SLIMIT and SOFFSET
		{
			s: `SELECT field1 FROM myseries SLIMIT 10 SOFFSET 5`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.VarRef{Val: "field1"}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				SLimit:     10,
				SOffset:    5,
			},
		},

		// SELECT * FROM cpu WHERE host = 'serverC' AND region =~ /.*west.*/
		{
			s: `SELECT * FROM cpu WHERE host = 'serverC' AND region =~ /.*west.*/`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op: influxql.AND,
					LHS: &influxql.BinaryExpr{
						Op:  influxql.EQ,
						LHS: &influxql.VarRef{Val: "host"},
						RHS: &influxql.StringLiteral{Val: "serverC"},
					},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.EQREGEX,
						LHS: &influxql.VarRef{Val: "region"},
						RHS: &influxql.RegexLiteral{Val: regexp.MustCompile(".*west.*")},
					},
				},
			},
		},

		// select percentile statements
		{
			s: `select percentile("field1", 2.0) from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "percentile", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.NumberLiteral{Val: 2.0}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `select percentile("field1", 2.0), field2 from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "percentile", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.NumberLiteral{Val: 2.0}}}},
					{Expr: &influxql.VarRef{Val: "field2"}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// select top statements
		{
			s: `select top("field1", 2) from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "top", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.IntegerLiteral{Val: 2}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `select top(field1, 2) from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "top", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.IntegerLiteral{Val: 2}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `select top(field1, 2), tag1 from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "top", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.IntegerLiteral{Val: 2}}}},
					{Expr: &influxql.VarRef{Val: "tag1"}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `select top(field1, tag1, 2), tag1 from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "top", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}, &influxql.VarRef{Val: "tag1"}, &influxql.IntegerLiteral{Val: 2}}}},
					{Expr: &influxql.VarRef{Val: "tag1"}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// select distinct statements
		{
			s: `select distinct(field1) from cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "distinct", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `select distinct field2 from network`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.Distinct{Val: "field2"}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "network"}},
			},
		},

		{
			s: `select count(distinct field3) from metrics`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.Distinct{Val: "field3"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "metrics"}},
			},
		},

		{
			s: `select count(distinct field3), sum(field4) from metrics`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.Distinct{Val: "field3"}}}},
					{Expr: &influxql.Call{Name: "sum", Args: []influxql.Expr{&influxql.VarRef{Val: "field4"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "metrics"}},
			},
		},

		{
			s: `select count(distinct(field3)), sum(field4) from metrics`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.Call{Name: "distinct", Args: []influxql.Expr{&influxql.VarRef{Val: "field3"}}}}}},
					{Expr: &influxql.Call{Name: "sum", Args: []influxql.Expr{&influxql.VarRef{Val: "field4"}}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "metrics"}},
			},
		},

		// SELECT * FROM WHERE time
		{
			s: fmt.Sprintf(`SELECT * FROM cpu WHERE time > '%s'`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
			},
		},

		// SELECT * FROM WHERE field comparisons
		{
			s: `SELECT * FROM cpu WHERE load > 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},
		{
			s: `SELECT * FROM cpu WHERE load >= 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GTE,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},
		{
			s: `SELECT * FROM cpu WHERE load = 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},
		{
			s: `SELECT * FROM cpu WHERE load <= 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LTE,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},
		{
			s: `SELECT * FROM cpu WHERE load < 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},
		{
			s: `SELECT * FROM cpu WHERE load != 100`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.NEQ,
					LHS: &influxql.VarRef{Val: "load"},
					RHS: &influxql.IntegerLiteral{Val: 100},
				},
			},
		},

		// SELECT * FROM /<regex>/
		{
			s: `SELECT * FROM /cpu.*/`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources: []influxql.Source{&influxql.Measurement{
					Regex: &influxql.RegexLiteral{Val: regexp.MustCompile("cpu.*")}},
				},
			},
		},

		// SELECT * FROM "db"."rp"./<regex>/
		{
			s: `SELECT * FROM "db"."rp"./cpu.*/`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources: []influxql.Source{&influxql.Measurement{
					Database:        `db`,
					RetentionPolicy: `rp`,
					Regex:           &influxql.RegexLiteral{Val: regexp.MustCompile("cpu.*")}},
				},
			},
		},

		// SELECT * FROM "db"../<regex>/
		{
			s: `SELECT * FROM "db"../cpu.*/`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources: []influxql.Source{&influxql.Measurement{
					Database: `db`,
					Regex:    &influxql.RegexLiteral{Val: regexp.MustCompile("cpu.*")}},
				},
			},
		},

		// SELECT * FROM "rp"./<regex>/
		{
			s: `SELECT * FROM "rp"./cpu.*/`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
				Sources: []influxql.Source{&influxql.Measurement{
					RetentionPolicy: `rp`,
					Regex:           &influxql.RegexLiteral{Val: regexp.MustCompile("cpu.*")}},
				},
			},
		},

		// SELECT statement with group by
		{
			s: `SELECT sum(value) FROM "kbps" WHERE time > now() - 120s AND deliveryservice='steam-dns' and cachegroup = 'total' GROUP BY time(60s)`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: false,
				Fields: []*influxql.Field{
					{Expr: &influxql.Call{Name: "sum", Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}},
				},
				Sources:    []influxql.Source{&influxql.Measurement{Name: "kbps"}},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 60 * time.Second}}}}},
				Condition: &influxql.BinaryExpr{ // 1
					Op: influxql.AND,
					LHS: &influxql.BinaryExpr{ // 2
						Op: influxql.AND,
						LHS: &influxql.BinaryExpr{ //3
							Op:  influxql.GT,
							LHS: &influxql.VarRef{Val: "time"},
							RHS: &influxql.BinaryExpr{
								Op:  influxql.SUB,
								LHS: &influxql.Call{Name: "now"},
								RHS: &influxql.DurationLiteral{Val: mustParseDuration("120s")},
							},
						},
						RHS: &influxql.BinaryExpr{
							Op:  influxql.EQ,
							LHS: &influxql.VarRef{Val: "deliveryservice"},
							RHS: &influxql.StringLiteral{Val: "steam-dns"},
						},
					},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.EQ,
						LHS: &influxql.VarRef{Val: "cachegroup"},
						RHS: &influxql.StringLiteral{Val: "total"},
					},
				},
			},
		},
		// SELECT statement with group by and multi digit duration (prevent regression from #731://github.com/influxdata/influxdb/pull/7316)
		{
			s: fmt.Sprintf(`SELECT count(value) FROM cpu where time < '%s' group by time(500ms)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "count",
						Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 500 * time.Millisecond}}}}},
			},
		},

		// SELECT statement with fill
		{
			s: fmt.Sprintf(`SELECT mean(value) FROM cpu where time < '%s' GROUP BY time(5m) fill(1)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 5 * time.Minute}}}}},
				Fill:       influxql.NumberFill,
				FillValue:  int64(1),
			},
		},

		// SELECT statement with FILL(none) -- check case insensitivity
		{
			s: fmt.Sprintf(`SELECT mean(value) FROM cpu where time < '%s' GROUP BY time(5m) FILL(none)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 5 * time.Minute}}}}},
				Fill:       influxql.NoFill,
			},
		},

		// SELECT statement with previous fill
		{
			s: fmt.Sprintf(`SELECT mean(value) FROM cpu where time < '%s' GROUP BY time(5m) FILL(previous)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 5 * time.Minute}}}}},
				Fill:       influxql.PreviousFill,
			},
		},

		// SELECT statement with average fill
		{
			s: fmt.Sprintf(`SELECT mean(value) FROM cpu where time < '%s' GROUP BY time(5m) FILL(linear)`, now.UTC().Format(time.RFC3339Nano)),
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.StringLiteral{Val: now.UTC().Format(time.RFC3339Nano)},
				},
				Dimensions: []*influxql.Dimension{{Expr: &influxql.Call{Name: "time", Args: []influxql.Expr{&influxql.DurationLiteral{Val: 5 * time.Minute}}}}},
				Fill:       influxql.LinearFill,
			},
		},

		// SELECT casts
		{
			s: `SELECT field1::float, field2::integer, field6::unsigned, field3::string, field4::boolean, field5::field, tag1::tag FROM cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{
						Expr: &influxql.VarRef{
							Val:  "field1",
							Type: influxql.Float,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "field2",
							Type: influxql.Integer,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "field6",
							Type: influxql.Unsigned,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "field3",
							Type: influxql.String,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "field4",
							Type: influxql.Boolean,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "field5",
							Type: influxql.AnyField,
						},
					},
					{
						Expr: &influxql.VarRef{
							Val:  "tag1",
							Type: influxql.Tag,
						},
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a bound parameter
		{
			s: `SELECT value FROM cpu WHERE value > $value`,
			params: map[string]interface{}{
				"value": int64(2),
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{Val: "value"}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "value"},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
			},
		},

		// SELECT statement with a bound parameter that contains spaces
		{
			s: `SELECT value FROM cpu WHERE value > $"multi-word value"`,
			params: map[string]interface{}{
				"multi-word value": int64(2),
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{Val: "value"}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "value"},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
			},
		},

		// SELECT statement with a field as a bound parameter.
		{
			s: `SELECT mean($field) FROM cpu`,
			params: map[string]interface{}{
				"field": map[string]interface{}{"identifier": "value"},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a function as a bound parameter.
		{
			s: `SELECT $fn(value) FROM cpu`,
			params: map[string]interface{}{
				"fn": map[string]interface{}{"identifier": "mean"},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a regex as a bound parameter.
		{
			s: `SELECT mean(value) FROM cpu WHERE host =~ $host`,
			params: map[string]interface{}{
				"host": map[string]interface{}{"regex": "^server.*"},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						}}},
				},
				Condition: &influxql.BinaryExpr{
					Op: influxql.EQREGEX,
					LHS: &influxql.VarRef{
						Val: "host",
					},
					RHS: &influxql.RegexLiteral{
						Val: regexp.MustCompile(`^server.*`),
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a field and type as a bound parameter.
		{
			s: `SELECT $field::$type FROM cpu`,
			params: map[string]interface{}{
				"field": map[string]interface{}{"identifier": "value"},
				"type":  map[string]interface{}{"identifier": "integer"},
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{
						Val:  "value",
						Type: influxql.Integer,
					}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a float as a bound parameter.
		{
			s: `SELECT value FROM cpu WHERE value > $f`,
			params: map[string]interface{}{
				"f": map[string]interface{}{"float": 2.0},
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{
						Val: "value",
					}},
				},
				Condition: &influxql.BinaryExpr{
					Op: influxql.GT,
					LHS: &influxql.VarRef{
						Val: "value",
					},
					RHS: &influxql.NumberLiteral{
						Val: 2,
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a float as an integer in a bound parameter.
		{
			s: `SELECT value FROM cpu WHERE value > $f`,
			params: map[string]interface{}{
				"f": map[string]interface{}{"float": int64(2)},
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{
						Val: "value",
					}},
				},
				Condition: &influxql.BinaryExpr{
					Op: influxql.GT,
					LHS: &influxql.VarRef{
						Val: "value",
					},
					RHS: &influxql.NumberLiteral{
						Val: 2,
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with an integer in a bound parameter.
		{
			s: `SELECT value FROM cpu WHERE value > $i`,
			params: map[string]interface{}{
				"i": map[string]interface{}{"integer": int64(2)},
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.VarRef{
						Val: "value",
					}},
				},
				Condition: &influxql.BinaryExpr{
					Op: influxql.GT,
					LHS: &influxql.VarRef{
						Val: "value",
					},
					RHS: &influxql.IntegerLiteral{
						Val: 2,
					},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with group by interval with a bound parameter.
		{
			s: `SELECT mean(value) FROM cpu GROUP BY time($interval)`,
			params: map[string]interface{}{
				"interval": map[string]interface{}{"duration": "10s"},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						},
					}},
				},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: 10 * time.Second},
						},
					},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with group by interval integer with a bound parameter.
		{
			s: `SELECT mean(value) FROM cpu GROUP BY time($interval)`,
			params: map[string]interface{}{
				"interval": map[string]interface{}{"duration": int64(10 * time.Second)},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						},
					}},
				},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: 10 * time.Second},
						},
					},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with group by interval integer with a bound parameter and nanosecond precision.
		{
			s: `SELECT mean(value) FROM cpu GROUP BY time($interval)`,
			params: map[string]interface{}{
				"interval": map[string]interface{}{"duration": int64(10)},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						},
					}},
				},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: 10},
						},
					},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with group by interval json number with a bound parameter and nanosecond precision.
		{
			s: `SELECT mean(value) FROM cpu GROUP BY time($interval)`,
			params: map[string]interface{}{
				"interval": map[string]interface{}{"duration": json.Number("10")},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"},
						},
					}},
				},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: 10},
						},
					},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SELECT statement with a subquery
		{
			s: `SELECT sum(derivative) FROM (SELECT derivative(value) FROM cpu GROUP BY host) WHERE time >= now() - 1d GROUP BY time(1h)`,
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "sum",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "derivative"},
						}},
				}},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: time.Hour},
						},
					},
				}},
				Sources: []influxql.Source{
					&influxql.SubQuery{
						Statement: &influxql.SelectStatement{
							Fields: []*influxql.Field{{
								Expr: &influxql.Call{
									Name: "derivative",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "value"},
									},
								},
							}},
							Dimensions: []*influxql.Dimension{{
								Expr: &influxql.VarRef{Val: "host"},
							}},
							Sources: []influxql.Source{
								&influxql.Measurement{Name: "cpu"},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GTE,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.SUB,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: 24 * time.Hour},
					},
				},
			},
		},

		{
			s: `SELECT sum(mean) FROM (SELECT mean(value) FROM cpu GROUP BY time(1h)) WHERE time >= now() - 1d`,
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "sum",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "mean"},
						}},
				}},
				Sources: []influxql.Source{
					&influxql.SubQuery{
						Statement: &influxql.SelectStatement{
							Fields: []*influxql.Field{{
								Expr: &influxql.Call{
									Name: "mean",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "value"},
									},
								},
							}},
							Dimensions: []*influxql.Dimension{{
								Expr: &influxql.Call{
									Name: "time",
									Args: []influxql.Expr{
										&influxql.DurationLiteral{Val: time.Hour},
									},
								},
							}},
							Sources: []influxql.Source{
								&influxql.Measurement{Name: "cpu"},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GTE,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.SUB,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: 24 * time.Hour},
					},
				},
			},
		},

		{
			s: `SELECT sum(mean) FROM (SELECT mean(value) FROM cpu WHERE time >= now() - 1d GROUP BY time(1h))`,
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "sum",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "mean"},
						}},
				}},
				Sources: []influxql.Source{
					&influxql.SubQuery{
						Statement: &influxql.SelectStatement{
							Fields: []*influxql.Field{{
								Expr: &influxql.Call{
									Name: "mean",
									Args: []influxql.Expr{
										&influxql.VarRef{Val: "value"},
									},
								},
							}},
							Dimensions: []*influxql.Dimension{{
								Expr: &influxql.Call{
									Name: "time",
									Args: []influxql.Expr{
										&influxql.DurationLiteral{Val: time.Hour},
									},
								},
							}},
							Condition: &influxql.BinaryExpr{
								Op:  influxql.GTE,
								LHS: &influxql.VarRef{Val: "time"},
								RHS: &influxql.BinaryExpr{
									Op:  influxql.SUB,
									LHS: &influxql.Call{Name: "now"},
									RHS: &influxql.DurationLiteral{Val: 24 * time.Hour},
								},
							},
							Sources: []influxql.Source{
								&influxql.Measurement{Name: "cpu"},
							},
						},
					},
				},
			},
		},

		{
			s: `SELECT sum(derivative) FROM (SELECT derivative(mean(value)) FROM cpu GROUP BY host) WHERE time >= now() - 1d GROUP BY time(1h)`,
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "sum",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "derivative"},
						}},
				}},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: time.Hour},
						},
					},
				}},
				Sources: []influxql.Source{
					&influxql.SubQuery{
						Statement: &influxql.SelectStatement{
							Fields: []*influxql.Field{{
								Expr: &influxql.Call{
									Name: "derivative",
									Args: []influxql.Expr{
										&influxql.Call{
											Name: "mean",
											Args: []influxql.Expr{
												&influxql.VarRef{Val: "value"},
											},
										},
									},
								},
							}},
							Dimensions: []*influxql.Dimension{{
								Expr: &influxql.VarRef{Val: "host"},
							}},
							Sources: []influxql.Source{
								&influxql.Measurement{Name: "cpu"},
							},
						},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GTE,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.SUB,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: 24 * time.Hour},
					},
				},
			},
		},

		// select statements with intertwined comments
		{
			s: `SELECT "user" /*, system, idle */ FROM cpu`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.VarRef{Val: "user"}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		{
			s: `SELECT /foo\/*bar/ FROM /foo\/*bar*/ WHERE x = 1`,
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{
					{Expr: &influxql.RegexLiteral{Val: regexp.MustCompile(`foo/*bar`)}},
				},
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`foo/*bar*`)},
					},
				},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "x"},
					RHS: &influxql.IntegerLiteral{Val: 1},
				},
			},
		},

		// SELECT statement with a time zone
		{
			s: `SELECT mean(value) FROM cpu WHERE time >= now() - 7d GROUP BY time(1d) TZ('America/Los_Angeles')`,
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "mean",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "value"}},
					}}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.GTE,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.SUB,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: 7 * 24 * time.Hour},
					},
				},
				Dimensions: []*influxql.Dimension{{
					Expr: &influxql.Call{
						Name: "time",
						Args: []influxql.Expr{
							&influxql.DurationLiteral{Val: 24 * time.Hour}}}}},
				Location: LosAngeles,
			},
		},

		// EXPLAIN ...
		{
			s: `EXPLAIN SELECT * FROM cpu`,
			stmt: &influxql.ExplainStatement{
				Statement: &influxql.SelectStatement{
					IsRawQuery: true,
					Fields: []*influxql.Field{
						{Expr: &influxql.Wildcard{}},
					},
					Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				},
			},
		},

		// EXPLAIN ANALYZE ...
		{
			s: `EXPLAIN ANALYZE SELECT * FROM cpu`,
			stmt: &influxql.ExplainStatement{
				Statement: &influxql.SelectStatement{
					IsRawQuery: true,
					Fields: []*influxql.Field{
						{Expr: &influxql.Wildcard{}},
					},
					Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				},
				Analyze: true,
			},
		},

		// See issues https://github.com/influxdata/influxdb/issues/1647
		// and https://github.com/influxdata/influxdb/issues/4404
		// DELETE statement
		//{
		//	s: `DELETE FROM myseries WHERE host = 'hosta.influxdb.org'`,
		//	stmt: &influxql.DeleteStatement{
		//		Source: &influxql.Measurement{Name: "myseries"},
		//		Condition: &influxql.BinaryExpr{
		//			Op:  influxql.EQ,
		//			LHS: &influxql.VarRef{Val: "host"},
		//			RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
		//		},
		//	},
		//},

		// SHOW GRANTS
		{
			s:    `SHOW GRANTS FOR jdoe`,
			stmt: &influxql.ShowGrantsForUserStatement{Name: "jdoe"},
		},

		// SHOW DATABASES
		{
			s:    `SHOW DATABASES`,
			stmt: &influxql.ShowDatabasesStatement{},
		},

		// SHOW SERIES statement
		{
			s:    `SHOW SERIES`,
			stmt: &influxql.ShowSeriesStatement{},
		},

		// SHOW SERIES FROM
		{
			s: `SHOW SERIES FROM cpu`,
			stmt: &influxql.ShowSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW SERIES ON db0
		{
			s: `SHOW SERIES ON db0`,
			stmt: &influxql.ShowSeriesStatement{
				Database: "db0",
			},
		},

		// SHOW SERIES FROM /<regex>/
		{
			s: `SHOW SERIES FROM /[cg]pu/`,
			stmt: &influxql.ShowSeriesStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW SERIES with OFFSET 0
		{
			s:    `SHOW SERIES OFFSET 0`,
			stmt: &influxql.ShowSeriesStatement{Offset: 0},
		},

		// SHOW SERIES with LIMIT 2 OFFSET 0
		{
			s:    `SHOW SERIES LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowSeriesStatement{Offset: 0, Limit: 2},
		},

		// SHOW SERIES WHERE with ORDER BY and LIMIT
		{
			skip: true,
			s:    `SHOW SERIES WHERE region = 'order by desc' ORDER BY DESC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.ShowSeriesStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				SortFields: []*influxql.SortField{
					&influxql.SortField{Ascending: false},
					&influxql.SortField{Name: "field1", Ascending: true},
					&influxql.SortField{Name: "field2"},
				},
				Limit: 10,
			},
		},

		// SHOW SERIES CARDINALITY statement
		{
			s:    `SHOW SERIES CARDINALITY`,
			stmt: &influxql.ShowSeriesCardinalityStatement{},
		},

		// SHOW SERIES CARDINALITY ON dbz statement
		{
			s:    `SHOW SERIES CARDINALITY ON dbz`,
			stmt: &influxql.ShowSeriesCardinalityStatement{Database: "dbz"},
		},

		// SHOW SERIES EXACT CARDINALITY statement
		{
			s:    `SHOW SERIES EXACT CARDINALITY`,
			stmt: &influxql.ShowSeriesCardinalityStatement{Exact: true},
		},

		// SHOW SERIES EXACT CARDINALITY FROM cpu
		{
			s: `SHOW SERIES EXACT CARDINALITY FROM cpu`,
			stmt: &influxql.ShowSeriesCardinalityStatement{
				Exact:   true,
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW SERIES EXACT CARDINALITY ON db0
		{
			s: `SHOW SERIES EXACT CARDINALITY ON db0`,
			stmt: &influxql.ShowSeriesCardinalityStatement{
				Exact:    true,
				Database: "db0",
			},
		},

		// SHOW SERIES EXACT CARDINALITY FROM /<regex>/
		{
			s: `SHOW SERIES EXACT CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowSeriesCardinalityStatement{
				Exact: true,
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW SERIES EXACT CARDINALITY with OFFSET 0
		{
			s:    `SHOW SERIES EXACT CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowSeriesCardinalityStatement{Exact: true, Offset: 0},
		},

		// SHOW SERIES EXACT CARDINALITY with LIMIT 2 OFFSET 0
		{
			s:    `SHOW SERIES EXACT CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowSeriesCardinalityStatement{Exact: true, Offset: 0, Limit: 2},
		},

		// SHOW SERIES EXACT CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW SERIES EXACT CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowSeriesCardinalityStatement{
				Exact: true,
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW MEASUREMENTS WHERE with ORDER BY and LIMIT
		{
			skip: true,
			s:    `SHOW MEASUREMENTS WHERE region = 'uswest' ORDER BY ASC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.ShowMeasurementsStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
				SortFields: []*influxql.SortField{
					{Ascending: true},
					{Name: "field1"},
					{Name: "field2"},
				},
				Limit: 10,
			},
		},

		// SHOW MEASUREMENTS ON db0
		{
			s: `SHOW MEASUREMENTS ON db0`,
			stmt: &influxql.ShowMeasurementsStatement{
				Database: "db0",
			},
		},

		// SHOW MEASUREMENTS WITH MEASUREMENT = cpu
		{
			s: `SHOW MEASUREMENTS WITH MEASUREMENT = cpu`,
			stmt: &influxql.ShowMeasurementsStatement{
				Source: &influxql.Measurement{Name: "cpu"},
			},
		},

		// SHOW MEASUREMENTS WITH MEASUREMENT =~ /regex/
		{
			s: `SHOW MEASUREMENTS WITH MEASUREMENT =~ /[cg]pu/`,
			stmt: &influxql.ShowMeasurementsStatement{
				Source: &influxql.Measurement{
					Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
				},
			},
		},

		// SHOW MEASUREMENT CARDINALITY statement
		{
			s:    `SHOW MEASUREMENT CARDINALITY`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{},
		},

		// SHOW MEASUREMENT CARDINALITY ON db0 statement
		{
			s: `SHOW MEASUREMENT CARDINALITY ON db0`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact:    false,
				Database: "db0",
			},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY statement
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact: true,
			},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY FROM cpu
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY FROM cpu`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact:   true,
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY ON db0
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY ON db0`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact:    true,
				Database: "db0",
			},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY FROM /<regex>/
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact: true,
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY with OFFSET 0
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact: true, Offset: 0},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY with LIMIT 2 OFFSET 0
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact: true, Offset: 0, Limit: 2},
		},

		// SHOW MEASUREMENT EXACT CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW MEASUREMENT EXACT CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowMeasurementCardinalityStatement{
				Exact: true,
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW QUERIES
		{
			s:    `SHOW QUERIES`,
			stmt: &influxql.ShowQueriesStatement{},
		},

		// KILL QUERY 4
		{
			s: `KILL QUERY 4`,
			stmt: &influxql.KillQueryStatement{
				QueryID: 4,
			},
		},

		// KILL QUERY 4 ON localhost
		{
			s: `KILL QUERY 4 ON localhost`,
			stmt: &influxql.KillQueryStatement{
				QueryID: 4,
				Host:    "localhost",
			},
		},

		// SHOW RETENTION POLICIES
		{
			s:    `SHOW RETENTION POLICIES`,
			stmt: &influxql.ShowRetentionPoliciesStatement{},
		},

		// SHOW RETENTION POLICIES ON db0
		{
			s: `SHOW RETENTION POLICIES ON db0`,
			stmt: &influxql.ShowRetentionPoliciesStatement{
				Database: "db0",
			},
		},
		// SHOW TAG KEY CARDINALITY statement
		{
			s:    `SHOW TAG KEY CARDINALITY`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{},
		},

		// SHOW TAG KEY CARDINALITY FROM cpu
		{
			s: `SHOW TAG KEY CARDINALITY FROM cpu`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW TAG KEY CARDINALITY ON db0
		{
			s: `SHOW TAG KEY CARDINALITY ON db0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Database: "db0",
			},
		},

		// SHOW TAG KEY CARDINALITY FROM /<regex>/
		{
			s: `SHOW TAG KEY CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW TAG KEY CARDINALITY with OFFSET 0
		{
			s:    `SHOW TAG KEY CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{Offset: 0},
		},

		// SHOW TAG KEY CARDINALITY with LIMIT 2 OFFSET 0
		{
			s:    `SHOW TAG KEY CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{Offset: 0, Limit: 2},
		},

		// SHOW TAG KEY CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW TAG KEY CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW TAG KEY EXACT CARDINALITY statement
		{
			s: `SHOW TAG KEY EXACT CARDINALITY`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Exact: true,
			},
		},

		// SHOW TAG KEY EXACT CARDINALITY FROM cpu
		{
			s: `SHOW TAG KEY EXACT CARDINALITY FROM cpu`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Exact:   true,
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW TAG KEY EXACT CARDINALITY ON db0
		{
			s: `SHOW TAG KEY EXACT CARDINALITY ON db0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Exact:    true,
				Database: "db0",
			},
		},

		// SHOW TAG KEY EXACT CARDINALITY FROM /<regex>/
		{
			s: `SHOW TAG KEY EXACT CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Exact: true,
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW TAG KEY EXACT CARDINALITY with OFFSET 0
		{
			s:    `SHOW TAG KEY EXACT CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{Exact: true, Offset: 0},
		},

		// SHOW TAG KEY EXACT CARDINALITY with LIMIT 2 OFFSET 0
		{
			s:    `SHOW TAG KEY EXACT CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{Exact: true, Offset: 0, Limit: 2},
		},

		// SHOW TAG KEY EXACT CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW TAG KEY EXACT CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowTagKeyCardinalityStatement{
				Exact: true,
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW TAG KEYS
		{
			s: `SHOW TAG KEYS FROM src`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
			},
		},

		// SHOW TAG KEYS ON db0
		{
			s: `SHOW TAG KEYS ON db0`,
			stmt: &influxql.ShowTagKeysStatement{
				Database: "db0",
			},
		},

		// SHOW TAG KEYS with LIMIT
		{
			s: `SHOW TAG KEYS FROM src LIMIT 2`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Limit:   2,
			},
		},

		// SHOW TAG KEYS with OFFSET
		{
			s: `SHOW TAG KEYS FROM src OFFSET 1`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Offset:  1,
			},
		},

		// SHOW TAG KEYS with LIMIT and OFFSET
		{
			s: `SHOW TAG KEYS FROM src LIMIT 2 OFFSET 1`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Limit:   2,
				Offset:  1,
			},
		},

		// SHOW TAG KEYS with SLIMIT
		{
			s: `SHOW TAG KEYS FROM src SLIMIT 2`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				SLimit:  2,
			},
		},

		// SHOW TAG KEYS with SOFFSET
		{
			s: `SHOW TAG KEYS FROM src SOFFSET 1`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				SOffset: 1,
			},
		},

		// SHOW TAG KEYS with SLIMIT and SOFFSET
		{
			s: `SHOW TAG KEYS FROM src SLIMIT 2 SOFFSET 1`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				SLimit:  2,
				SOffset: 1,
			},
		},

		// SHOW TAG KEYS with LIMIT, OFFSET, SLIMIT, and SOFFSET
		{
			s: `SHOW TAG KEYS FROM src LIMIT 4 OFFSET 3 SLIMIT 2 SOFFSET 1`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Limit:   4,
				Offset:  3,
				SLimit:  2,
				SOffset: 1,
			},
		},

		// SHOW TAG KEYS FROM /<regex>/
		{
			s: `SHOW TAG KEYS FROM /[cg]pu/`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW TAG KEYS
		{
			skip: true,
			s:    `SHOW TAG KEYS FROM src WHERE region = 'uswest' ORDER BY ASC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.ShowTagKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
				SortFields: []*influxql.SortField{
					{Ascending: true},
					{Name: "field1"},
					{Name: "field2"},
				},
				Limit: 10,
			},
		},

		// SHOW TAG VALUES FROM ... WITH KEY = ...
		{
			skip: true,
			s:    `SHOW TAG VALUES FROM src WITH KEY = region WHERE region = 'uswest' ORDER BY ASC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.ShowTagValuesStatement{
				Sources:    []influxql.Source{&influxql.Measurement{Name: "src"}},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "region"},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
				SortFields: []*influxql.SortField{
					{Ascending: true},
					{Name: "field1"},
					{Name: "field2"},
				},
				Limit: 10,
			},
		},

		// SHOW TAG VALUES FROM ... WITH KEY IN...
		{
			s: `SHOW TAG VALUES FROM cpu WITH KEY IN (region, host) WHERE region = 'uswest'`,
			stmt: &influxql.ShowTagValuesStatement{
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Op:         influxql.IN,
				TagKeyExpr: &influxql.ListLiteral{Vals: []string{"region", "host"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
			},
		},

		// SHOW TAG VALUES ... AND TAG KEY =
		{
			s: `SHOW TAG VALUES FROM cpu WITH KEY IN (region,service,host)WHERE region = 'uswest'`,
			stmt: &influxql.ShowTagValuesStatement{
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Op:         influxql.IN,
				TagKeyExpr: &influxql.ListLiteral{Vals: []string{"region", "service", "host"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
			},
		},

		// SHOW TAG VALUES WITH KEY = ...
		{
			s: `SHOW TAG VALUES WITH KEY = host WHERE region = 'uswest'`,
			stmt: &influxql.ShowTagValuesStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
			},
		},

		// SHOW TAG VALUES FROM /<regex>/ WITH KEY = ...
		{
			s: `SHOW TAG VALUES FROM /[cg]pu/ WITH KEY = host`,
			stmt: &influxql.ShowTagValuesStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES WITH KEY = "..."
		{
			s: `SHOW TAG VALUES WITH KEY = "host" WHERE region = 'uswest'`,
			stmt: &influxql.ShowTagValuesStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: `host`},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "uswest"},
				},
			},
		},

		// SHOW TAG VALUES WITH KEY =~ /<regex>/
		{
			s: `SHOW TAG VALUES WITH KEY =~ /(host|region)/`,
			stmt: &influxql.ShowTagValuesStatement{
				Op:         influxql.EQREGEX,
				TagKeyExpr: &influxql.RegexLiteral{Val: regexp.MustCompile(`(host|region)`)},
			},
		},

		// SHOW TAG VALUES ON db0
		{
			s: `SHOW TAG VALUES ON db0 WITH KEY = "host"`,
			stmt: &influxql.ShowTagValuesStatement{
				Database:   "db0",
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES CARDINALITY statement
		{
			s: `SHOW TAG VALUES CARDINALITY WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES CARDINALITY FROM cpu
		{
			s: `SHOW TAG VALUES CARDINALITY FROM cpu WITH KEY =  host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES CARDINALITY ON db0
		{
			s: `SHOW TAG VALUES CARDINALITY ON db0 WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Database:   "db0",
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES CARDINALITY FROM /<regex>/
		{
			s: `SHOW TAG VALUES CARDINALITY FROM /[cg]pu/ WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES CARDINALITY with OFFSET 0
		{
			s: `SHOW TAG VALUES CARDINALITY WITH KEY = host OFFSET 0`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Offset:     0,
			},
		},

		// SHOW TAG VALUES CARDINALITY with LIMIT 2 OFFSET 0
		{
			s: `SHOW TAG VALUES CARDINALITY WITH KEY = host LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Offset:     0,
				Limit:      2,
			},
		},

		// SHOW TAG VALUES CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW TAG VALUES CARDINALITY WITH KEY = host WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY statement
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY FROM cpu
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY FROM cpu WITH KEY =  host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu"}},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY ON db0
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY ON db0 WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Database:   "db0",
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY FROM /<regex>/
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY FROM /[cg]pu/ WITH KEY = host`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact: true,
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY with OFFSET 0
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY WITH KEY = host OFFSET 0`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Offset:     0,
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY with LIMIT 2 OFFSET 0
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY WITH KEY = host LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Offset:     0,
				Limit:      2,
			},
		},

		// SHOW TAG VALUES EXACT CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW TAG VALUES EXACT CARDINALITY WITH KEY = host WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowTagValuesCardinalityStatement{
				Exact:      true,
				Op:         influxql.EQ,
				TagKeyExpr: &influxql.StringLiteral{Val: "host"},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW USERS
		{
			s:    `SHOW USERS`,
			stmt: &influxql.ShowUsersStatement{},
		},

		// SHOW FIELD KEYS
		{
			skip: true,
			s:    `SHOW FIELD KEYS FROM src ORDER BY ASC, field1, field2 DESC LIMIT 10`,
			stmt: &influxql.ShowFieldKeysStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				SortFields: []*influxql.SortField{
					{Ascending: true},
					{Name: "field1"},
					{Name: "field2"},
				},
				Limit: 10,
			},
		},
		{
			s: `SHOW FIELD KEYS FROM /[cg]pu/`,
			stmt: &influxql.ShowFieldKeysStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},
		{
			s: `SHOW FIELD KEYS ON db0`,
			stmt: &influxql.ShowFieldKeysStatement{
				Database: "db0",
			},
		},

		// SHOW FIELD KEY CARDINALITY statement
		{
			s:    `SHOW FIELD KEY CARDINALITY`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{},
		},

		// SHOW FIELD KEY CARDINALITY FROM cpu
		{
			s: `SHOW FIELD KEY CARDINALITY FROM cpu`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW FIELD KEY CARDINALITY ON db0
		{
			s: `SHOW FIELD KEY CARDINALITY ON db0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Database: "db0",
			},
		},

		// SHOW FIELD KEY CARDINALITY FROM /<regex>/
		{
			s: `SHOW FIELD KEY CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW FIELD KEY CARDINALITY with OFFSET 0
		{
			s: `SHOW FIELD KEY CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Offset: 0,
			},
		},

		// SHOW FIELD KEY CARDINALITY with LIMIT 2 OFFSET 0
		{
			s: `SHOW FIELD KEY CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Offset: 0,
				Limit:  2,
			},
		},

		// SHOW FIELD KEY CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW FIELD KEY CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY statement
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact: true,
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY FROM cpu
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY FROM cpu`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact:   true,
				Sources: []influxql.Source{&influxql.Measurement{Name: "cpu"}},
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY ON db0
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY ON db0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact:    true,
				Database: "db0",
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY FROM /<regex>/
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY FROM /[cg]pu/`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact: true,
				Sources: []influxql.Source{
					&influxql.Measurement{
						Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`[cg]pu`)},
					},
				},
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY with OFFSET 0
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY OFFSET 0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact:  true,
				Offset: 0,
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY with LIMIT 2 OFFSET 0
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY LIMIT 2 OFFSET 0`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact:  true,
				Offset: 0,
				Limit:  2,
			},
		},

		// SHOW FIELD KEY EXACT CARDINALITY WHERE with ORDER BY and LIMIT
		{
			s: `SHOW FIELD KEY EXACT CARDINALITY WHERE region = 'order by desc' LIMIT 10`,
			stmt: &influxql.ShowFieldKeyCardinalityStatement{
				Exact: true,
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "region"},
					RHS: &influxql.StringLiteral{Val: "order by desc"},
				},
				Limit: 10,
			},
		},
		// DELETE statement
		{
			s:    `DELETE FROM src`,
			stmt: &influxql.DeleteSeriesStatement{Sources: []influxql.Source{&influxql.Measurement{Name: "src"}}},
		},
		{
			s: `DELETE WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DeleteSeriesStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DELETE FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DeleteSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},

		// DROP SERIES statement
		{
			s:    `DROP SERIES FROM src`,
			stmt: &influxql.DropSeriesStatement{Sources: []influxql.Source{&influxql.Measurement{Name: "src"}}},
		},
		{
			s: `DROP SERIES WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DropSeriesStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DROP SERIES FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DropSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},

		// SHOW CONTINUOUS QUERIES statement
		{
			s:    `SHOW CONTINUOUS QUERIES`,
			stmt: &influxql.ShowContinuousQueriesStatement{},
		},

		// CREATE CONTINUOUS QUERY ... INTO <measurement>
		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb RESAMPLE EVERY 1m FOR 1h BEGIN SELECT count(field1) INTO measure1 FROM myseries GROUP BY time(5m) END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					Fields:  []*influxql.Field{{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}}},
					Target:  &influxql.Target{Measurement: &influxql.Measurement{Name: "measure1", IsTarget: true}},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
					Dimensions: []*influxql.Dimension{
						{
							Expr: &influxql.Call{
								Name: "time",
								Args: []influxql.Expr{
									&influxql.DurationLiteral{Val: 5 * time.Minute},
								},
							},
						},
					},
				},
				ResampleEvery: time.Minute,
				ResampleFor:   time.Hour,
			},
		},

		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb RESAMPLE FOR 1h BEGIN SELECT count(field1) INTO measure1 FROM myseries GROUP BY time(5m) END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					Fields:  []*influxql.Field{{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}}},
					Target:  &influxql.Target{Measurement: &influxql.Measurement{Name: "measure1", IsTarget: true}},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
					Dimensions: []*influxql.Dimension{
						{
							Expr: &influxql.Call{
								Name: "time",
								Args: []influxql.Expr{
									&influxql.DurationLiteral{Val: 5 * time.Minute},
								},
							},
						},
					},
				},
				ResampleFor: time.Hour,
			},
		},

		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb RESAMPLE EVERY 1m BEGIN SELECT count(field1) INTO measure1 FROM myseries GROUP BY time(5m) END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					Fields:  []*influxql.Field{{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}}},
					Target:  &influxql.Target{Measurement: &influxql.Measurement{Name: "measure1", IsTarget: true}},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
					Dimensions: []*influxql.Dimension{
						{
							Expr: &influxql.Call{
								Name: "time",
								Args: []influxql.Expr{
									&influxql.DurationLiteral{Val: 5 * time.Minute},
								},
							},
						},
					},
				},
				ResampleEvery: time.Minute,
			},
		},

		{
			s: `create continuous query "this.is-a.test" on segments begin select * into measure1 from cpu_load_short end`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "this.is-a.test",
				Database: "segments",
				Source: &influxql.SelectStatement{
					IsRawQuery: true,
					Fields:     []*influxql.Field{{Expr: &influxql.Wildcard{}}},
					Target:     &influxql.Target{Measurement: &influxql.Measurement{Name: "measure1", IsTarget: true}},
					Sources:    []influxql.Source{&influxql.Measurement{Name: "cpu_load_short"}},
				},
			},
		},

		// CREATE CONTINUOUS QUERY ... INTO <retention-policy>.<measurement>
		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb BEGIN SELECT count(field1) INTO "1h.policy1"."cpu.load" FROM myseries GROUP BY time(5m) END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					Fields: []*influxql.Field{{Expr: &influxql.Call{Name: "count", Args: []influxql.Expr{&influxql.VarRef{Val: "field1"}}}}},
					Target: &influxql.Target{
						Measurement: &influxql.Measurement{RetentionPolicy: "1h.policy1", Name: "cpu.load", IsTarget: true},
					},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
					Dimensions: []*influxql.Dimension{
						{
							Expr: &influxql.Call{
								Name: "time",
								Args: []influxql.Expr{
									&influxql.DurationLiteral{Val: 5 * time.Minute},
								},
							},
						},
					},
				},
			},
		},

		// CREATE CONTINUOUS QUERY for non-aggregate SELECT stmts
		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb BEGIN SELECT value INTO "policy1"."value" FROM myseries END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					IsRawQuery: true,
					Fields:     []*influxql.Field{{Expr: &influxql.VarRef{Val: "value"}}},
					Target: &influxql.Target{
						Measurement: &influxql.Measurement{RetentionPolicy: "policy1", Name: "value", IsTarget: true},
					},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				},
			},
		},

		// CREATE CONTINUOUS QUERY for non-aggregate SELECT stmts with multiple values
		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb BEGIN SELECT transmit_rx, transmit_tx INTO "policy1"."network" FROM myseries END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					IsRawQuery: true,
					Fields: []*influxql.Field{{Expr: &influxql.VarRef{Val: "transmit_rx"}},
						{Expr: &influxql.VarRef{Val: "transmit_tx"}}},
					Target: &influxql.Target{
						Measurement: &influxql.Measurement{RetentionPolicy: "policy1", Name: "network", IsTarget: true},
					},
					Sources: []influxql.Source{&influxql.Measurement{Name: "myseries"}},
				},
			},
		},

		// CREATE CONTINUOUS QUERY with backreference measurement name
		{
			s: `CREATE CONTINUOUS QUERY myquery ON testdb BEGIN SELECT mean(value) INTO "policy1".:measurement FROM /^[a-z]+.*/ GROUP BY time(1m) END`,
			stmt: &influxql.CreateContinuousQueryStatement{
				Name:     "myquery",
				Database: "testdb",
				Source: &influxql.SelectStatement{
					Fields: []*influxql.Field{{Expr: &influxql.Call{Name: "mean", Args: []influxql.Expr{&influxql.VarRef{Val: "value"}}}}},
					Target: &influxql.Target{
						Measurement: &influxql.Measurement{RetentionPolicy: "policy1", IsTarget: true},
					},
					Sources: []influxql.Source{&influxql.Measurement{Regex: &influxql.RegexLiteral{Val: regexp.MustCompile(`^[a-z]+.*`)}}},
					Dimensions: []*influxql.Dimension{
						{
							Expr: &influxql.Call{
								Name: "time",
								Args: []influxql.Expr{
									&influxql.DurationLiteral{Val: 1 * time.Minute},
								},
							},
						},
					},
				},
			},
		},

		// CREATE DATABASE statement
		{
			s: `CREATE DATABASE testdb`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                  "testdb",
				RetentionPolicyCreate: false,
			},
		},
		{
			s: `CREATE DATABASE testdb WITH DURATION 24h`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                    "testdb",
				RetentionPolicyCreate:   true,
				RetentionPolicyDuration: duration(24 * time.Hour),
			},
		},
		{
			s: `CREATE DATABASE testdb WITH SHARD DURATION 30m`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                              "testdb",
				RetentionPolicyCreate:             true,
				RetentionPolicyShardGroupDuration: 30 * time.Minute,
			},
		},
		{
			s: `CREATE DATABASE testdb WITH REPLICATION 2`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                       "testdb",
				RetentionPolicyCreate:      true,
				RetentionPolicyReplication: intptr(2),
			},
		},
		{
			s: `CREATE DATABASE testdb WITH NAME test_name`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                  "testdb",
				RetentionPolicyCreate: true,
				RetentionPolicyName:   "test_name",
			},
		},
		{
			s: `CREATE DATABASE testdb WITH DURATION 24h REPLICATION 2 NAME test_name`,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                       "testdb",
				RetentionPolicyCreate:      true,
				RetentionPolicyDuration:    duration(24 * time.Hour),
				RetentionPolicyReplication: intptr(2),
				RetentionPolicyName:        "test_name",
			},
		},
		{
			s: `CREATE DATABASE testdb WITH DURATION 24h REPLICATION 2 SHARD DURATION 10m NAME test_name `,
			stmt: &influxql.CreateDatabaseStatement{
				Name:                              "testdb",
				RetentionPolicyCreate:             true,
				RetentionPolicyDuration:           duration(24 * time.Hour),
				RetentionPolicyReplication:        intptr(2),
				RetentionPolicyName:               "test_name",
				RetentionPolicyShardGroupDuration: 10 * time.Minute,
			},
		},

		// CREATE USER statement
		{
			s: `CREATE USER testuser WITH PASSWORD 'pwd1337'`,
			stmt: &influxql.CreateUserStatement{
				Name:     "testuser",
				Password: "pwd1337",
			},
		},

		// CREATE USER ... WITH ALL PRIVILEGES
		{
			s: `CREATE USER testuser WITH PASSWORD 'pwd1337' WITH ALL PRIVILEGES`,
			stmt: &influxql.CreateUserStatement{
				Name:     "testuser",
				Password: "pwd1337",
				Admin:    true,
			},
		},

		// SET PASSWORD FOR USER
		{
			s: `SET PASSWORD FOR testuser = 'pwd1337'`,
			stmt: &influxql.SetPasswordUserStatement{
				Name:     "testuser",
				Password: "pwd1337",
			},
		},

		// DROP CONTINUOUS QUERY statement
		{
			s:    `DROP CONTINUOUS QUERY myquery ON foo`,
			stmt: &influxql.DropContinuousQueryStatement{Name: "myquery", Database: "foo"},
		},

		// DROP DATABASE statement
		{
			s: `DROP DATABASE testdb`,
			stmt: &influxql.DropDatabaseStatement{
				Name: "testdb",
			},
		},

		// DROP MEASUREMENT statement
		{
			s:    `DROP MEASUREMENT cpu`,
			stmt: &influxql.DropMeasurementStatement{Name: "cpu"},
		},

		// DROP RETENTION POLICY
		{
			s: `DROP RETENTION POLICY "1h.cpu" ON mydb`,
			stmt: &influxql.DropRetentionPolicyStatement{
				Name:     `1h.cpu`,
				Database: `mydb`,
			},
		},

		// DROP USER statement
		{
			s:    `DROP USER jdoe`,
			stmt: &influxql.DropUserStatement{Name: "jdoe"},
		},

		// GRANT READ
		{
			s: `GRANT READ ON testdb TO jdoe`,
			stmt: &influxql.GrantStatement{
				Privilege: influxql.ReadPrivilege,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// GRANT WRITE
		{
			s: `GRANT WRITE ON testdb TO jdoe`,
			stmt: &influxql.GrantStatement{
				Privilege: influxql.WritePrivilege,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// GRANT ALL
		{
			s: `GRANT ALL ON testdb TO jdoe`,
			stmt: &influxql.GrantStatement{
				Privilege: influxql.AllPrivileges,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// GRANT ALL PRIVILEGES
		{
			s: `GRANT ALL PRIVILEGES ON testdb TO jdoe`,
			stmt: &influxql.GrantStatement{
				Privilege: influxql.AllPrivileges,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// GRANT ALL admin privilege
		{
			s: `GRANT ALL TO jdoe`,
			stmt: &influxql.GrantAdminStatement{
				User: "jdoe",
			},
		},

		// GRANT ALL PRVILEGES admin privilege
		{
			s: `GRANT ALL PRIVILEGES TO jdoe`,
			stmt: &influxql.GrantAdminStatement{
				User: "jdoe",
			},
		},

		// REVOKE READ
		{
			s: `REVOKE READ on testdb FROM jdoe`,
			stmt: &influxql.RevokeStatement{
				Privilege: influxql.ReadPrivilege,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// REVOKE WRITE
		{
			s: `REVOKE WRITE ON testdb FROM jdoe`,
			stmt: &influxql.RevokeStatement{
				Privilege: influxql.WritePrivilege,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// REVOKE ALL
		{
			s: `REVOKE ALL ON testdb FROM jdoe`,
			stmt: &influxql.RevokeStatement{
				Privilege: influxql.AllPrivileges,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// REVOKE ALL PRIVILEGES
		{
			s: `REVOKE ALL PRIVILEGES ON testdb FROM jdoe`,
			stmt: &influxql.RevokeStatement{
				Privilege: influxql.AllPrivileges,
				On:        "testdb",
				User:      "jdoe",
			},
		},

		// REVOKE ALL admin privilege
		{
			s: `REVOKE ALL FROM jdoe`,
			stmt: &influxql.RevokeAdminStatement{
				User: "jdoe",
			},
		},

		// REVOKE ALL PRIVILEGES admin privilege
		{
			s: `REVOKE ALL PRIVILEGES FROM jdoe`,
			stmt: &influxql.RevokeAdminStatement{
				User: "jdoe",
			},
		},

		// CREATE RETENTION POLICY
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 2`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:        "policy1",
				Database:    "testdb",
				Duration:    time.Hour,
				Replication: 2,
			},
		},

		// CREATE RETENTION POLICY with infinite retention
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION INF REPLICATION 2`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:        "policy1",
				Database:    "testdb",
				Duration:    0,
				Replication: 2,
			},
		},

		// CREATE RETENTION POLICY ... DEFAULT
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 2m REPLICATION 4 DEFAULT`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:        "policy1",
				Database:    "testdb",
				Duration:    2 * time.Minute,
				Replication: 4,
				Default:     true,
			},
		},
		// CREATE RETENTION POLICY
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 2 SHARD DURATION 30m`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:               "policy1",
				Database:           "testdb",
				Duration:           time.Hour,
				Replication:        2,
				ShardGroupDuration: 30 * time.Minute,
			},
		},
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 2 SHARD DURATION 0s`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:               "policy1",
				Database:           "testdb",
				Duration:           time.Hour,
				Replication:        2,
				ShardGroupDuration: 0,
			},
		},
		{
			s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 2 SHARD DURATION 1s`,
			stmt: &influxql.CreateRetentionPolicyStatement{
				Name:               "policy1",
				Database:           "testdb",
				Duration:           time.Hour,
				Replication:        2,
				ShardGroupDuration: time.Second,
			},
		},

		// ALTER RETENTION POLICY
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb DURATION 1m REPLICATION 4 DEFAULT`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", time.Minute, -1, 4, true),
		},

		// ALTER RETENTION POLICY with options in reverse order
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb DEFAULT REPLICATION 4 DURATION 1m`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", time.Minute, -1, 4, true),
		},

		// ALTER RETENTION POLICY with infinite retention
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb DEFAULT REPLICATION 4 DURATION INF`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", 0, -1, 4, true),
		},

		// ALTER RETENTION POLICY without optional DURATION
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb DEFAULT REPLICATION 4`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", -1, -1, 4, true),
		},

		// ALTER RETENTION POLICY without optional REPLICATION
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb DEFAULT`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", -1, -1, -1, true),
		},

		// ALTER RETENTION POLICY without optional DEFAULT
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb REPLICATION 4`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", -1, -1, 4, false),
		},
		// ALTER default retention policy unquoted
		{
			s:    `ALTER RETENTION POLICY default ON testdb REPLICATION 4`,
			stmt: newAlterRetentionPolicyStatement("default", "testdb", -1, -1, 4, false),
		},
		// ALTER RETENTION POLICY with SHARD duration
		{
			s:    `ALTER RETENTION POLICY policy1 ON testdb REPLICATION 4 SHARD DURATION 10m`,
			stmt: newAlterRetentionPolicyStatement("policy1", "testdb", -1, 10*time.Minute, 4, false),
		},
		// ALTER RETENTION POLICY with all options
		{
			s:    `ALTER RETENTION POLICY default ON testdb DURATION 0s REPLICATION 4 SHARD DURATION 10m DEFAULT`,
			stmt: newAlterRetentionPolicyStatement("default", "testdb", time.Duration(0), 10*time.Minute, 4, true),
		},
		// ALTER RETENTION POLICY with 0s shard duration
		{
			s:    `ALTER RETENTION POLICY default ON testdb DURATION 0s REPLICATION 1 SHARD DURATION 0s`,
			stmt: newAlterRetentionPolicyStatement("default", "testdb", time.Duration(0), 0, 1, false),
		},

		// SHOW STATS
		{
			s: `SHOW STATS`,
			stmt: &influxql.ShowStatsStatement{
				Module: "",
			},
		},
		{
			s: `SHOW STATS FOR 'cluster'`,
			stmt: &influxql.ShowStatsStatement{
				Module: "cluster",
			},
		},

		// SHOW SHARD GROUPS
		{
			s:    `SHOW SHARD GROUPS`,
			stmt: &influxql.ShowShardGroupsStatement{},
		},

		// SHOW SHARDS
		{
			s:    `SHOW SHARDS`,
			stmt: &influxql.ShowShardsStatement{},
		},

		// SHOW DIAGNOSTICS
		{
			s:    `SHOW DIAGNOSTICS`,
			stmt: &influxql.ShowDiagnosticsStatement{},
		},
		{
			s: `SHOW DIAGNOSTICS FOR 'build'`,
			stmt: &influxql.ShowDiagnosticsStatement{
				Module: "build",
			},
		},

		// CREATE SUBSCRIPTION
		{
			s: `CREATE SUBSCRIPTION "name" ON "db"."rp" DESTINATIONS ANY 'udp://host1:9093', 'udp://host2:9093'`,
			stmt: &influxql.CreateSubscriptionStatement{
				Name:            "name",
				Database:        "db",
				RetentionPolicy: "rp",
				Destinations:    []string{"udp://host1:9093", "udp://host2:9093"},
				Mode:            "ANY",
			},
		},

		// DROP SUBSCRIPTION
		{
			s: `DROP SUBSCRIPTION "name" ON "db"."rp"`,
			stmt: &influxql.DropSubscriptionStatement{
				Name:            "name",
				Database:        "db",
				RetentionPolicy: "rp",
			},
		},

		// SHOW SUBSCRIPTIONS
		{
			s:    `SHOW SUBSCRIPTIONS`,
			stmt: &influxql.ShowSubscriptionsStatement{},
		},

		// Errors
		{s: ``, err: `found EOF, expected SELECT, DELETE, SHOW, CREATE, DROP, EXPLAIN, GRANT, REVOKE, ALTER, SET, KILL at line 1, char 1`},
		{s: `SELECT`, err: `found EOF, expected identifier, string, number, bool at line 1, char 8`},
		{s: `blah blah`, err: `found blah, expected SELECT, DELETE, SHOW, CREATE, DROP, EXPLAIN, GRANT, REVOKE, ALTER, SET, KILL at line 1, char 1`},
		{s: `SELECT field1 X`, err: `found X, expected FROM at line 1, char 15`},
		{s: `SELECT field1 FROM "series" WHERE X +;`, err: `found ;, expected identifier, string, number, bool at line 1, char 38`},
		{s: `SELECT field1 FROM myseries GROUP`, err: `found EOF, expected BY at line 1, char 35`},
		{s: `SELECT field1 FROM myseries LIMIT`, err: `found EOF, expected integer at line 1, char 35`},
		{s: `SELECT field1 FROM myseries LIMIT 10.5`, err: `found 10.5, expected integer at line 1, char 35`},
		{s: `SELECT field1 FROM myseries OFFSET`, err: `found EOF, expected integer at line 1, char 36`},
		{s: `SELECT field1 FROM myseries OFFSET 10.5`, err: `found 10.5, expected integer at line 1, char 36`},
		{s: `SELECT field1 FROM myseries ORDER`, err: `found EOF, expected BY at line 1, char 35`},
		{s: `SELECT field1 FROM myseries ORDER BY`, err: `found EOF, expected identifier, ASC, DESC at line 1, char 38`},
		{s: `SELECT field1 FROM myseries ORDER BY /`, err: `found /, expected identifier, ASC, DESC at line 1, char 38`},
		{s: `SELECT field1 FROM myseries ORDER BY 1`, err: `found 1, expected identifier, ASC, DESC at line 1, char 38`},
		{s: `SELECT field1 FROM myseries ORDER BY time ASC,`, err: `found EOF, expected identifier at line 1, char 47`},
		{s: `SELECT field1 FROM myseries ORDER BY time, field1`, err: `only ORDER BY time supported at this time`},
		{s: `SELECT field1 AS`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `SELECT field1 FROM 12`, err: `found 12, expected identifier at line 1, char 20`},
		{s: `SELECT 1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 FROM myseries`, err: `unable to parse integer at line 1, char 8`},
		{s: `SELECT 10.5h FROM myseries`, err: `found h, expected FROM at line 1, char 12`},
		{s: `SELECT distinct FROM myseries`, err: `found FROM, expected identifier at line 1, char 17`},
		{s: `SELECT count(distinct) FROM myseries`, err: `found ), expected (, identifier at line 1, char 22`},
		{s: `SELECT field1 from myseries WHERE host =~ 'asd' LIMIT 1`, err: `found asd, expected regex at line 1, char 42`},
		{s: `SELECT value > 2 FROM cpu`, err: `invalid operator > in SELECT clause at line 1, char 8; operator is intended for WHERE clause`},
		{s: `SELECT value = 2 FROM cpu`, err: `invalid operator = in SELECT clause at line 1, char 8; operator is intended for WHERE clause`},
		{s: `SELECT s =~ /foo/ FROM cpu`, err: `invalid operator =~ in SELECT clause at line 1, char 8; operator is intended for WHERE clause`},
		{s: `SELECT mean(value) FROM cpu FILL + value`, err: `fill must be a function call`},
		// See issues https://github.com/influxdata/influxdb/issues/1647
		// and https://github.com/influxdata/influxdb/issues/4404
		//{s: `DELETE`, err: `found EOF, expected FROM at line 1, char 8`},
		//{s: `DELETE FROM`, err: `found EOF, expected identifier at line 1, char 13`},
		//{s: `DELETE FROM myseries WHERE`, err: `found EOF, expected identifier, string, number, bool at line 1, char 28`},
		{s: `DELETE`, err: `found EOF, expected FROM, WHERE at line 1, char 8`},
		{s: `DELETE FROM`, err: `found EOF, expected identifier at line 1, char 13`},
		{s: `DELETE FROM myseries WHERE`, err: `found EOF, expected identifier, string, number, bool at line 1, char 28`},
		{s: `DELETE FROM "foo".myseries`, err: `retention policy not supported at line 1, char 1`},
		{s: `DELETE FROM foo..myseries`, err: `database not supported at line 1, char 1`},
		{s: `DROP MEASUREMENT`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `DROP SERIES`, err: `found EOF, expected FROM, WHERE at line 1, char 13`},
		{s: `DROP SERIES FROM`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `DROP SERIES FROM src WHERE`, err: `found EOF, expected identifier, string, number, bool at line 1, char 28`},
		{s: `DROP SERIES FROM "foo".myseries`, err: `retention policy not supported at line 1, char 1`},
		{s: `DROP SERIES FROM foo..myseries`, err: `database not supported at line 1, char 1`},
		{s: `SHOW CONTINUOUS`, err: `found EOF, expected QUERIES at line 1, char 17`},
		{s: `SHOW RETENTION`, err: `found EOF, expected POLICIES at line 1, char 16`},
		{s: `SHOW RETENTION ON`, err: `found ON, expected POLICIES at line 1, char 16`},
		{s: `SHOW RETENTION POLICIES ON`, err: `found EOF, expected identifier at line 1, char 28`},
		{s: `SHOW SHARD`, err: `found EOF, expected GROUPS at line 1, char 12`},
		{s: `SHOW FOO`, err: `found FOO, expected CONTINUOUS, DATABASES, DIAGNOSTICS, FIELD, GRANTS, MEASUREMENT, MEASUREMENTS, QUERIES, RETENTION, SERIES, SHARD, SHARDS, STATS, SUBSCRIPTIONS, TAG, USERS at line 1, char 6`},
		{s: `SHOW STATS FOR`, err: `found EOF, expected string at line 1, char 16`},
		{s: `SHOW DIAGNOSTICS FOR`, err: `found EOF, expected string at line 1, char 22`},
		{s: `SHOW GRANTS`, err: `found EOF, expected FOR at line 1, char 13`},
		{s: `SHOW GRANTS FOR`, err: `found EOF, expected identifier at line 1, char 17`},
		{s: `DROP CONTINUOUS`, err: `found EOF, expected QUERY at line 1, char 17`},
		{s: `DROP CONTINUOUS QUERY`, err: `found EOF, expected identifier at line 1, char 23`},
		{s: `DROP CONTINUOUS QUERY myquery`, err: `found EOF, expected ON at line 1, char 31`},
		{s: `DROP CONTINUOUS QUERY myquery ON`, err: `found EOF, expected identifier at line 1, char 34`},
		{s: `CREATE CONTINUOUS`, err: `found EOF, expected QUERY at line 1, char 19`},
		{s: `CREATE CONTINUOUS QUERY`, err: `found EOF, expected identifier at line 1, char 25`},
		{s: `CREATE CONTINUOUS QUERY cq ON db RESAMPLE FOR 5s BEGIN SELECT mean(value) INTO cpu_mean FROM cpu GROUP BY time(10s) END`, err: `FOR duration must be >= GROUP BY time duration: must be a minimum of 10s, got 5s`},
		{s: `CREATE CONTINUOUS QUERY cq ON db RESAMPLE EVERY 10s FOR 5s BEGIN SELECT mean(value) INTO cpu_mean FROM cpu GROUP BY time(5s) END`, err: `FOR duration must be >= GROUP BY time duration: must be a minimum of 10s, got 5s`},
		{s: `DROP FOO`, err: `found FOO, expected CONTINUOUS, DATABASE, MEASUREMENT, RETENTION, SERIES, SHARD, SUBSCRIPTION, USER at line 1, char 6`},
		{s: `CREATE FOO`, err: `found FOO, expected CONTINUOUS, DATABASE, USER, RETENTION, SUBSCRIPTION at line 1, char 8`},
		{s: `CREATE DATABASE`, err: `found EOF, expected identifier at line 1, char 17`},
		{s: `CREATE DATABASE "testdb" WITH`, err: `found EOF, expected DURATION, NAME, REPLICATION, SHARD at line 1, char 31`},
		{s: `CREATE DATABASE "testdb" WITH DURATION`, err: `found EOF, expected duration at line 1, char 40`},
		{s: `CREATE DATABASE "testdb" WITH REPLICATION`, err: `found EOF, expected integer at line 1, char 43`},
		{s: `CREATE DATABASE "testdb" WITH NAME`, err: `found EOF, expected identifier at line 1, char 36`},
		{s: `CREATE DATABASE "testdb" WITH SHARD`, err: `found EOF, expected DURATION at line 1, char 37`},
		{s: `DROP DATABASE`, err: `found EOF, expected identifier at line 1, char 15`},
		{s: `DROP RETENTION`, err: `found EOF, expected POLICY at line 1, char 16`},
		{s: `DROP RETENTION POLICY`, err: `found EOF, expected identifier at line 1, char 23`},
		{s: `DROP RETENTION POLICY "1h.cpu"`, err: `found EOF, expected ON at line 1, char 31`},
		{s: `DROP RETENTION POLICY "1h.cpu" ON`, err: `found EOF, expected identifier at line 1, char 35`},
		{s: `DROP USER`, err: `found EOF, expected identifier at line 1, char 11`},
		{s: `DROP SUBSCRIPTION`, err: `found EOF, expected identifier at line 1, char 19`},
		{s: `DROP SUBSCRIPTION "name"`, err: `found EOF, expected ON at line 1, char 25`},
		{s: `DROP SUBSCRIPTION "name" ON `, err: `found EOF, expected identifier at line 1, char 30`},
		{s: `DROP SUBSCRIPTION "name" ON "db"`, err: `found EOF, expected . at line 1, char 33`},
		{s: `DROP SUBSCRIPTION "name" ON "db".`, err: `found EOF, expected identifier at line 1, char 34`},
		{s: `CREATE USER testuser`, err: `found EOF, expected WITH at line 1, char 22`},
		{s: `CREATE USER testuser WITH`, err: `found EOF, expected PASSWORD at line 1, char 27`},
		{s: `CREATE USER testuser WITH PASSWORD`, err: `found EOF, expected string at line 1, char 36`},
		{s: `CREATE USER testuser WITH PASSWORD 'pwd' WITH`, err: `found EOF, expected ALL at line 1, char 47`},
		{s: `CREATE USER testuser WITH PASSWORD 'pwd' WITH ALL`, err: `found EOF, expected PRIVILEGES at line 1, char 51`},
		{s: `CREATE SUBSCRIPTION`, err: `found EOF, expected identifier at line 1, char 21`},
		{s: `CREATE SUBSCRIPTION "name"`, err: `found EOF, expected ON at line 1, char 27`},
		{s: `CREATE SUBSCRIPTION "name" ON `, err: `found EOF, expected identifier at line 1, char 32`},
		{s: `CREATE SUBSCRIPTION "name" ON "db"`, err: `found EOF, expected . at line 1, char 35`},
		{s: `CREATE SUBSCRIPTION "name" ON "db".`, err: `found EOF, expected identifier at line 1, char 36`},
		{s: `CREATE SUBSCRIPTION "name" ON "db"."rp"`, err: `found EOF, expected DESTINATIONS at line 1, char 40`},
		{s: `CREATE SUBSCRIPTION "name" ON "db"."rp" DESTINATIONS`, err: `found EOF, expected ALL, ANY at line 1, char 54`},
		{s: `CREATE SUBSCRIPTION "name" ON "db"."rp" DESTINATIONS ALL `, err: `found EOF, expected string at line 1, char 59`},
		{s: `GRANT`, err: `found EOF, expected READ, WRITE, ALL [PRIVILEGES] at line 1, char 7`},
		{s: `GRANT BOGUS`, err: `found BOGUS, expected READ, WRITE, ALL [PRIVILEGES] at line 1, char 7`},
		{s: `GRANT READ`, err: `found EOF, expected ON at line 1, char 12`},
		{s: `GRANT READ FROM`, err: `found FROM, expected ON at line 1, char 12`},
		{s: `GRANT READ ON`, err: `found EOF, expected identifier at line 1, char 15`},
		{s: `GRANT READ ON TO`, err: `found TO, expected identifier at line 1, char 15`},
		{s: `GRANT READ ON testdb`, err: `found EOF, expected TO at line 1, char 22`},
		{s: `GRANT READ ON testdb TO`, err: `found EOF, expected identifier at line 1, char 25`},
		{s: `GRANT READ TO`, err: `found TO, expected ON at line 1, char 12`},
		{s: `GRANT WRITE`, err: `found EOF, expected ON at line 1, char 13`},
		{s: `GRANT WRITE FROM`, err: `found FROM, expected ON at line 1, char 13`},
		{s: `GRANT WRITE ON`, err: `found EOF, expected identifier at line 1, char 16`},
		{s: `GRANT WRITE ON TO`, err: `found TO, expected identifier at line 1, char 16`},
		{s: `GRANT WRITE ON testdb`, err: `found EOF, expected TO at line 1, char 23`},
		{s: `GRANT WRITE ON testdb TO`, err: `found EOF, expected identifier at line 1, char 26`},
		{s: `GRANT WRITE TO`, err: `found TO, expected ON at line 1, char 13`},
		{s: `GRANT ALL`, err: `found EOF, expected ON, TO at line 1, char 11`},
		{s: `GRANT ALL PRIVILEGES`, err: `found EOF, expected ON, TO at line 1, char 22`},
		{s: `GRANT ALL FROM`, err: `found FROM, expected ON, TO at line 1, char 11`},
		{s: `GRANT ALL PRIVILEGES FROM`, err: `found FROM, expected ON, TO at line 1, char 22`},
		{s: `GRANT ALL ON`, err: `found EOF, expected identifier at line 1, char 14`},
		{s: `GRANT ALL PRIVILEGES ON`, err: `found EOF, expected identifier at line 1, char 25`},
		{s: `GRANT ALL ON TO`, err: `found TO, expected identifier at line 1, char 14`},
		{s: `GRANT ALL PRIVILEGES ON TO`, err: `found TO, expected identifier at line 1, char 25`},
		{s: `GRANT ALL ON testdb`, err: `found EOF, expected TO at line 1, char 21`},
		{s: `GRANT ALL PRIVILEGES ON testdb`, err: `found EOF, expected TO at line 1, char 32`},
		{s: `GRANT ALL ON testdb FROM`, err: `found FROM, expected TO at line 1, char 21`},
		{s: `GRANT ALL PRIVILEGES ON testdb FROM`, err: `found FROM, expected TO at line 1, char 32`},
		{s: `GRANT ALL ON testdb TO`, err: `found EOF, expected identifier at line 1, char 24`},
		{s: `GRANT ALL PRIVILEGES ON testdb TO`, err: `found EOF, expected identifier at line 1, char 35`},
		{s: `GRANT ALL TO`, err: `found EOF, expected identifier at line 1, char 14`},
		{s: `GRANT ALL PRIVILEGES TO`, err: `found EOF, expected identifier at line 1, char 25`},
		{s: `KILL`, err: `found EOF, expected QUERY at line 1, char 6`},
		{s: `KILL QUERY 10s`, err: `found 10s, expected integer at line 1, char 12`},
		{s: `KILL QUERY 4 ON 'host'`, err: `found host, expected identifier at line 1, char 16`},
		{s: `REVOKE`, err: `found EOF, expected READ, WRITE, ALL [PRIVILEGES] at line 1, char 8`},
		{s: `REVOKE BOGUS`, err: `found BOGUS, expected READ, WRITE, ALL [PRIVILEGES] at line 1, char 8`},
		{s: `REVOKE READ`, err: `found EOF, expected ON at line 1, char 13`},
		{s: `REVOKE READ TO`, err: `found TO, expected ON at line 1, char 13`},
		{s: `REVOKE READ ON`, err: `found EOF, expected identifier at line 1, char 16`},
		{s: `REVOKE READ ON FROM`, err: `found FROM, expected identifier at line 1, char 16`},
		{s: `REVOKE READ ON testdb`, err: `found EOF, expected FROM at line 1, char 23`},
		{s: `REVOKE READ ON testdb FROM`, err: `found EOF, expected identifier at line 1, char 28`},
		{s: `REVOKE READ FROM`, err: `found FROM, expected ON at line 1, char 13`},
		{s: `REVOKE WRITE`, err: `found EOF, expected ON at line 1, char 14`},
		{s: `REVOKE WRITE TO`, err: `found TO, expected ON at line 1, char 14`},
		{s: `REVOKE WRITE ON`, err: `found EOF, expected identifier at line 1, char 17`},
		{s: `REVOKE WRITE ON FROM`, err: `found FROM, expected identifier at line 1, char 17`},
		{s: `REVOKE WRITE ON testdb`, err: `found EOF, expected FROM at line 1, char 24`},
		{s: `REVOKE WRITE ON testdb FROM`, err: `found EOF, expected identifier at line 1, char 29`},
		{s: `REVOKE WRITE FROM`, err: `found FROM, expected ON at line 1, char 14`},
		{s: `REVOKE ALL`, err: `found EOF, expected ON, FROM at line 1, char 12`},
		{s: `REVOKE ALL PRIVILEGES`, err: `found EOF, expected ON, FROM at line 1, char 23`},
		{s: `REVOKE ALL TO`, err: `found TO, expected ON, FROM at line 1, char 12`},
		{s: `REVOKE ALL PRIVILEGES TO`, err: `found TO, expected ON, FROM at line 1, char 23`},
		{s: `REVOKE ALL ON`, err: `found EOF, expected identifier at line 1, char 15`},
		{s: `REVOKE ALL PRIVILEGES ON`, err: `found EOF, expected identifier at line 1, char 26`},
		{s: `REVOKE ALL ON FROM`, err: `found FROM, expected identifier at line 1, char 15`},
		{s: `REVOKE ALL PRIVILEGES ON FROM`, err: `found FROM, expected identifier at line 1, char 26`},
		{s: `REVOKE ALL ON testdb`, err: `found EOF, expected FROM at line 1, char 22`},
		{s: `REVOKE ALL PRIVILEGES ON testdb`, err: `found EOF, expected FROM at line 1, char 33`},
		{s: `REVOKE ALL ON testdb TO`, err: `found TO, expected FROM at line 1, char 22`},
		{s: `REVOKE ALL PRIVILEGES ON testdb TO`, err: `found TO, expected FROM at line 1, char 33`},
		{s: `REVOKE ALL ON testdb FROM`, err: `found EOF, expected identifier at line 1, char 27`},
		{s: `REVOKE ALL PRIVILEGES ON testdb FROM`, err: `found EOF, expected identifier at line 1, char 38`},
		{s: `REVOKE ALL FROM`, err: `found EOF, expected identifier at line 1, char 17`},
		{s: `REVOKE ALL PRIVILEGES FROM`, err: `found EOF, expected identifier at line 1, char 28`},
		{s: `CREATE RETENTION`, err: `found EOF, expected POLICY at line 1, char 18`},
		{s: `CREATE RETENTION POLICY`, err: `found EOF, expected identifier at line 1, char 25`},
		{s: `CREATE RETENTION POLICY policy1`, err: `found EOF, expected ON at line 1, char 33`},
		{s: `CREATE RETENTION POLICY policy1 ON`, err: `found EOF, expected identifier at line 1, char 36`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb`, err: `found EOF, expected DURATION at line 1, char 43`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION`, err: `found EOF, expected duration at line 1, char 52`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION bad`, err: `found bad, expected duration at line 1, char 52`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h`, err: `found EOF, expected REPLICATION at line 1, char 54`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION`, err: `found EOF, expected integer at line 1, char 67`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 3.14`, err: `found 3.14, expected integer at line 1, char 67`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 0`, err: `invalid value 0: must be 1 <= n <= 2147483647 at line 1, char 67`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION bad`, err: `found bad, expected integer at line 1, char 67`},
		{s: `CREATE RETENTION POLICY policy1 ON testdb DURATION 1h REPLICATION 2 SHARD DURATION INF`, err: `invalid duration INF for shard duration at line 1, char 84`},
		{s: `ALTER`, err: `found EOF, expected RETENTION at line 1, char 7`},
		{s: `ALTER RETENTION`, err: `found EOF, expected POLICY at line 1, char 17`},
		{s: `ALTER RETENTION POLICY`, err: `found EOF, expected identifier at line 1, char 24`},
		{s: `ALTER RETENTION POLICY policy1`, err: `found EOF, expected ON at line 1, char 32`}, {s: `ALTER RETENTION POLICY policy1 ON`, err: `found EOF, expected identifier at line 1, char 35`},
		{s: `ALTER RETENTION POLICY policy1 ON testdb`, err: `found EOF, expected DURATION, REPLICATION, SHARD, DEFAULT at line 1, char 42`},
		{s: `ALTER RETENTION POLICY policy1 ON testdb REPLICATION 1 REPLICATION 2`, err: `found duplicate REPLICATION option at line 1, char 56`},
		{s: `ALTER RETENTION POLICY policy1 ON testdb DURATION 15251w`, err: `overflowed duration 15251w: choose a smaller duration or INF at line 1, char 51`},
		{s: `ALTER RETENTION POLICY policy1 ON testdb DURATION INF SHARD DURATION INF`, err: `invalid duration INF for shard duration at line 1, char 70`},
		{s: `SET`, err: `found EOF, expected PASSWORD at line 1, char 5`},
		{s: `SET PASSWORD`, err: `found EOF, expected FOR at line 1, char 14`},
		{s: `SET PASSWORD something`, err: `found something, expected FOR at line 1, char 14`},
		{s: `SET PASSWORD FOR`, err: `found EOF, expected identifier at line 1, char 18`},
		{s: `SET PASSWORD FOR dejan`, err: `found EOF, expected = at line 1, char 24`},
		{s: `SET PASSWORD FOR dejan =`, err: `found EOF, expected string at line 1, char 25`},
		{s: `SET PASSWORD FOR dejan = bla`, err: `found bla, expected string at line 1, char 26`},
		{s: `$SHOW$DATABASES`, err: `found $SHOW, expected SELECT, DELETE, SHOW, CREATE, DROP, EXPLAIN, GRANT, REVOKE, ALTER, SET, KILL at line 1, char 1`},
		{s: `SELECT * FROM cpu WHERE "tagkey" = $$`, err: `empty bound parameter`},

		// Create a database with a bound parameter.
		{
			s: `CREATE DATABASE $db`,
			params: map[string]interface{}{
				"db": map[string]interface{}{"identifier": "mydb"},
			},
			stmt: &influxql.CreateDatabaseStatement{
				Name: "mydb",
			},
		},

		// Count records in a measurement.
		{
			s: `SELECT count($value) FROM $m`,
			params: map[string]interface{}{
				"value": map[string]interface{}{"identifier": "my_value"},
				"m":     map[string]interface{}{"identifier": "my_measurement"},
			},
			stmt: &influxql.SelectStatement{
				Fields: []*influxql.Field{{
					Expr: &influxql.Call{
						Name: "count",
						Args: []influxql.Expr{
							&influxql.VarRef{Val: "my_value"},
						}}},
				},
				Sources: []influxql.Source{&influxql.Measurement{Name: "my_measurement"}},
			},
		},

		// Find the last 10 shapes records.
		{
			s: `SELECT * FROM $m LIMIT $limit`,
			params: map[string]interface{}{
				"m":     map[string]interface{}{"identifier": "shapes"},
				"limit": int64(10),
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.Wildcard{},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "shapes"}},
				Limit:   10,
			},
		},

		// Find the last 10 shapes records (advanced syntax).
		{
			s: `SELECT * FROM $m LIMIT $limit`,
			params: map[string]interface{}{
				"m":     map[string]interface{}{"identifier": "shapes"},
				"limit": map[string]interface{}{"integer": json.Number("10")},
			},
			stmt: &influxql.SelectStatement{
				IsRawQuery: true,
				Fields: []*influxql.Field{{
					Expr: &influxql.Wildcard{},
				}},
				Sources: []influxql.Source{&influxql.Measurement{Name: "shapes"}},
				Limit:   10,
			},
		},
	}

	for i, tt := range tests {
		if tt.skip {
			continue
		}
		p := influxql.NewParser(strings.NewReader(tt.s))
		if tt.params != nil {
			p.SetParams(tt.params)
		}
		stmt, err := p.ParseStatement()

		// if it's a CQ, there is a non-exported field that gets memoized during parsing that needs to be set
		if st, ok := stmt.(*influxql.CreateContinuousQueryStatement); ok {
			if st != nil && st.Source != nil {
				tt.stmt.(*influxql.CreateContinuousQueryStatement).Source.GroupByInterval()
			}
		}

		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" {
			if !reflect.DeepEqual(tt.stmt, stmt) {
				t.Logf("\n# %s\nexp=%s\ngot=%s\n", tt.s, mustMarshalJSON(tt.stmt), mustMarshalJSON(stmt))
				t.Logf("\nSQL exp=%s\nSQL got=%s\n", tt.stmt.String(), stmt.String())
				t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
			} else {
				// Attempt to reparse the statement as a string and confirm it parses the same.
				// Skip this if we have some kind of statement with a password since those will never be reparsed.
				switch stmt.(type) {
				case *influxql.CreateUserStatement, *influxql.SetPasswordUserStatement:
					continue
				}

				stmt2, err := influxql.ParseStatement(stmt.String())
				if err != nil {
					t.Errorf("%d. %q: unable to parse statement string: %s", i, stmt.String(), err)
				} else if !reflect.DeepEqual(tt.stmt, stmt2) {
					t.Logf("\n# %s\nexp=%s\ngot=%s\n", tt.s, mustMarshalJSON(tt.stmt), mustMarshalJSON(stmt2))
					t.Logf("\nSQL exp=%s\nSQL got=%s\n", tt.stmt.String(), stmt2.String())
					t.Errorf("%d. %q\n\nstmt reparse mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt2)
				}
			}
		}
	}
}

// Ensure the parser can parse expressions into an AST.
func TestParser_ParseExpr(t *testing.T) {
	var tests = []struct {
		s    string
		expr influxql.Expr
		err  string
	}{
		// Primitives
		{s: `100.0`, expr: &influxql.NumberLiteral{Val: 100}},
		{s: `100`, expr: &influxql.IntegerLiteral{Val: 100}},
		{s: `9223372036854775808`, expr: &influxql.UnsignedLiteral{Val: 9223372036854775808}},
		{s: `-9223372036854775808`, expr: &influxql.IntegerLiteral{Val: -9223372036854775808}},
		{s: `-9223372036854775809`, err: `constant -9223372036854775809 underflows int64`},
		{s: `-100.0`, expr: &influxql.NumberLiteral{Val: -100}},
		{s: `-100`, expr: &influxql.IntegerLiteral{Val: -100}},
		{s: `100.`, expr: &influxql.NumberLiteral{Val: 100}},
		{s: `-100.`, expr: &influxql.NumberLiteral{Val: -100}},
		{s: `.23`, expr: &influxql.NumberLiteral{Val: 0.23}},
		{s: `-.23`, expr: &influxql.NumberLiteral{Val: -0.23}},
		{s: `1s`, expr: &influxql.DurationLiteral{Val: time.Second}},
		{s: `-1s`, expr: &influxql.DurationLiteral{Val: -time.Second}},
		{s: `-+1`, err: `found +, expected identifier, number, duration, ( at line 1, char 2`},
		{s: `'foo bar'`, expr: &influxql.StringLiteral{Val: "foo bar"}},
		{s: `true`, expr: &influxql.BooleanLiteral{Val: true}},
		{s: `false`, expr: &influxql.BooleanLiteral{Val: false}},
		{s: `my_ident`, expr: &influxql.VarRef{Val: "my_ident"}},
		{s: `'2000-01-01 00:00:00'`, expr: &influxql.StringLiteral{Val: "2000-01-01 00:00:00"}},
		{s: `'2000-01-01'`, expr: &influxql.StringLiteral{Val: "2000-01-01"}},

		// Simple binary expression
		{
			s: `1 + 2`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.ADD,
				LHS: &influxql.IntegerLiteral{Val: 1},
				RHS: &influxql.IntegerLiteral{Val: 2},
			},
		},

		// Binary expression with LHS precedence
		{
			s: `1 * 2 + 3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.ADD,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.MUL,
					LHS: &influxql.IntegerLiteral{Val: 1},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Binary expression with RHS precedence
		{
			s: `1 + 2 * 3`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.ADD,
				LHS: &influxql.IntegerLiteral{Val: 1},
				RHS: &influxql.BinaryExpr{
					Op:  influxql.MUL,
					LHS: &influxql.IntegerLiteral{Val: 2},
					RHS: &influxql.IntegerLiteral{Val: 3},
				},
			},
		},

		// Binary expression with LHS precedence
		{
			s: `1 / 2 + 3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.ADD,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.DIV,
					LHS: &influxql.IntegerLiteral{Val: 1},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Binary expression with RHS precedence
		{
			s: `1 + 2 / 3`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.ADD,
				LHS: &influxql.IntegerLiteral{Val: 1},
				RHS: &influxql.BinaryExpr{
					Op:  influxql.DIV,
					LHS: &influxql.IntegerLiteral{Val: 2},
					RHS: &influxql.IntegerLiteral{Val: 3},
				},
			},
		},

		// Binary expression with LHS precedence
		{
			s: `1 % 2 + 3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.ADD,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.MOD,
					LHS: &influxql.IntegerLiteral{Val: 1},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Binary expression with RHS precedence
		{
			s: `1 + 2 % 3`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.ADD,
				LHS: &influxql.IntegerLiteral{Val: 1},
				RHS: &influxql.BinaryExpr{
					Op:  influxql.MOD,
					LHS: &influxql.IntegerLiteral{Val: 2},
					RHS: &influxql.IntegerLiteral{Val: 3},
				},
			},
		},

		// Binary expression with LHS paren group.
		{
			s: `(1 + 2) * 3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.MUL,
				LHS: &influxql.ParenExpr{
					Expr: &influxql.BinaryExpr{
						Op:  influxql.ADD,
						LHS: &influxql.IntegerLiteral{Val: 1},
						RHS: &influxql.IntegerLiteral{Val: 2},
					},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Binary expression with no precedence, tests left associativity.
		{
			s: `1 * 2 * 3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.MUL,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.MUL,
					LHS: &influxql.IntegerLiteral{Val: 1},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Addition and subtraction without whitespace.
		{
			s: `1+2-3`,
			expr: &influxql.BinaryExpr{
				Op: influxql.SUB,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.ADD,
					LHS: &influxql.IntegerLiteral{Val: 1},
					RHS: &influxql.IntegerLiteral{Val: 2},
				},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		{
			s: `time>now()-5m`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.GT,
				LHS: &influxql.VarRef{Val: "time"},
				RHS: &influxql.BinaryExpr{
					Op:  influxql.SUB,
					LHS: &influxql.Call{Name: "now"},
					RHS: &influxql.DurationLiteral{Val: 5 * time.Minute},
				},
			},
		},

		// Simple unary expression.
		{
			s: `-value`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.MUL,
				LHS: &influxql.IntegerLiteral{Val: -1},
				RHS: &influxql.VarRef{Val: "value"},
			},
		},

		{
			s: `-mean(value)`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.MUL,
				LHS: &influxql.IntegerLiteral{Val: -1},
				RHS: &influxql.Call{
					Name: "mean",
					Args: []influxql.Expr{
						&influxql.VarRef{Val: "value"}},
				},
			},
		},

		// Unary expressions with parenthesis.
		{
			s: `-(-4)`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.MUL,
				LHS: &influxql.IntegerLiteral{Val: -1},
				RHS: &influxql.ParenExpr{
					Expr: &influxql.IntegerLiteral{Val: -4},
				},
			},
		},

		// Multiplication with leading subtraction.
		{
			s: `-2 * 3`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.MUL,
				LHS: &influxql.IntegerLiteral{Val: -2},
				RHS: &influxql.IntegerLiteral{Val: 3},
			},
		},

		// Binary expression with regex.
		{
			s: `region =~ /us.*/`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.EQREGEX,
				LHS: &influxql.VarRef{Val: "region"},
				RHS: &influxql.RegexLiteral{Val: regexp.MustCompile(`us.*`)},
			},
		},

		// Binary expression with quoted '/' regex.
		{
			s: `url =~ /http\:\/\/www\.example\.com/`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.EQREGEX,
				LHS: &influxql.VarRef{Val: "url"},
				RHS: &influxql.RegexLiteral{Val: regexp.MustCompile(`http\://www\.example\.com`)},
			},
		},

		// Binary expression with quoted '/' regex without space around operator. Influxdb #9058
		{
			s: `url=~/http\:\/\/www\.example\.com/`,
			expr: &influxql.BinaryExpr{
				Op:  influxql.EQREGEX,
				LHS: &influxql.VarRef{Val: "url"},
				RHS: &influxql.RegexLiteral{Val: regexp.MustCompile(`http\://www\.example\.com`)},
			},
		},

		// Complex binary expression.
		{
			s: `value + 3 < 30 AND 1 + 2 OR true`,
			expr: &influxql.BinaryExpr{
				Op: influxql.OR,
				LHS: &influxql.BinaryExpr{
					Op: influxql.AND,
					LHS: &influxql.BinaryExpr{
						Op: influxql.LT,
						LHS: &influxql.BinaryExpr{
							Op:  influxql.ADD,
							LHS: &influxql.VarRef{Val: "value"},
							RHS: &influxql.IntegerLiteral{Val: 3},
						},
						RHS: &influxql.IntegerLiteral{Val: 30},
					},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.ADD,
						LHS: &influxql.IntegerLiteral{Val: 1},
						RHS: &influxql.IntegerLiteral{Val: 2},
					},
				},
				RHS: &influxql.BooleanLiteral{Val: true},
			},
		},

		// Complex binary expression.
		{
			s: `time > now() - 1d AND time < now() + 1d`,
			expr: &influxql.BinaryExpr{
				Op: influxql.AND,
				LHS: &influxql.BinaryExpr{
					Op:  influxql.GT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.SUB,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: mustParseDuration("1d")},
					},
				},
				RHS: &influxql.BinaryExpr{
					Op:  influxql.LT,
					LHS: &influxql.VarRef{Val: "time"},
					RHS: &influxql.BinaryExpr{
						Op:  influxql.ADD,
						LHS: &influxql.Call{Name: "now"},
						RHS: &influxql.DurationLiteral{Val: mustParseDuration("1d")},
					},
				},
			},
		},

		// Duration math with an invalid literal.
		{
			s:   `time > now() - 1y`,
			err: `invalid duration`,
		},

		// Function call (empty)
		{
			s: `my_func()`,
			expr: &influxql.Call{
				Name: "my_func",
			},
		},

		// Function call (multi-arg)
		{
			s: `my_func(1, 2 + 3)`,
			expr: &influxql.Call{
				Name: "my_func",
				Args: []influxql.Expr{
					&influxql.IntegerLiteral{Val: 1},
					&influxql.BinaryExpr{
						Op:  influxql.ADD,
						LHS: &influxql.IntegerLiteral{Val: 2},
						RHS: &influxql.IntegerLiteral{Val: 3},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		expr, err := influxql.NewParser(strings.NewReader(tt.s)).ParseExpr()
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.expr, expr) {
			t.Errorf("%d. %q\n\nexpr mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.expr, expr)
		} else if err == nil {
			// Attempt to reparse the expr as a string and confirm it parses the same.
			expr2, err := influxql.ParseExpr(expr.String())
			if err != nil {
				t.Errorf("%d. %q: unable to parse expr string: %s", i, expr.String(), err)
			} else if !reflect.DeepEqual(tt.expr, expr2) {
				t.Logf("\n# %s\nexp=%s\ngot=%s\n", tt.s, mustMarshalJSON(tt.expr), mustMarshalJSON(expr2))
				t.Logf("\nSQL exp=%s\nSQL got=%s\n", tt.expr.String(), expr2.String())
				t.Errorf("%d. %q\n\nexpr reparse mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.expr, expr2)
			}
		}
	}
}

// Ensure a time duration can be parsed.
func TestParseDuration(t *testing.T) {
	var tests = []struct {
		s   string
		d   time.Duration
		err string
	}{
		{s: `10ns`, d: 10},
		{s: `10u`, d: 10 * time.Microsecond},
		{s: `10µ`, d: 10 * time.Microsecond},
		{s: `15ms`, d: 15 * time.Millisecond},
		{s: `100s`, d: 100 * time.Second},
		{s: `2m`, d: 2 * time.Minute},
		{s: `2h`, d: 2 * time.Hour},
		{s: `2d`, d: 2 * 24 * time.Hour},
		{s: `2w`, d: 2 * 7 * 24 * time.Hour},
		{s: `1h30m`, d: time.Hour + 30*time.Minute},
		{s: `30ms3000u`, d: 30*time.Millisecond + 3000*time.Microsecond},
		{s: `-5s`, d: -5 * time.Second},
		{s: `-5m30s`, d: -5*time.Minute - 30*time.Second},

		{s: ``, err: "invalid duration"},
		{s: `3`, err: "invalid duration"},
		{s: `1000`, err: "invalid duration"},
		{s: `w`, err: "invalid duration"},
		{s: `ms`, err: "invalid duration"},
		{s: `1.2w`, err: "invalid duration"},
		{s: `10x`, err: "invalid duration"},
		{s: `10n`, err: "invalid duration"},
	}

	for i, tt := range tests {
		d, err := influxql.ParseDuration(tt.s)
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.d != d {
			t.Errorf("%d. %q\n\nduration mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.d, d)
		}
	}
}

// Ensure a time duration can be formatted.
func TestFormatDuration(t *testing.T) {
	var tests = []struct {
		d time.Duration
		s string
	}{
		{d: 3 * time.Microsecond, s: `3u`},
		{d: 1001 * time.Microsecond, s: `1001u`},
		{d: 15 * time.Millisecond, s: `15ms`},
		{d: 100 * time.Second, s: `100s`},
		{d: 2 * time.Minute, s: `2m`},
		{d: 2 * time.Hour, s: `2h`},
		{d: 2 * 24 * time.Hour, s: `2d`},
		{d: 2 * 7 * 24 * time.Hour, s: `2w`},
	}

	for i, tt := range tests {
		s := influxql.FormatDuration(tt.d)
		if tt.s != s {
			t.Errorf("%d. %v: mismatch: %s != %s", i, tt.d, tt.s, s)
		}
	}
}

// Ensure a string can be quoted.
func TestQuote(t *testing.T) {
	for i, tt := range []struct {
		in  string
		out string
	}{
		{``, `''`},
		{`foo`, `'foo'`},
		{"foo\nbar", `'foo\nbar'`},
		{`foo bar\\`, `'foo bar\\\\'`},
		{`'foo'`, `'\'foo\''`},
	} {
		if out := influxql.QuoteString(tt.in); tt.out != out {
			t.Errorf("%d. %s: mismatch: %s != %s", i, tt.in, tt.out, out)
		}
	}
}

// Ensure an identifier's segments can be quoted.
func TestQuoteIdent(t *testing.T) {
	for i, tt := range []struct {
		ident []string
		s     string
	}{
		{[]string{``}, `""`},
		{[]string{`select`}, `"select"`},
		{[]string{`in-bytes`}, `"in-bytes"`},
		{[]string{`foo`, `bar`}, `"foo".bar`},
		{[]string{`foo`, ``, `bar`}, `"foo"..bar`},
		{[]string{`foo bar`, `baz`}, `"foo bar".baz`},
		{[]string{`foo.bar`, `baz`}, `"foo.bar".baz`},
		{[]string{`foo.bar`, `rp`, `baz`}, `"foo.bar"."rp".baz`},
		{[]string{`foo.bar`, `rp`, `1baz`}, `"foo.bar"."rp"."1baz"`},
	} {
		if s := influxql.QuoteIdent(tt.ident...); tt.s != s {
			t.Errorf("%d. %s: mismatch: %s != %s", i, tt.ident, tt.s, s)
		}
	}
}

// Ensure DeleteSeriesStatement can convert to a string
func TestDeleteSeriesStatement_String(t *testing.T) {
	var tests = []struct {
		s    string
		stmt influxql.Statement
	}{
		{
			s:    `DELETE FROM src`,
			stmt: &influxql.DeleteSeriesStatement{Sources: []influxql.Source{&influxql.Measurement{Name: "src"}}},
		},
		{
			s: `DELETE FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DeleteSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DELETE FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DeleteSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DELETE WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DeleteSeriesStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
	}

	for _, test := range tests {
		s := test.stmt.String()
		if s != test.s {
			t.Errorf("error rendering string. expected %s, actual: %s", test.s, s)
		}
	}
}

// Ensure DropSeriesStatement can convert to a string
func TestDropSeriesStatement_String(t *testing.T) {
	var tests = []struct {
		s    string
		stmt influxql.Statement
	}{
		{
			s:    `DROP SERIES FROM src`,
			stmt: &influxql.DropSeriesStatement{Sources: []influxql.Source{&influxql.Measurement{Name: "src"}}},
		},
		{
			s: `DROP SERIES FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DropSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DROP SERIES FROM src WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DropSeriesStatement{
				Sources: []influxql.Source{&influxql.Measurement{Name: "src"}},
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
		{
			s: `DROP SERIES WHERE host = 'hosta.influxdb.org'`,
			stmt: &influxql.DropSeriesStatement{
				Condition: &influxql.BinaryExpr{
					Op:  influxql.EQ,
					LHS: &influxql.VarRef{Val: "host"},
					RHS: &influxql.StringLiteral{Val: "hosta.influxdb.org"},
				},
			},
		},
	}

	for _, test := range tests {
		s := test.stmt.String()
		if s != test.s {
			t.Errorf("error rendering string. expected %s, actual: %s", test.s, s)
		}
	}
}

func BenchmarkParserParseStatement(b *testing.B) {
	b.ReportAllocs()
	s := `SELECT "field" FROM "series" WHERE value > 10`
	for i := 0; i < b.N; i++ {
		if stmt, err := influxql.NewParser(strings.NewReader(s)).ParseStatement(); err != nil {
			b.Fatalf("unexpected error: %s", err)
		} else if stmt == nil {
			b.Fatalf("expected statement: %s", stmt)
		}
	}
	b.SetBytes(int64(len(s)))
}

// MustParseSelectStatement parses a select statement. Panic on error.
func MustParseSelectStatement(s string) *influxql.SelectStatement {
	stmt, err := influxql.NewParser(strings.NewReader(s)).ParseStatement()
	if err != nil {
		panic(err)
	}
	return stmt.(*influxql.SelectStatement)
}

// MustParseExpr parses an expression. Panic on error.
func MustParseExpr(s string) influxql.Expr {
	expr, err := influxql.NewParser(strings.NewReader(s)).ParseExpr()
	if err != nil {
		panic(err)
	}
	return expr
}

// errstring converts an error to its string representation.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// newAlterRetentionPolicyStatement creates an initialized AlterRetentionPolicyStatement.
func newAlterRetentionPolicyStatement(name string, DB string, d, sd time.Duration, replication int, dfault bool) *influxql.AlterRetentionPolicyStatement {
	stmt := &influxql.AlterRetentionPolicyStatement{
		Name:     name,
		Database: DB,
		Default:  dfault,
	}

	if d > -1 {
		stmt.Duration = &d
	}

	if sd > -1 {
		stmt.ShardGroupDuration = &sd
	}

	if replication > -1 {
		stmt.Replication = &replication
	}

	return stmt
}

// mustMarshalJSON encodes a value to JSON.
func mustMarshalJSON(v interface{}) []byte {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return b
}

func mustParseDuration(s string) time.Duration {
	d, err := influxql.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustLoadLocation(s string) *time.Location {
	l, err := time.LoadLocation(s)
	if err != nil {
		panic(err)
	}
	return l
}

var LosAngeles = mustLoadLocation("America/Los_Angeles")

func duration(v time.Duration) *time.Duration {
	return &v
}

func intptr(v int) *int {
	return &v
}
