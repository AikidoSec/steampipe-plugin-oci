---
title: "Steampipe Table: oci_identity_domain_user - Query OCI Identity Domain Users using SQL"
description: "Allows users to query users within OCI Identity Domains, including their enabled/disabled (active) status."
---

# Table: oci_identity_domain_user - Query OCI Identity Domain Users using SQL

Identity domains represent a user population in Oracle Cloud Infrastructure and are managed independently of the classic IAM users exposed by `oci_identity_user`. Resources within an identity domain (users, groups, dynamic resource groups, and identity providers) are managed through the SCIM-based Identity Domains API rather than the classic Identity API, and this includes each user's enabled/disabled state, which the console shows as the "Active" toggle under Identity > Domains > \<domain\> > Users.

## Table Usage Guide

The `oci_identity_domain_user` table provides insights into users across all ACTIVE identity domains in the tenancy, including whether each user's login is enabled or disabled. Note that this is a different underlying API and user population than `oci_identity_user`, which only reflects users in the tenancy's Default domain. Use `oci_identity_domain_user` when you need visibility into users belonging to any non-Default identity domain, or when you specifically need the `active` (enabled/disabled) status shown in the domain's user list in the console.

The table does not expose a `domain_display_name`/`domain_url` column; join against `oci_identity_domain` on `domain_id` if you need the domain's display name or URL alongside a user.

Several SCIM extension attributes - account state (`urn_state_user`), password state (`urn_password_state`), adaptive risk (`urn_adaptive_user`), capabilities (`urn_capabilities`), MFA (`urn_mfa`), federation/delegation (`urn_user`), and OCI freeform tags (`urn_oci_tags`) - are returned as raw JSON rather than flattened into individual columns, so query them with your database's JSON operators (`->`/`->>` in Postgres, `json_extract()` in SQLite). Some of these, notably `urn_password_state` and `urn_adaptive_user`, may be entirely absent (`null`) for a given user - either because the underlying feature (e.g. adaptive risk scoring) isn't enabled for the domain, or because the attribute is only returned when explicitly asked for by name and isn't reliably fetchable through the list-all-users call this table uses. `meta` (resource metadata, e.g. `created`/`lastModified`) and `schemas` (the list of SCIM schema URNs present on the resource) are likewise raw JSON.

## Examples

### Basic info

```sql+postgres
select
  user_name,
  id,
  domain_id,
  active,
  urn_state_user -> 'locked' ->> 'on' as is_locked,
  urn_state_user ->> 'lastSuccessfulLoginDate' as last_successful_login_date,
  meta ->> 'created' as time_created
from
  oci_identity_domain_user;
```

```sql+sqlite
select
  user_name,
  id,
  domain_id,
  active,
  json_extract(urn_state_user, '$.locked.on') as is_locked,
  json_extract(urn_state_user, '$.lastSuccessfulLoginDate') as last_successful_login_date,
  json_extract(meta, '$.created') as time_created
from
  oci_identity_domain_user;
```

### List disabled users

Find users whose console login is disabled in their identity domain.

```sql+postgres
select
  user_name,
  domain_id,
  ocid,
  urn_state_user -> 'locked' ->> 'on' as is_locked
from
  oci_identity_domain_user
where
  not active;
```

```sql+sqlite
select
  user_name,
  domain_id,
  ocid,
  json_extract(urn_state_user, '$.locked.on') as is_locked
from
  oci_identity_domain_user
where
  active = 0;
```

### List users in a specific identity domain

```sql+postgres
select
  d.display_name as domain_name,
  u.user_name,
  u.active
from
  oci_identity_domain_user as u
  join oci_identity_domain as d on d.id = u.domain_id
where
  d.display_name = 'Default';
```

```sql+sqlite
select
  d.display_name as domain_name,
  u.user_name,
  u.active
from
  oci_identity_domain_user as u
  join oci_identity_domain as d on d.id = u.domain_id
where
  d.display_name = 'Default';
```

