package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	amq "../local-mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAmq() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleGroupRolesCreate,
		ReadContext:   resourceRoleGroupRolesRead,
		// UpdateContext: resourceRoleGroupRolesUpdate,
		DeleteContext: resourceRoleGroupRolesDelete,
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

func resourceAmqCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("reg_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getAqmAuthCtx(ctx, &pco)
	// pre-fix below
	// body, errDiags := newAmqPostBody(orgid, envid, d)
	// if errDiags.HasError() {
	// 	diags = append(diags, errDiags...)
	// 	return diags
	// }

	body := newAmqPostBody(d)

	//request aqm creation
	_, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdPut(authctx, orgid, envid, regionid, queueid).QueueOptional(*body).Execute()
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
			Summary:  "Unable to make queue",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	d.SetId(orgid + "_" + envid)

	resourceAmqRead(ctx, d, m)

	return diags
}

func resourceAmqRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("regionId").(string)

	authctx := getAqmAuthCtx(ctx, &pco)

	res, httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsGet(authctx, orgid, envid, regionid).Execute()
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
			Summary:  "Unable to get queue",
			Detail:   details,
		})
		return diags
	}

	if res == nil {
		fmt.Fprintln(os.Stderr, "Result is null.")
	}

	return diags
}

func resourceAmqDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	envid := d.Get("env_id").(string)
	orgid := d.Get("org_id").(string)
	regionid := d.Get("region_id").(string)
	queueid := d.Get("queue_id").(string)

	authctx := getAqmAuthCtx(ctx, &pco)

	httpr, err := pco.amqclient.DefaultApi.V1OrganizationsOrgIdEnvironmentsEnvIdRegionsRegionIdDestinationsQueuesQueueIdDelete(authctx, orgid, envid, regionid, queueid).Execute()
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
			Summary:  "Unable to Delete queue",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

/**
 * Generates body object for creating rolegroup roles - PRE FIX
 */
// func newAmqPostBody(org_id string, env_id string, d *schema.ResourceData) ([]map[string]interface{}, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	queues := d.Get("queues").([]interface{})

// 	if len(queues) == 0 {
// 		return nil, diags
// 	}
// 	res := make([]map[string]interface{}, len(queues))
// 	for i, queue := range queues {
// 		item := make(map[string]interface{})
// 		item["role_id"] = queue.(map[string]interface{})["role_id"].(string)
// 		item["context_params"] = map[string]string{
// 			"org": org_id,
// 		}
// 		res[i] = item
// 	}
// 	return res, diags
// }

func newAmqPostBody(d *schema.ResourceData) *amq.QueueOptional {
	body := amq.NewQueueOptionalWithDefaults()
	body.SetDefaultTtl(d.Get("defaultTtl").(int32))
	body.SetDefaultLockTtl(d.Get("defaultLockTtl").(int32))
	body.SetType(d.Get("type").(string))
	body.SetEncrypted(d.Get("encrypted").(bool))
	return body
}

func flattenAmqData(queues []amq.QueueOptional) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(queues) > 0 {
		res := make([]map[string]interface{}, len(queues))

		for i, items := range queues {
			item := make(map[string]interface{})

			item["defaultTtl"] = items.GetDefaultTtl()
			item["defaultLockTtl"] = items.GetDefaultLockTtl()
			item["type"] = items.GetType()
			item["encrypted"] = items.GetEncrypted()

			res[i] = item
		}
		return res, diags
	}

	return make([]map[string]interface{}, 0), diags
}

/*
* Copies the given rolegroup assigned roles into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param assigned_roles map[string]interface{} the rolegroup assigned roles
 */
func setAmqToResourceData(d *schema.ResourceData, mqueues []map[string]interface{}) error {
	attributes := getAqmAttributes()

	if len(mqueues) == 0 {
		return nil
	}
	queues := make([]map[string]interface{}, len(mqueues))
	for i, queue := range mqueues {
		queue2 := make(map[string]interface{})
		for _, attr := range attributes {
			queue2[attr] = queue[attr]
		}
		queues[i] = queue2
	}
	if err := d.Set("queues", queues); err != nil {
		return fmt.Errorf("unable to set queue attribute \n details: %s", err)
	}
	return nil
}

/**
 * Returns Assigned Roles attributes (core attributes)
 */
func getAqmAttributes() []string {
	attributes := [...]string{
		"defaultTtl", "defaultLockTtl", "type", "encrypted",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getAqmAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	return context.WithValue(ctx, amq.ContextAccessToken, pco.access_token)
}
