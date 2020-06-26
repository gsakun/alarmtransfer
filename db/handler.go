package db

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	models "github.com/gsakun/alarmtransfer/model"
	log "github.com/sirupsen/logrus"
)

type Alert struct {
	ID           int    `gorm:"size:11;primary_key;AUTO_INCREMENT;not null" json:"id"`
	AlertName    string `gorm:"size:4;not null" json:"alert_name"`
	AlertSrc     string `gorm:"size:255;not null" json:"alert_src"`
	AlertSrcType int    `gorm:"size:4;DEFAULT NULL" json:"alert_src_type"`
	AlertType    int    `gorm:"size:255;not null" json:"alert_type"`
	AlertLevel   int    `gorm:"size:4;DEFAULT NULL" json:"alert_level"`
	AlertState   int    `gorm:"size:4;DEFAULT NULL" json:"alert_state"`
	DateSubmit   string `gorm:"size:255;not null" json:"date_submit"`
	DateHandle   string `gorm:"size:255;DEFAULT NULL" json:"date_handle"`
	Description  string `gorm:"size:255;not null" json:"description"`
	HandleMethod string `gorm:"size:255;DEFAULT NULL" json:"handle_method"`
	UUID         string `gorm:"size:255;not null" json:"uuid"`
	AlertCount   int    `gorm:"size:4;DEFAULT NULL" json:"alert_count"`
	System       string `gorm:"size:4;DEFAULT NULL" json:"system"`
	ZoneID       int    `gorm:"size:11";not null" json:"zone_id"`
	DataCenterID int    `gorm:"size:11";not null" json:"data_center_id"`
}

func md5V3(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}

//HandleMessage use for handler alertmanager alarm message
func HandleMessage(messages models.WebhookMessage) error {
	log.Infof("Message: %v", messages)
	var errinfo map[string]string = make(map[string]string)
	for _, i := range messages.Alerts {
		if i.Status == "firing" {
			alert := new(Alert)
			alert.AlertName = i.Labels["alertname"]
			alarmsourcetypename := i.Labels["alert_source_type"]
			if alarmsourcetypename == "" {
				log.Errorf("This alert %v data can't be analysis, can't get sourcetype field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get sourcetype file in labels", i)
				continue
			}
			alarmsourcetype := AlertSourceTypemap[alarmsourcetypename]
			if alarmsourcetype == 1 || alarmsourcetype == 2 {
				alert.AlertSrc = strings.Split(i.Labels["instance"], ":")[0]
			} else if alarmsourcetype == 3 {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s", i.Labels["cluster"])
			} else if alarmsourcetype == 4 {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s-%s-pod-%s", i.Labels["cluster"], i.Labels["namespace"], i.Labels["pod_name"])
			} else {
				alert.AlertSrc = i.Labels["instance"]
			}
			alert.AlertSrcType = alarmsourcetype
			if i.Labels["alert_source_type"] == "k8s" || i.Labels["alert_source_type"] == "pod" {
				if i.Labels["user"] != "" {
					alert.System = i.Labels["user"]
				}
				alert.System = i.Labels["cluster"]
			}
			if i.Labels["alert_source_type"] == "kvm" {
				if i.Labels["tenant"] != "" {
					alert.System = i.Labels["tenant"]
				}
			}
			zone := i.Labels["region"]
			if zone == "" {
				log.Errorf("This alert %v data can't be analysis, can't get region field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get region field in labels", i)
				continue
			}
			alert.ZoneID = ZoneInfomap[zone]
			datacenter := i.Labels["datacenter"]
			if datacenter == "" {
				log.Errorf("This alert %v data can't be analysis, can't get datacenter field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get datacenter field in labels", i)
				continue
			}
			datacenterid := DataCentermap[datacenter]
			alert.DataCenterID = datacenterid
			alertlevel := i.Labels["alert_level"]
			if alertlevel == "" {
				log.Errorf("This alert %v data can't be analysis, can't get alert_level field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get alert_level field in labels", i)
				continue
			}
			alert.AlertLevel = AlertLevelmap[alertlevel]
			alerttype := i.Labels["alert_type"]
			if alerttype == "" {
				log.Errorf("This alert %v data can't be analysis, can't get alert_type field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get alert_type field in labels", i)
				continue
			}
			alert.AlertType = AlertTypemap[alerttype]
			l, _ := time.LoadLocation("Asia/Shanghai")
			alert.DateSubmit = i.StartsAt.In(l).Format("2006-01-02 15:04:05")
			description := fmt.Sprintf("%s-%s", i.Annotations["description"], i.Annotations["summary"])
			alert.Description = description
			alert.UUID = md5V3(fmt.Sprintf("%s-%s", alert.AlertSrc, alert.AlertName))
			alert.AlertState = 0
			log.Infof("Start HandleMessage %s-%s", alert.AlertSrc, alert.AlertName)
			err := handleralert(alert)
			if err != nil {
				log.Errorf("%s-%s commitorupdate err %v", alert.AlertSrc, alert.AlertName, err)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data insert failed errinfo %v", i, err)
			} else {
				log.Infof("Insert alert %s-%s success", alert.AlertSrc, alert.AlertName)
			}
		}
		if i.Status == "resolved" {
			alert := new(Alert)
			alert.AlertName = i.Labels["alertname"]
			alarmsourcetypename := i.Labels["alert_source_type"]
			if alarmsourcetypename == "" {
				log.Errorf("This alert %v data can't be analysis, can't get sourcetype field in labels", i)
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data can't be analysis, can't get sourcetype file in labels", i)
				continue
			}
			alarmsourcetype := AlertSourceTypemap[alarmsourcetypename]
			if alarmsourcetype == 1 || alarmsourcetype == 2 {
				alert.AlertSrc = strings.Split(i.Labels["instance"], ":")[0]
			} else if alarmsourcetype == 3 {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s", i.Labels["cluster"])
			} else if alarmsourcetype == 4 {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s-%s-pod-%s", i.Labels["cluster"], i.Labels["namespace"], i.Labels["pod_name"])
			} else {
				alert.AlertSrc = i.Labels["instance"]
			}
			alert.AlertSrcType = alarmsourcetype
			alert.UUID = md5V3(fmt.Sprintf("%s-%s", alert.AlertSrc, alert.AlertName))
			alert.AlertState = 1
			l, _ := time.LoadLocation("Asia/Shanghai")
			alert.DateHandle = i.EndsAt.In(l).Format("2006-01-02 15:04:05")
			err := handleralert(alert)
			if err != nil {
				errinfo[alert.AlertName] = fmt.Sprintf("This alert %v data cancel failed errinfo %v", i, err)
				log.Errorf("%s-%s cancel err %v", alert.AlertSrc, alert.AlertName, err)
			}
			log.Infoln("Cancel Alert success for %s-%s", alert.AlertSrc, alert.AlertName)
		}
	}
	if len(errinfo) != 0 {
		return fmt.Errorf("%v", errinfo)
	}
	return nil
}

