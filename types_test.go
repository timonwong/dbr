package dbr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/timonwong/dbr/dialect"
)

var (
	filledRecord = nullTypedRecord{
		StringVal:  NewNullString("wow"),
		Int64Val:   NewNullInt64(4211223344),
		Float64Val: NewNullFloat64(1.618),
		TimeVal:    NewNullTime(time.Date(2009, 1, 3, 18, 15, 5, 0, time.UTC)),
		BoolVal:    NewNullBool(true),
	}
)

func TestNullTypesScanning(t *testing.T) {
	for _, test := range []struct {
		in nullTypedRecord
	}{
		{},
		{
			in: filledRecord,
		},
	} {
		for _, sess := range testSession {
			test.in.Id = nextID()
			_, err := sess.InsertInto("null_types").Columns("id", "string_val", "int64_val", "float64_val", "time_val", "bool_val").Record(test.in).Exec()
			if !assert.NoError(t, err) {
				continue
			}

			var record nullTypedRecord
			err = sess.Select("*").From("null_types").Where(Eq("id", test.in.Id)).LoadStruct(&record)
			if !assert.NoError(t, err) {
				continue
			}

			if sess.Dialect == dialect.PostgreSQL {
				// TODO: https://github.com/lib/pq/issues/329
				if !record.TimeVal.Time.IsZero() {
					record.TimeVal.Time = record.TimeVal.Time.UTC()
				}
			}
			assert.Equal(t, test.in, record)
		}
	}
}

func TestNullTypesJSON(t *testing.T) {
	for _, test := range []struct {
		in   interface{}
		in2  interface{}
		out  interface{}
		want string
	}{
		{
			in:   &filledRecord.BoolVal,
			in2:  filledRecord.BoolVal,
			out:  new(NullBool),
			want: "true",
		},
		{
			in:   new(NullBool),
			in2:  NullBool{},
			out:  new(NullBool),
			want: "null",
		},
		{
			in:   &filledRecord.Float64Val,
			in2:  filledRecord.Float64Val,
			out:  new(NullFloat64),
			want: "1.618",
		},
		{
			in:   new(NullFloat64),
			in2:  NullFloat64{},
			out:  new(NullFloat64),
			want: "null",
		},
		{
			in:   &filledRecord.Int64Val,
			in2:  filledRecord.Int64Val,
			out:  new(NullInt64),
			want: "4211223344",
		},
		{
			in:   new(NullInt64),
			in2:  NullInt64{},
			out:  new(NullInt64),
			want: "null",
		},
		{
			in:   &filledRecord.StringVal,
			in2:  filledRecord.StringVal,
			out:  new(NullString),
			want: `"wow"`,
		},
		{
			in:   new(NullString),
			in2:  NullString{},
			out:  new(NullString),
			want: "null",
		},
		{
			in:   &filledRecord.TimeVal,
			in2:  filledRecord.TimeVal,
			out:  new(NullTime),
			want: `"2009-01-03T18:15:05Z"`,
		},
		{
			in:   new(NullTime),
			in2:  NullTime{},
			out:  new(NullTime),
			want: "null",
		},
	} {
		// marshal ptr
		b, err := json.Marshal(test.in)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, test.want, string(b))

		// marshal value
		b, err = json.Marshal(test.in2)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, test.want, string(b))

		// unmarshal
		err = json.Unmarshal(b, test.out)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, test.in, test.out)
	}
}
