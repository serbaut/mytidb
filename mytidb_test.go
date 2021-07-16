package mytidb_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/serbaut/mytidb"
)

func TestDriver(t *testing.T) {
	db, err := sql.Open("mytidb", "root@(127.0.0.1:14000)/test?parseTime=true")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	now := time.Now().UTC().Round(time.Millisecond)
	var version string
	err = db.QueryRow(`select version()`).Scan(&version)
	assert.NoError(t, err)
	assert.Equal(t, "5.7.25-TiDB-None", version) // FIXME: is there some setup missing?

	_, err = db.Exec(`drop table if exists foo`)
	assert.NoError(t, err)
	_, err = db.Exec(`create table foo (id integer primary key not null, name varchar(255), date datetime(6))`)
	assert.NoError(t, err)
	_, err = db.Exec(`insert into foo (id, name, date) values (?, ?, ?)`, 1, "zz1", now)
	assert.NoError(t, err)

	var id int
	var name string
	var date time.Time
	err = db.QueryRow("select id, name, date from foo where id=?", 1).Scan(&id, &name, &date)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "zz1", name)
	assert.Equal(t, now, date)
}
test
