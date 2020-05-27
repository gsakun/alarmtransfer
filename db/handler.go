package db

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gsakun/alarmtransfer/types"
	log "github.com/sirupsen/logrus"
)

type Alert struct {
	ID           int    `gorm:"size:11;primary_key;AUTO_INCREMENT;not null" json:"id"`
	AlertName    string `gorm:"size:4;not null" json:"alert_name"`
	AlertSrc     string `gorm:"size:255;not null" json:"alert_src"`
	AlertSrcType int    `gorm:"size:4;DEFAULT NULL" json:"alert_src_type"`
	AlertType    string `gorm:"size:255;not null" json:"alert_type"`
	AlertLevel   int    `gorm:"size:4;DEFAULT NULL" json:"alert_level"`
	AlertState   int    `gorm:"size:4;DEFAULT NULL" json:"alert_state"`
	DateSubmit   string `gorm:"size:255;not null" json:"date_submit"`
	DateHandle   string `gorm:"size:255;DEFAULT NULL" json:"date_handle"`
	Description  string `gorm:"size:255;not null" json:"date_handle"`
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
func HandleMessage(messages types.WebhookMessage) error {

	status := messages.Status
	if status == "firing" {
		for _, i := range messages.Alerts {
			alert := new(Alert)
			alert.AlertName = i.Labels["alertname"]
			alarmsourcetype := i.Labels["sourcetype"]
			if alarmsourcetype == "" {
				log.Errorf("This alert data can't be analysis")
				continue
			}
			if alarmsourcetype == "0" || alarmsourcetype == "1" {
				alert.AlertSrc = strings.Split(i.Labels["instance"], ":")[0]
			} else if alarmsourcetype == "3" {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s", i.Labels["cluster"])
			} else if alarmsourcetype == "2" {
				alert.AlertSrc = fmt.Sprintf("k8s-cluster-%s-%s-pod-%s", i.Labels["cluster"], i.Labels["namespace"], i.Labels["container_name"])
			}

			zone := i.Labels["cluster"]
			if zone == "" {
				zone = i.Labels["tenant"]
				if zone == "" {
					continue
				}
			}
			alert.ZoneID = ZoneInfomap[zone]
			datacenter := i.Labels["datacenter"]
			if datacenter == "" {
				continue
			}
			datacenterid := DataCentermap[datacenter]
			alert.DataCenterID = datacenterid
			alarmlevel := i.Labels["severity"]
			if alarmlevel == "" {
				continue
			}
			alertlevel, err := strconv.Atoi(alarmlevel)
			if err != nil {
				continue
			}
			alert.AlertLevel = alertlevel
			alarmtype := i.Labels["alarmtype"]
			createtime := i.StartsAt
			description := fmt.Sprintf("%s-%s", i.Annotations["description"], i.Annotations["summary"])

		}
	}
}

func insert(alert *Alert) error {
	return nil
}

func update(alert *Alert) error {
	return nil
}
