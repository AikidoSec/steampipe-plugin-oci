package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableIdentityDomain(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "oci_identity_domain",
		Description: "OCI Identity Domain",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getDomain,
		},
		List: &plugin.ListConfig{
			Hydrate: listDomains,
			KeyColumns: []*plugin.KeyColumn{
				{
					Name:    "lifecycle_state",
					Require: plugin.Optional,
				},
				{
					Name:    "display_name",
					Require: plugin.Optional,
				},
				{
					Name:    "url",
					Require: plugin.Optional,
				},
				{
					Name:    "home_region_url",
					Require: plugin.Optional,
				},
				{
					Name:    "type",
					Require: plugin.Optional,
				},
				{
					Name:    "license_type",
					Require: plugin.Optional,
				},
				{
					Name:    "is_hidden_on_login",
					Require: plugin.Optional,
				},
			},
		},
		Columns: commonColumnsForAllResource([]*plugin.Column{
			{
				Name:        "display_name",
				Description: "The mutable display name of the identity domain.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "The OCID of the identity domain.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromCamel(),
			},
			{
				Name:        "description",
				Description: "The identity domain description. You can have an empty description.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "time_created",
				Description: "Date and time the identity domain was created, in the format defined by RFC3339.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("TimeCreated.Time"),
			},
			{
				Name:        "lifecycle_state",
				Description: "The domain's current state.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "lifecycle_details",
				Description: "Any additional details about the current state of the identity domain.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_hidden_on_login",
				Description: "Indicates whether the identity domain is hidden on the sign-in screen or not.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "url",
				Description: "Region-agnostic identity domain URL.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "home_region_url",
				Description: "Region-specific identity domain URL.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "home_region",
				Description: "The home region for the identity domain.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The type of the identity domain.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "license_type",
				Description: "The license type of the identity domain.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "replica_regions",
				Description: "The regions where replicas of the identity domain exist.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "defined_tags",
				Description: ColumnDescriptionDefinedTags,
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "freeform_tags",
				Description: ColumnDescriptionFreefromTags,
				Type:        proto.ColumnType_JSON,
			},

			// Standard Steampipe columns
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.From(domainTags),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Id").Transform(transform.EnsureStringArray),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DisplayName"),
			},

			// Standard OCI columns
			{
				Name:        "compartment_id",
				Description: ColumnDescriptionTenantId,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CompartmentId"),
			},
		}),
	}
}

//// LIST FUNCTION

func listDomains(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	equalQuals := d.EqualsQuals

	buildRequest := func(compartmentId *string) identity.ListDomainsRequest {
		request := identity.ListDomainsRequest{
			CompartmentId: compartmentId,
			Limit:         types.Int(1000),
			RequestMetadata: common.RequestMetadata{
				RetryPolicy: getDefaultRetryPolicy(d.Connection),
			},
		}

		// Check for additional filters
		if equalQuals["display_name"] != nil {
			name := d.EqualsQualString("display_name")
			request.DisplayName = types.String(name)
		}
		if equalQuals["url"] != nil {
			url := d.EqualsQualString("url")
			request.Url = types.String(url)
		}
		if equalQuals["home_region_url"] != nil {
			homeRegionUrl := d.EqualsQualString("home_region_url")
			request.HomeRegionUrl = types.String(homeRegionUrl)
		}
		if equalQuals["type"] != nil {
			domainType := d.EqualsQualString("type")
			request.Type = types.String(domainType)
		}
		if equalQuals["license_type"] != nil {
			licenseType := d.EqualsQualString("license_type")
			request.LicenseType = types.String(licenseType)
		}
		if equalQuals["is_hidden_on_login"] != nil {
			isLoginHidden := equalQuals["is_hidden_on_login"].GetBoolValue()
			request.IsHiddenOnLogin = types.Bool(isLoginHidden)
		}
		if equalQuals["lifecycle_state"] != nil {
			lifecycleState := d.EqualsQualString("lifecycle_state")
			request.LifecycleState = identity.DomainLifecycleStateEnum(lifecycleState)
		}

		limit := d.QueryContext.Limit
		if limit != nil && *limit < int64(*request.Limit) {
			request.Limit = types.Int(int(*limit))
		}

		return request
	}

	return nil, fetchAllDomains(ctx, d, "listDomains", buildRequest, func(domain identity.DomainSummary) bool {
		d.StreamListItem(ctx, domain)

		// Context can be cancelled due to manual cancellation or the limit has been hit
		return d.RowsRemaining(ctx) != 0
	})
}

