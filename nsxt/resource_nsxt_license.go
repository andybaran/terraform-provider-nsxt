/* Copyright © 2018 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLicenseCreate,
		Read:   resourceNsxtLicenseRead,
		Update: resourceNsxtLicenseUpdate,
		Delete: resourceNsxtLicenseDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtLicenseImport,
		},

		Schema: map[string]*schema.Schema{
			"license_key": {
				Type:        schema.TypeString,
				Description: "License Key",
				Required:    true,
				// ToDo ValidateFunc: validation.IsLicenseKey(),
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Description: "Accept Eula? True of False",
				Optional:    true,
			},

			// computed properties returned by the API
			"capacity_type": {
				Type:        schema.TypeString,
				Description: "License metrics specifying the capacity type of license key.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "License edition",
				Computed:    true,
			},
			"expiry": {
				Type:        schema.TypeString,
				Description: "Time to expiry (milliseconds since UNIX epoch)",
				Computed:    true,
			},
			"features": {
				Type:        schema.TypeString,
				Description: "Semicolon delimited feature list",
				Computed:    true,
			},
			"is_eval": {
				Type:        schema.TypeBool,
				Description: "True for evalution license",
				Computed:    true,
			},
			"is_expired": {
				Type:        schema.TypeBool,
				Description: "Whether the license has expired",
				Computed:    true,
			},
			"is_mh": {
				Type:        schema.TypeBool,
				Description: "True for multi-hypervisor support",
				Computed:    true,
			},
			"product_name": {
				Type:        schema.TypeString,
				Description: "Product name",
				Computed:    true,
			},
			"product_version": {
				Type:        schema.TypeString,
				Description: "Product Version",
				Computed:    true,
			},
			"quantity": {
				Type:        schema.TypeString,
				Description: "License capacity; 0 for unlimited",
				Computed:    true,
			},
		},
	}
}

func resourceNsxtLicenseCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	l := d.Get("license_key").(string)
	e := d.Get("accept_eula").(bool)

	
	licenseAndEula, resp, err := nsxClient.LicensingApi.CreateLicense()
	
	
	LogicalRoutingAndServicesApi.AddStaticRoute(nsxClient.Context, logicalRouterID, staticRoute)

	if err != nil {
		return fmt.Errorf("Error during StaticRoute create on router %s: %v", logicalRouterID, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during StaticRoute create on router %s: %v", logicalRouterID, resp.StatusCode)
	}
	d.SetId(staticRoute.Id)

	return resourceNsxtLicenseRead(d, m)
}

func resourceNsxtLicenseRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route read")
	}

	staticRoute, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadStaticRoute(nsxClient.Context, logicalRouterID, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] StaticRoute %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during StaticRoute read: %v", err)
	}

	d.Set("revision", staticRoute.Revision)
	d.Set("description", staticRoute.Description)
	d.Set("display_name", staticRoute.DisplayName)
	setTagsInSchema(d, staticRoute.Tags)
	d.Set("logical_router_id", staticRoute.LogicalRouterId)
	d.Set("network", staticRoute.Network)
	err = setNextHopsInSchema(d, staticRoute.NextHops)
	if err != nil {
		return fmt.Errorf("Error during StaticRoute set in schema: %v", err)
	}

	return nil
}

func resourceNsxtLicenseUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route update")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	network := d.Get("network").(string)
	nextHops := getNextHopsFromSchema(d)
	staticRoute := manager.StaticRoute{
		Revision:        revision,
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		LogicalRouterId: logicalRouterID,
		Network:         network,
		NextHops:        nextHops,
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateStaticRoute(nsxClient.Context, logicalRouterID, id, staticRoute)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during StaticRoute update: %v", err)
	}

	return resourceNsxtLicenseRead(d, m)
}

func resourceNsxtLicenseDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route deletion")
	}

	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteStaticRoute(nsxClient.Context, logicalRouterID, id)
	if err != nil {
		return fmt.Errorf("Error during StaticRoute delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] StaticRoute %s for router %s not found", id, logicalRouterID)
		d.SetId("")
	}
	return nil
}

func resourceNsxtLicenseImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <router-id>/<static-route-id> as an input")
	}
	d.SetId(s[1])
	d.Set("logical_router_id", s[0])
	return []*schema.ResourceData{d}, nil
}
