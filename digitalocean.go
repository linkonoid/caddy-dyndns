package dyndns

import (
	"strings"

	cloudfl "github.com/cloudflare/cloudflare-go"
)

func digitaloceanupd(configs Config) error {

	api, err := cloudfl.New(configs.Auth.Apikey, configs.Auth.Email)
	if err != nil {
		return err
	}

	for _, domain := range configs.Domains {

		zoneID := ""
		zones, err := api.ListZones()
		if err != nil {
			return err
		}
		for _, zone := range zones {
			if strings.HasSuffix(domain, zone.Name) {
				zoneID = zone.ID
			}
		}

		recordID := ""
		records, err := api.DNSRecords(zoneID, cloudfl.DNSRecord{})
		if err != nil {
			return err
		}
		for _, r := range records {
			if r.Type == "A" && r.Name == domain {
				recordID = r.ID
			}
		}

		record, err := api.DNSRecord(zoneID, recordID)
		if err != nil {
			return err
		}

		if configs.Ipupdate != record.Content {
			record.Content = configs.Ipupdate
			err = api.UpdateDNSRecord(zoneID, recordID, record)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
