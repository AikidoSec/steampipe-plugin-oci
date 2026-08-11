package oci

import (
	"context"
	"fmt"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/identitydomains"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

// identityDomainUser wraps a SCIM user with the identity domain it belongs to, since the
// Identity Domains API is scoped to a single domain per request/endpoint and does not
// otherwise identify which domain a returned user came from.
type identityDomainUser struct {
	identitydomains.User
	DomainId          *string
	DomainDisplayName *string
	DomainUrl         *string
}

func tableIdentityDomainUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_identity_domain_user",
		Description: "OCI Identity Domain User",
		List: &plugin.ListConfig{
			Hydrate: listIdentityDomainUsers,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "domain_id",
					Require: plugin.Optional,
				},
				{
					Name:    "user_name",
					Require: plugin.Optional,
				},
				{
					Name:    "active",
					Require: plugin.Optional,
				},
			},
		},
		Columns: commonColumnsForAllResource([]*plugin.Column{
			{
				Name:        "user_name",
				Description: "The user's login name.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
			},
			{
				Name:        "id",
				Description: "The identifier of the user within the identity domain (SCIM id). Required, together with domain_id, to address the user through the Identity Domains API.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Id"),
			},
			{
				Name:        "ocid",
				Description: "The OCID of the user, if one has been assigned.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Ocid"),
			},
			{
				Name:        "domain_id",
				Description: "The OCID of the identity domain this user belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DomainId"),
			},
			{
				Name:        "domain_display_name",
				Description: "The display name of the identity domain this user belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DomainDisplayName"),
			},
			{
				Name:        "domain_url",
				Description: "The region-agnostic URL of the identity domain this user belongs to.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DomainUrl"),
			},
			{
				Name:        "active",
				Description: "Indicates whether the user's login is enabled (true) or disabled (false) in this identity domain. This is the value shown as the Enabled/Disabled toggle under Identity > Domains > Users in the console, and is only available through the Identity Domains API (unlike oci_identity_user, which only reflects the Default domain).",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("Active"),
			},
			{
				Name:        "display_name",
				Description: "The name of the user, suitable for display to end users.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DisplayName"),
			},
			{
				Name:        "name",
				Description: "The components of the user's name (given name, family name, formatted name, etc).",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "emails",
				Description: "The email addresses assigned to the user.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Emails"),
			},
			{
				Name:        "user_type",
				Description: "Used to identify the relationship between the user and the identity domain, for example 'internal' or 'external'.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserType"),
			},
			{
				Name:        "external_id",
				Description: "Identifier of the user in the identity provider, if the user was federated into the domain.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ExternalId"),
			},
			{
				Name:        "is_locked",
				Description: "Indicates whether the user's account is locked, for example after too many failed login attempts.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("UrnIetfParamsScimSchemasOracleIdcsExtensionUserStateUser.Locked.On"),
			},
			{
				Name:        "last_successful_login_date",
				Description: "Date and time the user last successfully logged in.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UrnIetfParamsScimSchemasOracleIdcsExtensionUserStateUser.LastSuccessfulLoginDate"),
			},
			{
				Name:        "last_failed_login_date",
				Description: "Date and time of the user's last failed login attempt.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UrnIetfParamsScimSchemasOracleIdcsExtensionUserStateUser.LastFailedLoginDate"),
			},
			{
				Name:        "login_attempts",
				Description: "The number of consecutive failed login attempts recorded for the user.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("UrnIetfParamsScimSchemasOracleIdcsExtensionUserStateUser.LoginAttempts"),
			},
			{
				Name:        "time_created",
				Description: "Date and time the user was created.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Meta.Created"),
			},
			{
				Name:        "time_last_modified",
				Description: "Date and time the user was last modified.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Meta.LastModified"),
			},
			{
				Name:        "groups",
				Description: "List of groups the user belongs to, either through direct membership, nested groups, or dynamically calculated.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Groups"),
			},
			{
				Name:        "urn_capabilities",
				Description: "The user's capabilities within the identity domain, for example whether they can use API keys, auth tokens, or console passwords.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("UrnIetfParamsScimSchemasOracleIdcsExtensionCapabilitiesUser"),
			},

			// Standard Steampipe columns
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(identityDomainUserAkas),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
			},
		}),
	}
}

