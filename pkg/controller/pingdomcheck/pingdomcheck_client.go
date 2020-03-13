package pingdomcheck

import (
        "github.com/russellcardullo/go-pingdom/pingdom"
        "net/url"
        "github.com/go-logr/logr"
)

const (
        httpCheckResolution = 5
)

// Interface definition
type PingdomClient interface {
  	CreateHttpPingdomCheck(reqLogger logr.Logger, name string, url string) (int, error)
  	UpdateHttpPingdomCheck(reqLogger logr.Logger, ID int, name string, url string) error 
  	DeleteHttpPingdomCheck(reqLogger logr.Logger, ID int) error
}

// Implementation based on russellcardullo pingdom client
type RCPingdomClient struct {
	innerClient *pingdom.Client
}

// Get instance
func NewRCPingdomClient(user string, password string, apikey string) (PingdomClient, error){
        innerClient, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
                User:     user,
                Password: password,
                APIKey:   apikey,
        })
        client := &RCPingdomClient{
		innerClient: innerClient,
	}
        return client, err
}

// Create new http pingdomcheck
func (rc *RCPingdomClient) CreateHttpPingdomCheck(reqLogger logr.Logger, name string, Url string) (int, error) {
        parsedUrl, err := url.Parse(Url)
        //Create the http check
        newCheck := pingdom.HttpCheck{Name: name, Hostname: parsedUrl.Host, Resolution: httpCheckResolution}
	log.Info("Calling pingdom API to create a check", " with name ", name, " and url", Url)
        check, err := rc.innerClient.Checks.Create(&newCheck)
        return check.ID, err
}

// Update http pingdomcheck
func (rc *RCPingdomClient) UpdateHttpPingdomCheck(reqLogger logr.Logger, ID int, name string, Url string) error {
        //Get check resource from pingdom
 	log.Info("Calling pingdom API to get check", " with ID ", ID)
        check, err := rc.innerClient.Checks.Read(ID)
        if err != nil {
                return err
        }
        parsedUrl, err := url.Parse(Url)
        if check.Name != name || check.Hostname != parsedUrl.Host {
                //Update required
		log.Info("Calling pingdom API to update check", " with ID ", ID)
                updatedCheck := pingdom.HttpCheck{Name: name, Hostname: parsedUrl.Host, Resolution: httpCheckResolution}
                _ , err := rc.innerClient.Checks.Update(ID, &updatedCheck)
		return err
        } else {
                //No updated required
                log.Info("No update action required for check:", "with ID", ID)
		return nil
        }
}

// Delete http pingdomcheck
func (rc *RCPingdomClient) DeleteHttpPingdomCheck(reqLogger logr.Logger, ID int) error {
	log.Info("Calling pingdom API to delete check", " with ID ", ID)
        _ , err := rc.innerClient.Checks.Delete(ID)
	return err
}

