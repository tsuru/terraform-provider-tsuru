// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const roleDefaultResourceID = "default"

const (
	roleEventUserCreate = "user-create"
	roleEventTeamCreate = "team-create"
)

type tsuruRoleWithEvents struct {
	Name   string   `json:"name"`
	Events []string `json:"events"`
}

func resourceTsuruRoleDefault() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages tsuru default roles: roles automatically assigned to a user when it is created, or to a user when it creates a new team. This is a singleton resource: there is a single default roles configuration per tsuru target.",
		CreateContext: resourceTsuruRoleDefaultCreate,
		ReadContext:   resourceTsuruRoleDefaultRead,
		UpdateContext: resourceTsuruRoleDefaultUpdate,
		DeleteContext: resourceTsuruRoleDefaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_create": {
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Role names automatically assigned to a user when it is created",
				AtLeastOneOf: []string{"user_create", "team_create"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					StateFunc:    normalizeRoleValue,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
			},
			"team_create": {
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Role names automatically assigned to a user when it creates a new team",
				AtLeastOneOf: []string{"user_create", "team_create"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					StateFunc:    normalizeRoleValue,
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
			},
		},
	}
}

func resourceTsuruRoleDefaultCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	payload := roleDefaultPayload(d, "user_create", "team_create")
	if err := roleRequest(ctx, provider, http.MethodPost, "1.0", "/role/default", payload, nil); err != nil {
		return diag.Errorf("Could not set tsuru default roles: %s", err)
	}

	d.SetId(roleDefaultResourceID)

	return resourceTsuruRoleDefaultRead(ctx, d, meta)
}

func resourceTsuruRoleDefaultRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	var roles []tsuruRoleWithEvents
	if err := roleRequest(ctx, provider, http.MethodGet, "1.0", "/role/default", nil, &roles); err != nil {
		return diag.Errorf("Could not read tsuru default roles: %s", err)
	}

	var userCreate, teamCreate []string
	for _, role := range roles {
		for _, evt := range role.Events {
			switch evt {
			case roleEventUserCreate:
				userCreate = append(userCreate, role.Name)
			case roleEventTeamCreate:
				teamCreate = append(teamCreate, role.Name)
			}
		}
	}

	if err := d.Set("user_create", userCreate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("team_create", teamCreate); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(roleDefaultResourceID)

	return nil
}

func resourceTsuruRoleDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	oldUserCreate, newUserCreate := permissionChangeField(d, "user_create")
	oldTeamCreate, newTeamCreate := permissionChangeField(d, "team_create")

	removePayload := map[string][]string{}
	if removed := permissionDifference(oldUserCreate, newUserCreate); len(removed) > 0 {
		removePayload[roleEventUserCreate] = removed
	}
	if removed := permissionDifference(oldTeamCreate, newTeamCreate); len(removed) > 0 {
		removePayload[roleEventTeamCreate] = removed
	}
	if len(removePayload) > 0 {
		if err := roleRequest(ctx, provider, http.MethodDelete, "1.0", "/role/default", removePayload, nil); err != nil {
			return diag.Errorf("Could not remove tsuru default roles: %s", err)
		}
	}

	addPayload := map[string][]string{}
	if added := permissionDifference(newUserCreate, oldUserCreate); len(added) > 0 {
		addPayload[roleEventUserCreate] = added
	}
	if added := permissionDifference(newTeamCreate, oldTeamCreate); len(added) > 0 {
		addPayload[roleEventTeamCreate] = added
	}
	if len(addPayload) > 0 {
		if err := roleRequest(ctx, provider, http.MethodPost, "1.0", "/role/default", addPayload, nil); err != nil {
			return diag.Errorf("Could not add tsuru default roles: %s", err)
		}
	}

	return resourceTsuruRoleDefaultRead(ctx, d, meta)
}

func resourceTsuruRoleDefaultDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	payload := roleDefaultPayload(d, "user_create", "team_create")
	if len(payload[roleEventUserCreate]) == 0 && len(payload[roleEventTeamCreate]) == 0 {
		return nil
	}

	if err := roleRequest(ctx, provider, http.MethodDelete, "1.0", "/role/default", payload, nil); err != nil {
		return diag.Errorf("Could not remove tsuru default roles: %s", err)
	}

	return nil
}

func roleDefaultPayload(d *schema.ResourceData, userCreateField, teamCreateField string) map[string][]string {
	return map[string][]string{
		roleEventUserCreate: permissionSet(d.Get(userCreateField)),
		roleEventTeamCreate: permissionSet(d.Get(teamCreateField)),
	}
}

func permissionChangeField(d *schema.ResourceData, field string) ([]string, []string) {
	oldValue, newValue := d.GetChange(field)
	return permissionSet(oldValue), permissionSet(newValue)
}
