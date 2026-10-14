resource "tsuru_role" "team_member" {
  name        = "team-member"
  context     = "global"
  description = "Default role assigned to new users"

  permissions = [
    "app.read",
  ]
}

resource "tsuru_role" "team_owner" {
  name        = "team-owner"
  context     = "team"
  description = "Default role assigned to the user that creates a team"

  permissions = [
    "app",
    "team.update",
  ]
}

# tsuru_role_default is a singleton resource: there is a single default
# roles configuration per tsuru target.
resource "tsuru_role_default" "this" {
  user_create = [tsuru_role.team_member.name]
  team_create = [tsuru_role.team_owner.name]
}