//// TRANSFORM FUNCTION

// identityDomainUserAkas prefers the globally-unique OCID (consistent with every other table's
// akas column) and only falls back to the domain-scoped SCIM id for the rare user that hasn't
// been assigned an OCID yet.
func identityDomainUserAkas(_ context.Context, d *transform.TransformData) (interface{}, error) {
	user := d.HydrateItem.(identityDomainUser)

	if user.Ocid != nil {
		return []string{*user.Ocid}, nil
	}
	if user.Id != nil {
		return []string{*user.Id}, nil
	}
	return nil, nil
}

// scimFilterQuote escapes a value for embedding in a double-quoted SCIM filter string, per the
// SCIM filter grammar (RFC 7644 section 3.4.2.2 / draft-ietf-scim-api section 3.4.2.2).
func scimFilterQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

//// LIST FUNCTION

func listIdentityDomainUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	equalQuals := d.EqualsQuals

	domains, err := listAllIdentityDomains(ctx, d)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domain.Id == nil || domain.Url == nil {
			continue
		}

		// The Identity Domains API is scoped to a single domain per request; skip domains
		// that aren't ACTIVE, since deleted/deleting domains no longer serve requests.
		if domain.LifecycleState != identity.DomainLifecycleStateActive {
			continue
		}

		if equalQuals["domain_id"] != nil && equalQuals["domain_id"].GetStringValue() != *domain.Id {
			continue
		}

		session, err := identityDomainsService(ctx, d, *domain.Url)
		if err != nil {
			plugin.Logger(ctx).Error("listIdentityDomainUsers", "identityDomainsService.Error", err, "DomainId", *domain.Id)
			continue
		}

		request := identitydomains.ListUsersRequest{
			// Request the full SCIM attribute set (e.g. groups, extension attributes) up
			// front, to avoid an extra per-user hydrate call.
			AttributeSets: []identitydomains.AttributeSetsEnum{identitydomains.AttributeSetsAll},
			Count:         types.Int(1000),
			RequestMetadata: common.RequestMetadata{
				RetryPolicy: getDefaultRetryPolicy(d.Connection),
			},
		}

		var filters []string
		if equalQuals["user_name"] != nil {
			filters = append(filters, fmt.Sprintf(`userName eq "%s"`, scimFilterQuote(equalQuals["user_name"].GetStringValue())))
		}
		if equalQuals["active"] != nil {
			filters = append(filters, fmt.Sprintf("active eq %t", equalQuals["active"].GetBoolValue()))
		}
		if len(filters) > 0 {
			request.Filter = types.String(strings.Join(filters, " and "))
		}

		startIndex := 1
		for {
			request.StartIndex = types.Int(startIndex)

			response, err := session.IdentityDomainsClient.ListUsers(ctx, request)
			if err != nil {
				plugin.Logger(ctx).Error("listIdentityDomainUsers", "ListUsersError", err, "DomainId", *domain.Id)
				break
			}

			for _, user := range response.Resources {
				d.StreamListItem(ctx, identityDomainUser{
					User:              user,
					DomainId:          domain.Id,
					DomainDisplayName: domain.DisplayName,
					DomainUrl:         domain.Url,
				})

				// Context can be cancelled due to manual cancellation or the limit has been hit
				if d.RowsRemaining(ctx) == 0 {
					return nil, nil
				}
			}

			if response.ItemsPerPage == nil || *response.ItemsPerPage == 0 || response.TotalResults == nil {
				break
			}
			startIndex += *response.ItemsPerPage
			if startIndex > *response.TotalResults {
				break
			}
		}
	}

	return nil, nil
}
