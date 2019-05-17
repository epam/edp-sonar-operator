package client

import (
	"gopkg.in/resty.v1"
)

type SonarClient struct {
	resty resty.Client
}

func (sc *SonarClient) InitNewRestClient(url string, user string, password string) error {
	sc.resty = *resty.SetHostURL(url).SetBasicAuth(user, password)
	return nil
}
