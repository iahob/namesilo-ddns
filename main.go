package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	fVersion  = flag.Bool("v", false, "print version information and exit")
	fKey      = flag.String("k", "12345", " To be replaced by your unique API key. Visit the API Manager page within your account for details.")
	fDomain   = flag.String("d", "namesilo.com", "The domain associated with the DNS resource record to modify")
	fHost     = flag.String("h", "www", "The hostname to use (there is no need to include the \".DOMAIN\")")
	fInterval = flag.Int64("i", 300, "The seconds of updating interval")
	fConf     = flag.String("f", "", "config file *.conf")
)

var version = "1.0.0"

type ServerParam struct {
	Key      string `json:"key,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Host     string `json:"host,omitempty"`
	Interval int64  `json:"interval,omitempty"`
}

func main() {
	flag.Parse()
	if *fVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	sp := new(ServerParam)
	if len(*fConf) == 0 {
		sp.Domain = *fDomain
		sp.Key = *fKey
		sp.Host = *fKey
		sp.Interval = *fInterval
	} else {
		file, err := os.Open(*fConf)
		err = json.NewDecoder(file).Decode(sp)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}(file)
		err = json.NewDecoder(file).Decode(sp)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	updateDNSLoop(sp)
}

func updateDNSLoop(sp *ServerParam) {
	tick := time.NewTicker(time.Duration(sp.Interval) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := doUpdateDNS(sp.Domain, sp.Host, sp.Key)
			if err != nil {
				log.Printf("[%v] update DNS record failed, domain:%s, host:%s, error:%v", time.Now().Format(time.RFC3339), *fDomain, *fHost, err)
			}
		}
	}
}

func doUpdateDNS(domain, host, key string) error {
	listResp, err := dnsList(domain, key)
	if err != nil {
		return err
	}

	// find the one need to be updated
	fullHost := host + "." + domain
	for _, item := range listResp.ListReply.DNSRecords {
		if item.Host == fullHost {
			if item.Value != listResp.Request.IP {
				// update record
				log.Printf("[%v] host: %s, update old IP: %s to new IP: %s", time.Now().Format(time.RFC3339), item.Host, item.Value, listResp.Request.IP)
				err = dnsUpdate(key, domain, item.RecordID, host, listResp.Request.IP)
				if err != nil {
					return err
				}
			} else {
				log.Printf("[%v] host: %s, old IP: %s, new IP: %s are same, nothing to do", time.Now().Format(time.RFC3339), item.Host, item.Value, listResp.Request.IP)
			}
			return nil
		}
	}
	return nil
}
