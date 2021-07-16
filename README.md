# mytidb
database/sql driver for embeded tidb using unistore

* tidb 5.1.0
* Storage path is /tmp/mytidb (tidb-server uses /tmp/tidb by default)
* Status port is disabled
* tidb log level is fatal
* Host/Port is taken from the mysql dsn, e.g. root@tcp(127.0.0.1:14000)/test?parseTime=true
* tidb_txn_mode is pessimistic by default
