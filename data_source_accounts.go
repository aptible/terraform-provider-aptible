package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAccounts() *schema.Resource {
	return &schema.Resource{
		Read:   resourceAccountsRead,
		Schema: map[string]*schema.Schema{},
	}
}

func resourceAccountsRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	accounts, err := getAccounts("https://api.aptible.com/accounts", client)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return fmt.Errorf("No accounts exist.")
	}
	return nil
}

func getAccounts(url string, client *http.Client) ([]Account, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := AccountsResponse{}

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}