### List locked out users

```sql+postgres
select
  user_name,
  domain_id,
  urn_state_user ->> 'loginAttempts' as login_attempts,
  urn_state_user ->> 'lastFailedLoginDate' as last_failed_login_date
from
  oci_identity_domain_user
where
  (urn_state_user -> 'locked' ->> 'on')::boolean;
```

```sql+sqlite
select
  user_name,
  domain_id,
  json_extract(urn_state_user, '$.loginAttempts') as login_attempts,
  json_extract(urn_state_user, '$.lastFailedLoginDate') as last_failed_login_date
from
  oci_identity_domain_user
where
  json_extract(urn_state_user, '$.locked.on') = 1;
```

### List users with a console password but no MFA enrolled

A common CSPM finding: a user who can log in to the console with a password but has not enrolled in Multi-Factor Authentication.

```sql+postgres
select
  user_name,
  domain_id,
  urn_capabilities ->> 'canUseConsolePassword' as can_use_console_password,
  urn_mfa ->> 'mfaStatus' as mfa_status
from
  oci_identity_domain_user
where
  (urn_capabilities ->> 'canUseConsolePassword')::boolean
  and (urn_mfa ->> 'mfaStatus' is distinct from 'ENROLLED');
```

```sql+sqlite
select
  user_name,
  domain_id,
  json_extract(urn_capabilities, '$.canUseConsolePassword') as can_use_console_password,
  json_extract(urn_mfa, '$.mfaStatus') as mfa_status
from
  oci_identity_domain_user
where
  json_extract(urn_capabilities, '$.canUseConsolePassword') = 1
  and (json_extract(urn_mfa, '$.mfaStatus') is null or json_extract(urn_mfa, '$.mfaStatus') != 'ENROLLED');
```

### List users with a password that never expires

```sql+postgres
select
  user_name,
  domain_id,
  urn_password_state ->> 'cantExpire' as password_cant_expire,
  urn_password_state ->> 'lastSuccessfulSetDate' as password_last_successful_set_date
from
  oci_identity_domain_user
where
  (urn_password_state ->> 'cantExpire')::boolean;
```

```sql+sqlite
select
  user_name,
  domain_id,
  json_extract(urn_password_state, '$.cantExpire') as password_cant_expire,
  json_extract(urn_password_state, '$.lastSuccessfulSetDate') as password_last_successful_set_date
from
  oci_identity_domain_user
where
  json_extract(urn_password_state, '$.cantExpire') = 1;
```

### List users flagged as high risk by adaptive authentication

```sql+postgres
select
  user_name,
  domain_id,
  urn_adaptive_user ->> 'riskLevel' as risk_level,
  urn_state_user ->> 'lastSuccessfulLoginDate' as last_successful_login_date
from
  oci_identity_domain_user
where
  urn_adaptive_user ->> 'riskLevel' = 'HIGH';
```

```sql+sqlite
select
  user_name,
  domain_id,
  json_extract(urn_adaptive_user, '$.riskLevel') as risk_level,
  json_extract(urn_state_user, '$.lastSuccessfulLoginDate') as last_successful_login_date
from
  oci_identity_domain_user
where
  json_extract(urn_adaptive_user, '$.riskLevel') = 'HIGH';
```

### List federated users

Federated users authenticate at an external identity provider rather than locally, which changes how findings like "no MFA enrolled" or "has a console password" should be interpreted.

```sql+postgres
select
  user_name,
  domain_id,
  urn_user ->> 'provider' as provider
from
  oci_identity_domain_user
where
  (urn_user ->> 'isFederatedUser')::boolean;
```

```sql+sqlite
select
  user_name,
  domain_id,
  json_extract(urn_user, '$.provider') as provider
from
  oci_identity_domain_user
where
  json_extract(urn_user, '$.isFederatedUser') = 1;
```
