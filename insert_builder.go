package dbr

import (
	"database/sql"
	"reflect"
	"context"
)

type InsertBuilder struct {
	runner
	EventReceiver
	Dialect Dialect

	RecordID reflect.Value

	*InsertStmt
}

func (sess *Session) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		InsertStmt:    InsertInto(table),
	}
}

func (tx *Tx) InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		InsertStmt:    InsertInto(table),
	}
}

func (sess *Session) InsertBySql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		runner:        sess,
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		InsertStmt:    InsertBySql(query, value...),
	}
}

func (tx *Tx) InsertBySql(query string, value ...interface{}) *InsertBuilder {
	return &InsertBuilder{
		runner:        tx,
		EventReceiver: tx,
		Dialect:       tx.Dialect,
		InsertStmt:    InsertBySql(query, value...),
	}
}

func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.Column = append(b.Column, column)
	switch len(b.Value) {
	case 0:
		b.InsertStmt.Values(value)
	case 1:
		b.Value[0] = append(b.Value[0], value)
	default:
		panic("pair only allows one record to insert")
	}
	return b
}

func (b *InsertBuilder) ExecContext(ctx context.Context) (sql.Result, error) {
	result, err := exec(ctx, b.runner, b.EventReceiver, b, b.Dialect)
	if err != nil {
		return nil, err
	}

	if b.RecordID.IsValid() {
		if id, err := result.LastInsertId(); err == nil {
			b.RecordID.SetInt(id)
		}
	}

	return result, nil
}

func (b *InsertBuilder) Exec() (sql.Result, error) {
	return b.ExecContext(context.Background())
}

func (b *InsertBuilder) Columns(column ...string) *InsertBuilder {
	b.InsertStmt.Columns(column...)
	return b
}

func (b *InsertBuilder) Record(structValue interface{}) *InsertBuilder {
	v := reflect.Indirect(reflect.ValueOf(structValue))
	if v.Kind() == reflect.Struct && v.CanSet() {
		// ID is recommended by golint here
		for _, name := range []string{"Id", "ID"} {
			field := v.FieldByName(name)
			if field.IsValid() && field.Kind() == reflect.Int64 {
				b.RecordID = field
				break
			}
		}
	}

	b.InsertStmt.Record(structValue)
	return b
}

func (b *InsertBuilder) Values(value ...interface{}) *InsertBuilder {
	b.InsertStmt.Values(value...)
	return b
}

// OnConflictMap allows to add actions for constraint violation, e.g UPSERT
func (b *InsertBuilder) OnConflictMap(constraint string, actions map[string]interface{}) *InsertBuilder {
	b.InsertStmt.OnConflictMap(constraint, actions)
	return b
}
