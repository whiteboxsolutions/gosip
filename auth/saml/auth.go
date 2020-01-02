/*
Package saml implements SAML Auth (SharePoint Online user credentials authentication)

This authentication option uses Microsoft Online Security Token Service `https://login.microsoftonline.com/extSTS.srf` and SAML tokens in order to obtain authentication cookie.

Amongst supported platform versions are:
	- SharePoint Online (SPO)
*/
package saml

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/cpass"
)

// AuthCnfg - SAML auth config structure
/* SharePoint Online config sample:
{
  "siteUrl": "https://contoso.sharepoint.com/sites/test",
  "username": "john.doe@contoso.onmicrosoft.com",
  "password": "this-is-not-a-real-password"
}
*/
type AuthCnfg struct {
	SiteURL  string `json:"siteUrl"`  // SPSite or SPWeb URL, which is the context target for the API calls
	Username string `json:"username"` // Username for SharePoint Online, for example `[user]@[company].onmicrosoft.com`
	Password string `json:"password"` // User or App password

	masterKey string
}

// ReadConfig : reads private config with auth options
func (c *AuthCnfg) ReadConfig(privateFile string) error {
	jsonFile, err := os.Open(privateFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &c)

	crypt := cpass.Cpass(c.masterKey)
	pass, err := crypt.Decode(c.Password)
	if err == nil {
		c.Password = pass
	}

	return nil
}

// WriteConfig : writes private config with auth options
func (c *AuthCnfg) WriteConfig(privateFile string) error {
	crypt := cpass.Cpass(c.masterKey)
	pass, err := crypt.Encode(c.Password)
	if err != nil {
		pass = c.Password
	}
	config := &AuthCnfg{
		SiteURL:  c.SiteURL,
		Username: c.Username,
		Password: pass,
	}
	file, _ := json.MarshalIndent(config, "", "  ")
	return ioutil.WriteFile(privateFile, file, 0644)
}

// SetMasterkey : defines custom masterkey
func (c *AuthCnfg) SetMasterkey(masterKey string) {
	c.masterKey = masterKey
}

// GetAuth : authenticates, receives access token
func (c *AuthCnfg) GetAuth() (string, error) {
	return GetAuth(c)
}

// GetSiteURL : gets siteURL
func (c *AuthCnfg) GetSiteURL() string {
	return c.SiteURL
}

// GetStrategy : gets auth strategy name
func (c *AuthCnfg) GetStrategy() string {
	return "saml"
}

// SetAuth : authenticate request
func (c *AuthCnfg) SetAuth(req *http.Request, httpClient *gosip.SPClient) error {
	authCookie, err := c.GetAuth()
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", authCookie)
	return nil
}
