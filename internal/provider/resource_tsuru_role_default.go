// Copyright 2026 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package provider

import (
	"context"
	"fmt"
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
			StateContext: resourceTsuruRoleDefaultImport,
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

func resourceTsuruRoleDefaultImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if d.Id() != roleDefaultResourceID {
		return nil, fmt.Errorf("tsuru_role_default is a singleton resource, the only valid import ID is %q", roleDefaultResourceID)
	}
	return []*schema.ResourceData{d}, nil
}

func resourceTsuruRoleDefaultCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	currentUserCreate, currentTeamCreate, err := readRoleDefaults(ctx, provider)
	if err != nil {
		return diag.Errorf("Could not read tsuru default roles: %s", err)
	}

	newUserCreate := permissionSet(d.Get("user_create"))
	newTeamCreate := permissionSet(d.Get("team_create"))

	if diags := applyRoleDefaultDiff(ctx, provider, currentUserCreate, newUserCreate, currentTeamCreate, newTeamCreate); diags != nil {
		return diags
	}

	d.SetId(roleDefaultResourceID)

	return resourceTsuruRoleDefaultRead(ctx, d, meta)
}

func resourceTsuruRoleDefaultRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	userCreate, teamCreate, err := readRoleDefaults(ctx, provider)
	if err != nil {
		return diag.Errorf("Could not read tsuru default roles: %s", err)
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

func readRoleDefaults(ctx context.Context, provider *tsuruProvider) (userCreate, teamCreate []string, err error) {
	var roles []tsuruRoleWithEvents
	if err := roleRequest(ctx, provider, http.MethodGet, "1.0", "/role/default", nil, &roles); err != nil {
		return nil, nil, err
	}

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

	return userCreate, teamCreate, nil
}

func resourceTsuruRoleDefaultUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	oldUserCreate, newUserCreate := permissionChangeField(d, "user_create")
	oldTeamCreate, newTeamCreate := permissionChangeField(d, "team_create")

	if diags := applyRoleDefaultDiff(ctx, provider, oldUserCreate, newUserCreate, oldTeamCreate, newTeamCreate); diags != nil {
		return diags
	}

	return resourceTsuruRoleDefaultRead(ctx, d, meta)
}

func applyRoleDefaultDiff(ctx context.Context, provider *tsuruProvider, oldUserCreate, newUserCreate, oldTeamCreate, newTeamCreate []string) diag.Diagnostics {
	removeValues := roleDefaultValues{
		UserCreate: permissionDifference(oldUserCreate, newUserCreate),
		TeamCreate: permissionDifference(oldTeamCreate, newTeamCreate),
	}
	if len(removeValues.UserCreate) > 0 || len(removeValues.TeamCreate) > 0 {
		if err := roleRequest(ctx, provider, http.MethodDelete, "1.0", "/role/default", removeValues, nil); err != nil {
			return diag.Errorf("Could not remove tsuru default roles: %s", err)
		}
	}

	addValues := roleDefaultValues{
		UserCreate: permissionDifference(newUserCreate, oldUserCreate),
		TeamCreate: permissionDifference(newTeamCreate, oldTeamCreate),
	}
	if len(addValues.UserCreate) > 0 || len(addValues.TeamCreate) > 0 {
		if err := roleRequest(ctx, provider, http.MethodPost, "1.0", "/role/default", addValues, nil); err != nil {
			return diag.Errorf("Could not add tsuru default roles: %s", err)
		}
	}

	return nil
}

func resourceTsuruRoleDefaultDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	provider := meta.(*tsuruProvider)

	values := roleDefaultPayload(d, "user_create", "team_create")
	if len(values.UserCreate) == 0 && len(values.TeamCreate) == 0 {
		return nil
	}

	if err := roleRequest(ctx, provider, http.MethodDelete, "1.0", "/role/default", values, nil); err != nil {
		return diag.Errorf("Could not remove tsuru default roles: %s", err)
	}

	return nil
}

type roleDefaultValues struct {
	UserCreate []string `json:"user-create,omitempty"`
	TeamCreate []string `json:"team-create,omitempty"`
}

func roleDefaultPayload(d *schema.ResourceData, userCreateField, teamCreateField string) roleDefaultValues {
	return roleDefaultValues{
		UserCreate: permissionSet(d.Get(userCreateField)),
		TeamCreate: permissionSet(d.Get(teamCreateField)),
	}
}

func permissionChangeField(d *schema.ResourceData, field string) ([]string, []string) {
	oldValue, newValue := d.GetChange(field)
	return permissionSet(oldValue), permissionSet(newValue)
}
