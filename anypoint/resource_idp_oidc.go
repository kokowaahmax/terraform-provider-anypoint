package anypoint

import (
	"context"
	"io/ioutil"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	idp "github.com/mulesoft-consulting/anypoint-client-go/idp"
)

func resourceOIDC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOIDCCreate,
		ReadContext:   resourceOIDCRead,
		UpdateContext: resourceOIDCUpdate,
		DeleteContext: resourceOIDCDelete,
		Description: `
		Creates an ` + "`" + `identity provider` + "`" + ` OIDC type configuration in your account.
		`,
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The business group id",
			},
			"provider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider id",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the provider",
			},
			"type": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The type of the provider, contains description and the name of the type of the provider (saml or oidc)",
			},
			"oidc_provider": {
				Type:        schema.TypeSet,
				Description: "The description of provider specific for OIDC types",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The token url of the openid-connect provider",
						},
						"redirect_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The redirect url of the openid-connect provider",
						},
						"userinfo_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The userinfo url of the openid-connect provider",
						},
						"authorize_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The authorization url of the openid-connect provider",
						},
						"client_registration_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The registration url, for dynamic client registration, of the openid-connect provider. Mutually exclusive with credentials id/secret, if both are given registration url is prioritized.",
						},
						"client_credentials_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The client's credentials id. This should only be provided if manual registration is wanted. Mutually exclusive with registration url, if both are given registration url is prioritized.",
						},
						"client_credentials_secret": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The client's credentials secret. This should only be provided if manual registration is wanted. Mutually exclusive with registration url, if both are given registration url is prioritized.",
						},
						"client_token_endpoint_auth_methods_supported": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of authentication methods supported",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"issuer": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The provider token issuer url",
						},
						"group_scope": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The provider group scopes",
						},
						"allow_untrusted_certificates": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "The certification validation trigger",
						},
					},
				},
			},
			"sp_sign_on_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider's sign on url",
			},
			"sp_sign_out_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The provider's sign out url, only available for SAML",
			},
		},
	}
}

func resourceOIDCCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)

	authctx := getIDPAuthCtx(ctx, &pco)
	body, errDiags := newOIDCPostBody(d)
	if errDiags.HasError() {
		diags = append(diags, errDiags...)
		return diags
	}

	res, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersPost(authctx, orgid).IdpPostBody(*body).Execute()
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
			Summary:  "Unable to create OIDC provider for org " + orgid,
			Detail:   details,
		})
		return diags
	}

	d.SetId(res.GetProviderId())

	return resourceOIDCRead(ctx, d, m)
}

func resourceOIDCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getIDPAuthCtx(ctx, &pco)

	//request idp
	res, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdGet(authctx, orgid, idpid).Execute()
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
			Summary:  "Unable to Get IDP " + idpid + " in org " + orgid,
			Detail:   details,
		})
		return diags
	}
	//process data
	idpinstance := flattenIDPData(&res)
	//save in data source schema
	if err := setIDPAttributesToResourceData(d, idpinstance); err != nil {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set IDP " + idpid + " in org " + orgid,
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceOIDCUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)

	if d.HasChanges(getIDPAttributes()...) {
		authctx := getIDPAuthCtx(ctx, &pco)
		body, errDiags := newOIDCPatchBody(d)
		if errDiags.HasError() {
			diags = append(diags, errDiags...)
			return diags
		}
		_, httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdPatch(authctx, orgid, idpid).IdpPatchBody(*body).Execute()
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
				Summary:  "Unable to Update IDP " + idpid + " in org " + orgid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceOIDCRead(ctx, d, m)
}

func resourceOIDCDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	idpid := d.Id()
	orgid := d.Get("org_id").(string)
	authctx := getIDPAuthCtx(ctx, &pco)

	httpr, err := pco.idpclient.DefaultApi.OrganizationsOrgIdIdentityProvidersIdpIdDelete(authctx, orgid, idpid).Execute()
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
			Summary:  "Unable to Delete OIDC provider " + idpid,
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

