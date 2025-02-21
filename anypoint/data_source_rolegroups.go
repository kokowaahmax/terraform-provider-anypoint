package anypoint

import (
	"context"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mulesoft-consulting/anypoint-client-go/rolegroup"
)

func dataSourceRoleGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleGroupsRead,
		DeprecationMessage: `
		This resource is deprecated, please use ` + "`" + `teams` + "`" + `, ` + "`" + `team_members` + "`" + `team_roles` + "`" + ` instead.
		`,
		Description: `
		Reads all ` + "`" + `rolegroups` + "`" + ` available in your business group.
		`,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"org_id": {
							Type:     schema.TypeString,
							Computed: true,
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
						"user_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceRoleGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	authctx := getRoleGroupAuthCtx(ctx, &pco)
	res, httpr, err := pco.rolegroupclient.DefaultApi.OrganizationsOrgIdRolegroupsGet(authctx, orgid).Execute()
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
			Summary:  "Unable to Get rolegroups",
			Detail:   details,
		})
		return diags
	}
	//process data
	data := res.GetData()
	rolegroups := flattenRoleGroupsData(&data)
	//save in data source schema
	if err := d.Set("role_groups", rolegroups); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set rolegroups",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("total", res.GetTotal()); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set total number rolegroups",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

/*
* Transforms a set of rolegroups to the dataSourceRoleGroups schema
* @param rolegroups *[]rolegroup.Rolegroup the list of rolegroups
* @return list of generic items
 */
func flattenRoleGroupsData(rolegroups *[]rolegroup.Rolegroup) []interface{} {
	if rolegroups != nil && len(*rolegroups) > 0 {
		res := make([]interface{}, len(*rolegroups))

		for i, rolegroup := range *rolegroups {
			item := make(map[string]interface{})

			item["role_group_id"] = rolegroup.GetRoleGroupId()
			item["name"] = rolegroup.GetName()
			item["external_names"] = rolegroup.GetExternalNames()
			item["description"] = rolegroup.GetDescription()
			item["org_id"] = rolegroup.GetOrgId()
			item["editable"] = rolegroup.GetEditable()
			item["created_at"] = rolegroup.GetCreatedAt()
			item["updated_at"] = rolegroup.GetUpdatedAt()
			if val, ok := rolegroup.GetUserCountOk(); ok {
				item["user_count"] = val
			}

			res[i] = item
		}
		return res
	}
	return make([]interface{}, 0)
}
