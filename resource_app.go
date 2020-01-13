package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate, // POST
		Read:   resourceAppRead,   // GET
		Update: resourceAppUpdate, // PUT
		Delete: resourceAppDelete, // DELETE

		Schema: map[string]*schema.Schema{
			"account_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"data": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
			},
		},
	}
}

func resourceAppCreate(d *schema.ResourceData, m interface{}) error {
	// TODO: maybe use "id" instead?
	var client = &http.Client{Timeout: 10 * time.Second}
	account_id := d.Get("account_id").(string)
	handle := d.Get("handle").(string)

	requestBody, err := json.Marshal(map[string]string{
		"handle": handle,
	})
	if err != nil {
		CreateLogger.Println("Error while marshalling JSON.\n[ERROR] -", err)
		return err
	}
	CreateLogger.Println("This is the JSON: ", requestBody.(string))

	url := fmt.Sprintf("https://api-rachel.aptible-sandbox.com/accounts/%s/apps", account_id)

	// Append access token
	var token = "Authorization: " + os.Getenv("AUTH_TOKEN")
	CreateLogger.Println("This is the access token: \n", token)

	// Create a new request using http
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		CreateLogger.Println("Error while creating request.\n[ERROR] -", err)
	}

	// add content type and authorization header to the req
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", token)
	CreateLogger.Println("This is the request sent: ", req)

	resp, err := client.Do(req)
	if err != nil {
		CreateLogger.Println("Error on response.\n[ERROR] -", err)
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	d.Set("data", result)
	CreateLogger.Println("This is the data retrieved: ", result)
	d.SetId(handle)
	return resourceAppRead(d, m)
}

func resourceAppRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceAppRead(d, m)
}

func resourceAppDelete(d *schema.ResourceData, m interface{}) error {
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return nil
}
