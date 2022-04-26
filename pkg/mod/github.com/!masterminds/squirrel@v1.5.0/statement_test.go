package squirrel

import (
	"database/sql"
	"testing"

	"github.com/lann/builder"
	"github.com/stretchr/testify/assert"
)

func TestStatementBuilder(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db)

	sb.Select("test").Exec()
	assert.Equal(t, "SELECT test", db.LastExecSql)
}

func TestStatementBuilderPlaceholderFormat(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db).PlaceholderFormat(Dollar)

	sb.Select("test").Where("x = ?").Exec()
	assert.Equal(t, "SELECT test WHERE x = $1", db.LastExecSql)
}

func TestRunWithDB(t *testing.T) {
	db := &sql.DB{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(db))
		builder.GetStruct(Insert("t").RunWith(db))
		builder.GetStruct(Update("t").RunWith(db))
		builder.GetStruct(Delete("t").RunWith(db))
	}, "RunWith(*sql.DB) should not panic")

}

func TestRunWithTx(t *testing.T) {
	tx := &sql.Tx{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(tx))
		builder.GetStruct(Insert("t").RunWith(tx))
		builder.GetStruct(Update("t").RunWith(tx))
		builder.GetStruct(Delete("t").RunWith(tx))
	}, "RunWith(*sql.Tx) should not panic")
}

type fakeBaseRunner struct{}

func (fakeBaseRunner) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (fakeBaseRunner) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func TestRunWithBaseRunner(t *testing.T) {
	sb := StatementBuilder.RunWith(fakeBaseRunner{})
	_, err := sb.Select("test").Exec()
	assert.NoError(t, err)
}

func TestRunWithBaseRunnerQueryRowError(t *testing.T) {
	sb := StatementBuilder.RunWith(fakeBaseRunner{})
	assert.Error(t, RunnerNotQueryRunner, sb.Select("test").QueryRow().Scan(nil))

}

func TestStatementBuilderWhere(t *testing.T) {
	sb := StatementBuilder.Where("x = ?", 1)

	sql, args, err := sb.Select("test").Where("y = ?", 2).ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT test WHERE x = ? AND y = ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2}
	assert.Equal(t, expectedArgs, args)
}
