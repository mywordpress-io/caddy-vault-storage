# caddy-vault-storage

This is a Storage backend for Caddy (CertMagic) which allows storing of TLS certificates managed by Caddy in 
HashiCorp's Vault.

This plugin can be pulled in via Caddy's build system--to review the CertMagic Storage implementation, review the
associated repo here: https://github.com/mywordpress-io/certmagic-vault-storage

## Usage

### Build

Build Caddy using `xcaddy` with the vault storage plugins:
- `xcaddy build --output bin/caddy --with github.com/mywordpress-io/caddy-vault-storage@<tag> --with github.com/mywordpress-io/certmagic-vault-storage@<tag>`

### Config

Once built, use the following config block to communicate with Vault:

```
vault <address> {
    token <value>

    approle_login_path <value>
    approle_logout_path <value>
    approle_role_id <value>
    approle_secret_id <value>

    secrets_path <value>
    path_prefix <value>

    insecure_skip_verify <value>

    lock_timeout <value>
    lock_polling_interval <value>
}
```

For more information, review `Caddyfile.example` and `Caddyfile.json`.

Either 'address' + 'token' -OR- 'address' + 'approle_role_id'+'approle_secret_id' settings are required:
- If using 'approle' authentication, short-lived tokens are managed on the fly.
- If using 'token' authentication, management of the token (renewal, revocation, etc.) is up to the caller.

| Name                    | Type       | Required?     | Description                                                                                 | Default                |
|-------------------------|------------|---------------|---------------------------------------------------------------------------------------------|------------------------|
| `address`               | `url`      | yes           | Vault address URL                                                                           | -                      |
| `token`                 | `string`   | conditionally | Vault static Token to authenticate (this or approle_role_id+approle_secret_id are required) | -                      |
| `approle_login_path`    | `string`   | no            | Login path for approle authentication                                                       | auth/approle/login     |
| `approle_logout_path`   | `string`   | no            | Logout path for approle authentication                                                      | auth/token/revoke-self |
| `approle_role_id`       | `string`   | conditionally | Approle RoleID value for authentication (required if 'token' empty)                         | -                      |
| `approle_secret_id`     | `string`   | conditionally | Approle SecretID value for authentication (required if 'token' empty)                       | -                      |
| `secrets_path`          | `string`   | yes           | Base path to secrets (KV-V2) mount in Vault                                                 | -                      |
| `path_prefix`           | `string`   | no            | Prefix path in the KV-V2 mount in Vault                                                     | -                      |
| `insecure_skip_verify`  | `bool`     | no            | Disable verification of TLS certificate when communicating with Vault                       | false                  |
| `lock_timeout`          | `duration` | no            | Storage lock timeout duration                                                               | 5m                     |
| `lock_polling_interval` | `duration` | no            | Storage lock polling interval                                                               | 5s                     |

## Additional Help

Report any problems or questions with the plugin using a GitHub issue.
