package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	env "github.com/mulesoft-consulting/anypoint-client-go/env"
)

func dataSourceENV() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceENVRead,
		Description: `
		Reads all ` + "`" + `environments` + "`" + ` in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The environment id",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The business group id",
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_production": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceENVRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("id").(string)
	orgid := d.Get("org_id").(string)
	authctx := getENVAuthCtx(ctx, &pco)

	//request env
	res, httpr, err := pco.envclient.DefaultApi.OrganizationsOrgIdEnvironmentsEnvironmentIdGet(authctx, orgid, envid).Execute()
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
	envinstance := flattenENVData(&res)
	//save in data source schema
	if err := setENVCoreAttributesToResourceData(d, envinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set ENV",
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
func setENVCoreAttributesToResourceData(d *schema.ResourceData, envitem map[string]interface{}) error {
	attributes := getENVCoreAttributes()
	if envitem != nil {
		for _, attr := range attributes {
			if err := d.Set(attr, envitem[attr]); err != nil {
				return fmt.Errorf("unable to set ENV attribute %s\n details: %s", attr, err)
			}
		}
	}
	return nil
}

/*
* Transforms a env.Env object to the dataSourceENV schema
* @param envitem *env.Env the env struct
* @return the env mapped struct
 */
func flattenENVData(envitem *env.Env) map[string]interface{} {
	if envitem != nil {
		item := make(map[string]interface{})

		item["id"] = envitem.GetId()
		item["name"] = envitem.GetName()
		item["organization_id"] = envitem.GetOrganizationId()
		item["is_production"] = envitem.GetIsProduction()
		item["type"] = envitem.GetType()
		item["client_id"] = envitem.GetClientId()

		return item
	}

	return nil
}

func getENVCoreAttributes() []string {
	attributes := [...]string{
		"name", "organization_id", "is_production", "type", "client_id",
	}
	return attributes[:]
}