/* Prepares the body required to post an OIDC provider*/
func newOIDCPostBody(d *schema.ResourceData) (*idp.IdpPostBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	oidc_provider_input := d.Get("oidc_provider")

	body := idp.NewIdpPostBody()

	oidc_type := idp.NewIdpPostBodyType()
	oidc_type.SetName("openid")
	oidc_type.SetDescription("OpenID Connect")

	oidc_provider := idp.NewOidcProvider1()

	if oidc_provider_input != nil {
		set := oidc_provider_input.(*schema.Set)
		list := set.List()
		if len(list) > 0 {
			item := list[0]
			data := item.(map[string]interface{})
			// reads client registration or credentials depending on which one is added
			client := idp.NewClient1()
			client_urls := idp.NewUrls1()
			if client_registration_url, ok := data["client_registration_url"]; ok {
				client_urls.SetRegister(client_registration_url.(string))
				client.SetUrls(*client_urls)
			} else {
				credentials := idp.NewCredentials1()
				if client_credentials_id, ok := data["client_credentials_id"]; ok {
					credentials.SetId(client_credentials_id.(string))
				}
				if client_credentials_secret, ok := data["client_credentials_secret"]; ok {
					credentials.SetSecret(client_credentials_secret.(string))
				}
				client.SetCredentials(*credentials)
			}
			oidc_provider.SetClient(*client)

			//Parsing URLs
			urls := idp.NewUrls3()
			if token_url, ok := data["token_url"]; ok {
				urls.SetToken(token_url.(string))
			}
			if userinfo_url, ok := data["userinfo_url"]; ok {
				urls.SetUserinfo(userinfo_url.(string))
			}
			if authorize_url, ok := data["authorize_url"]; ok {
				urls.SetAuthorize(authorize_url.(string))
			}
			oidc_provider.SetUrls(*urls)

			if issuer, ok := data["issuer"]; ok {
				oidc_provider.SetIssuer(issuer.(string))
			}
			if group_scope, ok := data["group_scope"]; ok {
				oidc_provider.SetGroupScope(group_scope.(string))
			}
			if allow_untrusted_certificates, ok := data["allow_untrusted_certificates"]; ok {
				body.SetAllowUntrustedCertificates(allow_untrusted_certificates.(bool))
			}
		}
	}

	body.SetType(*oidc_type)
	body.SetName(name)
	body.SetOidcProvider(*oidc_provider)

	return body, diags
}

/* Prepares the body required to patch an OIDC provider*/
func newOIDCPatchBody(d *schema.ResourceData) (*idp.IdpPatchBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	oidc_provider_input := d.Get("oidc_provider")

	body := idp.NewIdpPatchBody()

	oidc_type := idp.NewIdpPatchBodyType()
	oidc_type.SetDescription("OpenID Connect")

	oidc_provider := idp.NewOidcProvider1()

	if oidc_provider_input != nil {
		set := oidc_provider_input.(*schema.Set)
		list := set.List()
		if set.Len() > 0 {
			item := list[0]
			data := item.(map[string]interface{})
			// reads client registration or credentials depending on which one is added
			client := idp.NewClient1()
			client_urls := idp.NewUrls1()
			if client_registration_url, ok := data["client_registration_url"]; ok {
				client_urls.SetRegister(client_registration_url.(string))
				client.SetUrls(*client_urls)
			} else {
				credentials := idp.NewCredentials1()
				if client_credentials_id, ok := data["client_credentials_id"]; ok {
					credentials.SetId(client_credentials_id.(string))
				}
				if client_credentials_secret, ok := data["client_credentials_secret"]; ok {
					credentials.SetSecret(client_credentials_secret.(string))
				}
				client.SetCredentials(*credentials)
			}
			oidc_provider.SetClient(*client)

			//Parsing URLs
			urls := idp.NewUrls3()
			if token_url, ok := data["token_url"]; ok {
				urls.SetToken(token_url.(string))
			}
			if userinfo_url, ok := data["userinfo_url"]; ok {
				urls.SetUserinfo(userinfo_url.(string))
			}
			if authorize_url, ok := data["authorize_url"]; ok {
				urls.SetAuthorize(authorize_url.(string))
			}
			oidc_provider.SetUrls(*urls)

			if issuer, ok := data["issuer"]; ok {
				oidc_provider.SetIssuer(issuer.(string))
			}
			if group_scope, ok := data["group_scope"]; ok {
				oidc_provider.SetGroupScope(group_scope.(string))
			}
			if allow_untrusted_certificates, ok := data["allow_untrusted_certificates"]; ok {
				body.SetAllowUntrustedCertificates(allow_untrusted_certificates.(bool))
			}
		}
	}

	body.SetType(*oidc_type)
	body.SetName(name)
	body.SetOidcProvider(*oidc_provider)

	return body, diags
}

func getIDPAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, idp.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, idp.ContextServerIndex, pco.server_index)
}
