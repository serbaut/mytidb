package mytidb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/go-sql-driver/mysql"

	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/server"
	"github.com/pingcap/tidb/session"
	kvstore "github.com/pingcap/tidb/store"
	"github.com/pingcap/tidb/store/mockstore"
	"github.com/pingcap/tidb/util/logutil"
)

type mytidb struct{}

var (
	once    sync.Once
	onceErr error
)

func init() {
	sql.Register("mytidb", &mytidb{})
}

func startTidb(dsn string) error {
	myCfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return err
	}
	cfg := config.NewConfig()
	cfg.Path = "/tmp/mytidb"
	cfg.Status.ReportStatus = false
	cfg.Log.Level = "fatal"
	if host, sport, err := net.SplitHostPort(myCfg.Addr); err != nil {
		return err
	} else if iport, err := strconv.Atoi(sport); err != nil {
		return err
	} else {
		cfg.Host = host
		cfg.Port = uint(iport)
	}
	if err := cfg.Valid(); err != nil {
		return err
	}
	if err := logutil.InitZapLogger(cfg.Log.ToLogConfig()); err != nil {
		return err
	}
	if err := kvstore.Register("unistore", mockstore.EmbedUnistoreDriver{}); err != nil {
		return err
	}
	storage, err := kvstore.New(fmt.Sprintf("%s://%s", cfg.Store, cfg.Path))
	if err != nil {
		return err
	}
	dom, err := session.BootstrapSession(storage)
	if err != nil {
		return err
	}
	driver := server.NewTiDBDriver(storage)
	svr, err := server.NewServer(cfg, driver)
	if err != nil {
		return err
	}

	svr.SetDomain(dom)
	svr.InitGlobalConnID(dom.ServerID)
	go dom.ExpensiveQueryHandle().SetSessionManager(svr).Run()
	dom.InfoSyncer().SetSessionManager(svr)

	go func() {
		if err := svr.Run(); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (d *mytidb) Open(dsn string) (driver.Conn, error) {
	once.Do(func() {
		onceErr = startTidb(dsn)
	})

	if onceErr != nil {
		return nil, onceErr
	}

	return mysql.MySQLDriver{}.Open(dsn)
}
