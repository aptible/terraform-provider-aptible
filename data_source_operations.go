package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceOperations() *schema.Resource {
	return &schema.Resource{
		Read:   resourceOperationsRead,
		Schema: map[string]*schema.Schema{},
	}
}

func resourceOperationsRead(d *schema.ResourceData, m interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	accounts, err := getAccounts("https://api.aptible.com/accounts", client)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		CreateLogger.Println("No accounts exist.")
	}
	return nil
}


func getOperations(url string, client *http.Client) ([]Account, error) {
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	resp := map[string]interface{}

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return resp.Accounts, nil
}
