package anypoint

import (
	"context"
	"io/ioutil"

	amq "../local-mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAMQ() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceENVRead,
		Schema: map[string]*schema.Schema{
			"defaultTtl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"defaultLockTtl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("regionId").(string)

	if envid == "" || orgid == "" {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "ENV id (id) and Organization ID (organization_id) are required",
			Detail:   "ENV id (id) and Organization ID (organization_id) must be provided",
		})
		return diags
	}

	// NEEDS RESOURCE_AMQ ----------------------
	authctx := getENVAuthCtx(ctx, &pco)

	//request env
	res, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsGet(authctx, orgid, envid, regionid).Execute()
	defer httpr.Body.Close()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Get ENV",
			Detail:   details,
		})
		return diags
	}
	//process data
	amqinstance := flattenAMQData(&res)
	//save in data source schema
	// if err := setAMQCoreAttributesToResourceData(d, amqinstance); err != nil {
	// 	diags := append(diags, diag.Diagnostic{
	// 		Severity: diag.Error,
	// 		Summary:  "Unable to set AMQ",
	// 		Detail:   err.Error(),
	// 	})
	// 	return diags
	// }

	// New error catcher
	if err := d.Set("roles", amqinstance); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create queue",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(envid)

	return diags
}

/*
* Copies the given env instance into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param envitem map[string]interface{} the env instance
 */
// func setAMQCoreAttributesToResourceData(d *schema.ResourceData, amqitem map[string]interface{}) error {
// 	attributes := getAMQCoreAttributes()
// 	if amqitem != nil {
// 		for _, attr := range attributes {
// 			if err := d.Set(attr, amqitem[attr]); err != nil {
// 				return fmt.Errorf("unable to set AMQ attribute %s\n details: %s", attr, err)
// 			}
// 		}
// 	}
// 	return nil
// }

// New error catcher

/*
* Transforms a env.Env object to the dataSourceENV schema
* @param envitem *env.Env the env struct
* @return the env mapped struct
 */
func flattenAMQData(amqitem *[]amq.Queue) []interface{} {
	if amqitem != nil && len(*amqitem) > 0 {
		res := make([]interface{}, len(*amqitem))

		for i, items := range *amqitem {
			item := make(map[string]interface{})

			item["defaultTtl"] = items.GetDefaultTtl()
			item["defaultLockTtl"] = items.GetDefaultLockTtl()
			item["type"] = items.GetType()
			item["encrypted"] = items.GetEncrypted()

			res[i] = item
		}

		return res
	}

	return nil
}

func getAMQCoreAttributes() []string {
	attributes := [...]string{
		"name", "organization_id", "is_production", "type", "client_id",
	}
	return attributes[:]
}