func handleralert(alert *Alert) error {
	var id int = 0
	querysql := "select id from alert where uuid=" + "\"" + alert.UUID + "\""
	err := DB.QueryRow(querysql).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("Query DB failed %s-%s %v", alert.AlertSrc, alert.AlertName, err)
			return err
		}
	}
	if id == 0 {
		if alert.AlertState == 0 {
			stmt, err := DB.Prepare(`INSERT alert(alert_src,alert_src_type_id,alert_type_id,alert_level_id,alert_state,date_submit,description,uuid,alert_count,system,zone_id,data_center_id,alert_name) values(?,?,?,?,?,?,?,?,?,?,?,?,?)`)
			if err != nil {
				log.Errorf("Insert prepare err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
			_, err = stmt.Exec(alert.AlertSrc, alert.AlertSrcType, alert.AlertType, alert.AlertLevel, alert.AlertState, alert.DateSubmit, alert.Description, alert.UUID, 1, alert.System, alert.ZoneID, alert.DataCenterID, alert.AlertName)
			if err != nil {
				log.Errorf("Insert exec err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
		}
	} else {
		if alert.AlertState == 0 {
			stmt, err := DB.Prepare(`UPDATE alert set alert_state=?,date_submit=?,description=?,alert_count=alert_count+1 where uuid=?`)
			if err != nil {
				log.Errorf("Update prepare err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
			_, err = stmt.Exec(alert.AlertState, alert.DateSubmit, alert.Description, alert.UUID)
			if err != nil {
				log.Errorf("Update exec err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
		} else {
			stmt, err := DB.Prepare(`UPDATE alert set alert_state=?,date_handle=? where uuid=?`)
			if err != nil {
				log.Errorf("Update prepare err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
			_, err = stmt.Exec(alert.AlertState, alert.DateHandle, alert.UUID)
			if err != nil {
				log.Errorf("Update exec err %s-%s %v", alert.AlertSrc, alert.AlertName, err)
				return err
			}
		}
	}
	return nil
}
