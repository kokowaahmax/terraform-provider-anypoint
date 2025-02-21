package anypoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	rolegroup "github.com/mulesoft-consulting/anypoint-client-go/rolegroup"
)

func resourceRoleGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleGroupCreate,
		ReadContext:   resourceRoleGroupRead,
		UpdateContext: resourceRoleGroupUpdate,
		DeleteContext: resourceRoleGroupDelete,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Creates a ` + "`" + `rolegroup` + "`" + ` component for your ` + "`" + `org` + "`" + `.
		`,
		Schema: map[string]*schema.Schema{
			"role_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"external_names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"editable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRoleGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)

	authctx := getRoleGroupAuthCtx(ctx, &pco)
	body, errDiags := newRolegroupPostBody(d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	res, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsPost(authctx, orgid).RolegroupPostBody(*body).Execute()
	defer httpr.Body.Close()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create rolegroups",
			Detail:   details,
		})
		return diags
	}
	d.SetId(res.GetRoleGroupId())

	resourceRoleGroupRead(ctx, d, m)

	return diags
}

func resourceRoleGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	rolegroupid := d.Id()

	authctx := getRoleGroupAuthCtx(ctx, &pco)

	res, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdGet(authctx, orgid, rolegroupid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get rolegroup",
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()

	//process data
	rolegroup := flattenRoleGroupData(&res)
	//save in data source schema
	if err := setRolegroupAttributesToResourceData(d, rolegroup); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read rolegroup",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceRoleGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	rolegroupid := d.Id()

	authctx := getRoleGroupAuthCtx(ctx, &pco)

	if d.HasChanges(getRolegroupWatchAttributes()...) {
		body, errDiags := newRolegroupPutBody(d)
		if errDiags.HasError() {
			diags = append(diags, errDiags...)
			return diags
		}

		_, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdPut(authctx, orgid, rolegroupid).RolegroupPutBody(*body).Execute()
		if err != nil {
			var details string
			if httpr != nil {
				b, _ := ioutil.ReadAll(httpr.Body)
				details = string(b)
			} else {
				details = err.Error()
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update rolegroup",
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceRoleGroupRead(ctx, d, m)
}

func resourceRoleGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	rolegroupid := d.Id()

	authctx := getRoleGroupAuthCtx(ctx, &pco)

	_, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsRolegroupIdDelete(authctx, orgid, rolegroupid).Execute()
	if err != nil {
		var details string
		if httpr != nil {
			b, _ := ioutil.ReadAll(httpr.Body)
			details = string(b)
		} else {
			details = err.Error()
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get rolegroup",
			Detail:   details,
		})
		return diags
	}
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

/**
 * Generates body object for creating rolegroup
 */
func newRolegroupPostBody(d *schema.ResourceData) (*rolegroup.RolegroupPostBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	external_names := d.Get("external_names")
	description := d.Get("description")

	body := rolegroup.NewRolegroupPostBodyWithDefaults()

	body.SetName(name)
	if external_names != nil {
		list := external_names.([]interface{})
		ext_names := make([]string, 0)
		for _, val := range list {
			if val != nil {
				ext_names = append(ext_names, val.(string))
			}
		}
		body.SetExternalNames(ext_names)
	}
	if description != nil {
		body.SetDescription(description.(string))
	}

	return body, diags
}

/**
 * Generates body object for updating rolegroup
 */
func newRolegroupPutBody(d *schema.ResourceData) (*rolegroup.RolegroupPutBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := rolegroup.NewRolegroupPutBodyWithDefaults()

	name := d.Get("name").(string)
	external_names := d.Get("external_names")
	description := d.Get("description")

	body.SetName(name)
	if external_names != nil {
		list := external_names.([]interface{})
		ext_names := make([]string, 0)
		for _, val := range list {
			if val != nil {
				ext_names = append(ext_names, val.(string))
			}
		}
		body.SetExternalNames(ext_names)
	}
	if description != nil {
		body.SetDescription(description.(string))
	}

	return body, diags
}

/*
* Copies the given rolegroup into the given resource data
* @param d *schema.ResourceData the resource data schema
* @param rolegroup map[string]interface{} the rolegroup
 */
func setRolegroupAttributesToResourceData(d *schema.ResourceData, rolegroup map[string]interface{}) error {
	attributes := getRolegroupAttributes()
	for _, attr := range attributes {
		if val, ok := rolegroup[attr]; ok {
			if err := d.Set(attr, val); err != nil {
				return fmt.Errorf("unable to set assigned rolegroup attribute %s with value: %s \n details:%s", attr, val, err)
			}
		} else {
			log.Printf("The attribute %s not found in rolegroup.\n", attr)
		}
	}
	return nil
}

func getRolegroupWatchAttributes() []string {
	attributes := [...]string{
		"name", "external_names", "description",
	}
	return attributes[:]
}

/*
 * Returns authentication context (includes authorization header)
 */
func getRoleGroupAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, rolegroup.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, rolegroup.ContextServerIndex, pco.server_index)
}
