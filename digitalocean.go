package dyndns

import (
	"context"
	"strings"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
)

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func digitaloceanupd(configs Config) error {

	tokenSource := &tokenSource{
		AccessToken: configs.Auth.Apikey,
	}
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)
	ctx := context.TODO()

	for _, domainName := range configs.Domains {
		// split domain
		domainArray := strings.Split(domainName, ".")
		domain := domainArray[len(domainArray)-2] + "." + domainArray[len(domainArray)-1]
		var subDomain string
		if len(domainArray) > 2 {
			subDomain = domainName[:len(domainName)-len(domain)-1]
		}
		// get origin records
		records, _, err := client.Domains.Records(ctx, domain, nil)
		if err != nil {
			return err
		}
		var record *godo.DomainRecord
		for i := 0; i < len(records); i++ {
			if records[i].Name == subDomain {
				record = &records[i]
				break
			}
		}
		if record == nil {
			_, _, err = client.Domains.CreateRecord(ctx, domain, &godo.DomainRecordEditRequest{
				Type: "A",
				Name: subDomain,
				Data: configs.Ipupdate,
			})
			if err != nil {
				return err
			}
		} else if configs.Ipupdate != record.Data {
			_, _, err = client.Domains.EditRecord(ctx, domain, record.ID, &godo.DomainRecordEditRequest{
				Type: "A",
				Name: subDomain,
				Data: configs.Ipupdate,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
