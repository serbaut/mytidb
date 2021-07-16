# mytidb
In-process database/sql driver for tidb

* Storage path is /tmp/mytidb (tidb-server uses /tmp/tidb by default)
* Status port is disabled
* tidb log level is error
* Host/Port is taken from the mysql dsn, e.g. root@tcp(127.0.0.1:14000)/test?parseTime=true
