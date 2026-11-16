# JWT key storage example

Runnable, standalone Go example using HS512, matching the application's signing
algorithm. It does **not** change the application's existing authentication.
Requires Go 1.22+.

## What belongs in configuration?

Store a **reference**, not the signing key:

```json
{
  "key_file": "C:\\secure\\admin-console\\signing.key",
  "issuer": "jwt-key-store-example",
  "audience": "admin-console"
}
```

`JWT_KEY_FILE` overrides `key_file` and contains only an absolute file path, never
the secret itself. The application reads exactly 64 raw bytes once at startup.
The `init` command generates these bytes with `crypto/rand`, never displays them,
and refuses to overwrite an existing file. Missing keys are errors, not an
invitation to silently generate a new key and invalidate everyone's tokens.

## Run on Windows (PowerShell)

From the repository root:

```powershell
Set-Location examples\jwt-key-store

# Use a NEW directory outside the repository. Restrict its ACL BEFORE generating
# the key; Go's 0600 mode alone does NOT secure files on Windows.
$secretDir = Join-Path $env:LOCALAPPDATA ("admin-console-jwt-" + [guid]::NewGuid().ToString("N"))
$null = New-Item -ItemType Directory -Path $secretDir -ErrorAction Stop
$sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
icacls.exe $secretDir /inheritance:r /grant:r "*${sid}:(OI)(CI)F"
if ($LASTEXITCODE -ne 0) { throw "Cannot restrict secret directory ACL" }
icacls.exe $secretDir

$env:JWT_KEY_FILE = Join-Path $secretDir "signing.key"
go run . init -key "$env:JWT_KEY_FILE"
if ($LASTEXITCODE -ne 0) { throw "Key generation failed" }
go run . demo -config config.example.json
go test .
```

Expected demo output:

```text
Signed and verified an HS512 JWT for "demo-user"; expires at ... . Token not displayed.
```

Keep the path for subsequent sessions; set `JWT_KEY_FILE` to that same file, or
copy `config.example.json` to `config.json` and set `key_file` there. Do not rerun
`init` on every start. JSON Windows paths need escaped backslashes as shown above.

The commands use the current user's identity for a local demonstration. For a
Windows service, provision the directory for its dedicated service identity
instead. Review its DACL: no broad groups or other untrusted identities should
have read/write access. The Go loader does **not** validate Windows DACLs.
Administrators and a compromised service account remain outside this protection.

## Run on Linux/macOS (Bash)

From this example's directory:

```bash
# Private directory outside the repository, not a shared /tmp key.
umask 077
secret_dir="$(mktemp -d "$HOME/admin-console-jwt.XXXXXXXX")"
export JWT_KEY_FILE="$secret_dir/signing.key"
go run . init -key "$JWT_KEY_FILE"
go run . demo -config config.example.json
go test .
```

Use an owner-only directory (`0700`) and key file (`0600`). The loader rejects
group/other file permissions, symlinks, non-regular files and incorrect lengths.
It does not audit parent directory ownership/permissions: provision these
correctly and protect configuration against modification too.

## Why this is safer

- No hardcoded key, password-derived key, checked-in secret, raw secret in an
  environment variable, command-line argument, or log.
- A cryptographically random 512-bit HMAC key; length validation alone cannot
  prove randomness, so use `init`, not a hand-written 64-character password.
- Explicit HS512 allowlist, required expiration, issuer/audience checks,
  issued-at/not-before validation and five-minute tokens.
- No unauthenticated token-issuing HTTP endpoint. `demo` signs a local test
  identity, verifies it, and prints neither token nor key.
- Separate module: run commands here, not from the parent examples module.

The ignore rules are a safety net, not secret protection. Do not put keys in
source control, container images, unencrypted backups, or shared folders.
Keys here are plaintext at rest: file ACLs are the security boundary, not
encryption. Use disk/backup encryption where appropriate. Anyone able to read an
HMAC verification key can also mint tokens.

## Production and rotation

Prefer your platform's secret manager (for example Azure Key Vault, AWS Secrets
Manager, or Vault), authenticated using workload identity instead of a static
cloud credential in configuration. A trusted agent/orchestrator can mount its
value as a restricted file; then configure only the mounted file path. This
example requires a regular file, so symlink-based secret mounts must use a
compatible projection or a separately implemented, trusted provider loader.

If the private signing key must never leave the key service, use asymmetric
signing through KMS/HSM and distribute only public verification keys. Do not
substitute a KMS identifier for an HMAC key or put a decryption password beside an
encrypted key file.

This minimal example loads one key at startup, with no automatic reload or key
ring. Replacing it and restarting immediately invalidates old tokens. For
overlapping rotation, implement an allowlisted `kid`-to-key map: sign with the new
key, retain the previous verification key until its tokens expire, then remove
it. Never turn an untrusted JWT `kid` into a filesystem path or URL.
