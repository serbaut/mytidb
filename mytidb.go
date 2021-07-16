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

func (d *mytidb) Open(dsn string) (driver.Conn, error) {
	once.Do(func() {
		myCfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			onceErr = err
			return
		}
		cfg := config.NewConfig()
		cfg.Path = "/tmp/mytidb"
		cfg.Status.ReportStatus = false
		cfg.Log.Level = "fatal"
		if host, sport, err := net.SplitHostPort(myCfg.Addr); err != nil {
			onceErr = err
			return
		} else if iport, err := strconv.Atoi(sport); err != nil {
			onceErr = err
			return
		} else {
			cfg.Host = host
			cfg.Port = uint(iport)
		}
		if onceErr = cfg.Valid(); onceErr != nil {
			return
		}
		if onceErr = logutil.InitZapLogger(cfg.Log.ToLogConfig()); onceErr != nil {
			return
		}
		if onceErr = kvstore.Register("unistore", mockstore.EmbedUnistoreDriver{}); onceErr != nil {
			return
		}
		storage, err := kvstore.New(fmt.Sprintf("%s://%s", cfg.Store, cfg.Path))
		if err != nil {
			onceErr = err
			return
		}
		dom, err := session.BootstrapSession(storage)
		if err != nil {
			onceErr = err
			return
		}
		driver := server.NewTiDBDriver(storage)
		svr, err := server.NewServer(cfg, driver)
		if err != nil {
			onceErr = err
			return
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
	})

	if onceErr != nil {
		return nil, onceErr
	}

	return mysql.MySQLDriver{}.Open(dsn)
}
