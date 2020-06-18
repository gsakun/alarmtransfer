package db

import (
	"time"

	"database/sql"

	log "github.com/sirupsen/logrus"
)

// DB connection
var DB *sql.DB

// DataCentermap store datacenter info
var DataCentermap map[string]int = make(map[string]int)

// ZoneInfomap store zone info
var ZoneInfomap map[string]int = make(map[string]int)

// AlertLevelmap store alertlevel info
var AlertLevelmap map[string]int = make(map[string]int)

// AlertTypemap store alerttype info
var AlertTypemap map[string]int = make(map[string]int)

// AlertSourceTypemap store alertsourcetype info
var AlertSourceTypemap map[string]int = make(map[string]int)

// Init use for init mysql connection
func Init(dbaddress string, maxconn, maxidle int) {
	var err error
	DB, err = sql.Open("mysql", dbaddress)
	if err != nil {
		log.Fatalln("open db fail:", err)
	}

	DB.SetMaxIdleConns(maxidle)
	DB.SetMaxOpenConns(maxconn)

	err = DB.Ping()
	if err != nil {
		log.Fatalf("ping db fail:", err)
	}
}

// SyncMap use for sync DataCentermap and ZoneInfomap
func SyncMap() {
	go querydatacenter()
	go queryzone()
	go queryalertlevel()
	go queryalerttype()
	go queryalertsrctype()
	time.Sleep(600 * time.Second)
}

func queryalertlevel() {
	sql := "select id, alert_level from alert_level"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Errorf("Query alert_level table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id          int
				alert_level string
			)

			err = rows.Scan(&id, &alert_level)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			AlertLevelmap[alert_level] = id
		}
	}
	log.Infof("sync alert_level table success %v", AlertLevelmap)
}

func queryalerttype() {
	sql := "select id, alert_type from alert_level"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Errorf("Query alert_type table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id         int
				alert_type string
			)

			err = rows.Scan(&id, &alert_type)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			AlertTypemap[alert_type] = id
		}
	}
	log.Infof("sync alert_type table success %v", AlertTypemap)
}

func queryalertsrctype() {
	sql := "select id, alert_src_type from alert_level"
	rows, err := DB.Query(sql)
	if err != nil {
		log.Errorf("Query alert_src_type table Failed")
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				id             int
				alert_src_type string
			)

			err = rows.Scan(&id, &alert_src_type)
			if err != nil {
				log.Errorf("ERROR: %v", err)
				continue
			}
			AlertSourceTypemap[alert_src_type] = id
		}
	}
	log.Infof("sync alert_src_type table success %v", AlertSourceTypemap)
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