// fetchAllDomains fans ListDomains out across every compartment in the tenancy, paginating each
// compartment's results and de-duplicating domains (by OCID) that are visible from more than one
// compartment. This mechanic (compartment fan-out + pagination + dedup) previously had to be
// fixed for duplicate results (see PR #674); it is centralized here so that fix only has to live
// in one place. buildRequest constructs the per-compartment request (server-side qual filters,
// paging limit, etc.); onDomain is invoked for each de-duplicated domain and should return false
// to stop iterating early (e.g. once a row limit has been reached).
func fetchAllDomains(ctx context.Context, d *plugin.QueryData, callerName string, buildRequest func(compartmentId *string) identity.ListDomainsRequest, onDomain func(identity.DomainSummary) bool) error {
	session, err := identityService(ctx, d)
	if err != nil {
		return err
	}

	compartments, err := listAllCompartments(ctx, d)
	if err != nil {
		return err
	}

	// Track domains we've already seen to avoid duplicates
	seenDomains := make(map[string]bool)

	for _, compartment := range compartments {
		if compartment.Id == nil {
			continue
		}

		request := buildRequest(compartment.Id)

		pagesLeft := true
		for pagesLeft {
			response, err := session.IdentityClient.ListDomains(ctx, request)
			if err != nil {
				// Log error but continue with other compartments
				plugin.Logger(ctx).Error(callerName, "ListDomainsError", err, "CompartmentId", *compartment.Id)
				break
			}

			for _, domain := range response.Items {
				// Skip if we've already seen this domain (by ID)
				if domain.Id != nil {
					if seenDomains[*domain.Id] {
						continue
					}
					seenDomains[*domain.Id] = true
				}

				if !onDomain(domain) {
					return nil
				}
			}
			if response.OpcNextPage != nil {
				request.Page = response.OpcNextPage
			} else {
				pagesLeft = false
			}
		}
	}

	return nil
}

// listAllIdentityDomains returns every identity domain in the tenancy, across all compartments
// and regardless of lifecycle state. It is used by tables (e.g. oci_identity_domain_user) that
// need to fan out per-domain calls against the Identity Domains (SCIM) API, which is addressed
// by domain URL rather than by region. Callers that only want usable domains should filter on
// LifecycleState (e.g. identity.DomainLifecycleStateActive) themselves.
func listAllIdentityDomains(ctx context.Context, d *plugin.QueryData) ([]identity.DomainSummary, error) {
	serviceCacheKey := "listAllIdentityDomains"
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.([]identity.DomainSummary), nil
	}

	buildRequest := func(compartmentId *string) identity.ListDomainsRequest {
		return identity.ListDomainsRequest{
			CompartmentId: compartmentId,
			Limit:         types.Int(1000),
			RequestMetadata: common.RequestMetadata{
				RetryPolicy: getDefaultRetryPolicy(d.Connection),
			},
		}
	}

	var domains []identity.DomainSummary
	err := fetchAllDomains(ctx, d, "listAllIdentityDomains", buildRequest, func(domain identity.DomainSummary) bool {
		domains = append(domains, domain)
		return true
	})
	if err != nil {
		return nil, err
	}

	// save domains in cache
	d.ConnectionManager.Cache.Set(serviceCacheKey, domains)

	return domains, nil
}

//// HYDRATE FUNCTIONS

func getDomain(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	id := d.EqualsQuals["id"].GetStringValue()

	// Create Session
	session, err := identityService(ctx, d)
	if err != nil {
		return nil, err
	}

	request := identity.GetDomainRequest{
		DomainId: types.String(id),
		RequestMetadata: common.RequestMetadata{
			RetryPolicy: getDefaultRetryPolicy(d.Connection),
		},
	}

	response, err := session.IdentityClient.GetDomain(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Domain, nil
}

//// TRANSFORM FUNCTION

func domainTags(_ context.Context, d *transform.TransformData) (interface{}, error) {
	switch domain := d.HydrateItem.(type) {
	case identity.Domain:
		return extractTags(domain.FreeformTags, domain.DefinedTags), nil
	case identity.DomainSummary:
		return extractTags(domain.FreeformTags, domain.DefinedTags), nil
	}
	return nil, nil
}
