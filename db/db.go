package db

import (
	"time"

	"database/sql"

	"github.com/gsakun/alarmtransfer/config"
	log "github.com/sirupsen/logrus"
)

// DB connection
var DB *sql.DB

// DataCentermap store datacenter info
var DataCentermap map[string]int = make(map[string]int)

// ZoneInfomap store zone info
var ZoneInfomap map[string]int = make(map[string]int)

// Init use for init mysql connection
func Init(conf config.DbConfig) {
	var err error
	DB, err = sql.Open("mysql", conf.Database)
	if err != nil {
		log.Fatalln("open db fail:", err)
	}

	DB.SetMaxIdleConns(conf.Maxidle)
	DB.SetMaxOpenConns(conf.Maxconn)

	err = DB.Ping()
	if err != nil {
		log.Fatalf("ping db fail:", err)
	}
}

// SyncMap use for sync DataCentermap and ZoneInfomap
func SyncMap() {
	go querydatacenter()
	go queryzone()
	time.Sleep(3600 * time.Second)
}

func querydatacenter() {
	sql := "select id, data_center_name from data_center"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Errorf("Query data_center table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id             int
				dataCenterName string
			)

			err = rows.Scan(&id, &dataCenterName)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			DataCentermap[dataCenterName] = id
		}
	}
	log.Infof("sync datacenter table success %v", DataCentermap)
}

func queryzone() {
	sql := "select id, zone_name from zone"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Errorf("Query zone table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id       int
				zoneName string
			)

			err = rows.Scan(&id, &zoneName)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			ZoneInfomap[zoneName] = id
		}
	}
	log.Infof("sync zone table success %v", ZoneInfomap)
}
