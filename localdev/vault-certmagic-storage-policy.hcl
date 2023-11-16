path "secrets/*" {
  policy = "deny"
}

path "secrets/data/certificates/*" {
  capabilities = ["create", "read", "update", "delete", "list", "patch"]
}

path "secrets/metadata/certificates/*" {
  capabilities = ["create", "read", "update", "delete", "list", "patch"]
}

path "auth/token/renew" {
  capabilities = ["read", "update"]
}

path "auth/token/renew-self" {
  capabilities = ["read", "update"]
}

path "auth/token/revoke" {
  capabilities = ["create", "read", "update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}
