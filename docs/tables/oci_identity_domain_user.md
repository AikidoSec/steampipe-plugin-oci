---
title: "Steampipe Table: oci_identity_domain_user - Query OCI Identity Domain Users using SQL"
description: "Allows users to query users within OCI Identity Domains, including their enabled/disabled (active) status."
---

# Table: oci_identity_domain_user - Query OCI Identity Domain Users using SQL

Identity domains represent a user population in Oracle Cloud Infrastructure and are managed independently of the classic IAM users exposed by `oci_identity_user`. Resources within an identity domain (users, groups, dynamic resource groups, and identity providers) are managed through the SCIM-based Identity Domains API rather than the classic Identity API, and this includes each user's enabled/disabled state, which the console shows as the "Active" toggle under Identity > Domains > \<domain\> > Users.

## Table Usage Guide

The `oci_identity_domain_user` table provides insights into users across all ACTIVE identity domains in the tenancy, including whether each user's login is enabled or disabled. Note that this is a different underlying API and user population than `oci_identity_user`, which only reflects users in the tenancy's Default domain. Use `oci_identity_domain_user` when you need visibility into users belonging to any non-Default identity domain, or when you specifically need the `active` (enabled/disabled) status shown in the domain's user list in the console.

## Examples

### Basic info

```sql+postgres
select
  user_name,
  id,
  domain_display_name,
  active,
  is_locked,
  last_successful_login_date
from
  oci_identity_domain_user;
```

```sql+sqlite
select
  user_name,
  id,
  domain_display_name,
  active,
  is_locked,
  last_successful_login_date
from
  oci_identity_domain_user;
```

### List disabled users

Find users whose console login is disabled in their identity domain.

```sql+postgres
select
  user_name,
  domain_display_name,
  ocid,
  is_locked
from
  oci_identity_domain_user
where
  not active;
```

```sql+sqlite
select
  user_name,
  domain_display_name,
  ocid,
  is_locked
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
  domain_display_name,
  login_attempts,
  last_failed_login_date
from
  oci_identity_domain_user
where
  is_locked;
```

```sql+sqlite
select
  user_name,
  domain_display_name,
  login_attempts,
  last_failed_login_date
from
  oci_identity_domain_user
where
  is_locked = 1;
```
