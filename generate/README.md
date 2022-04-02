Backup Repository - Backup & Restore commands generator
-------------------------------------------------------

Purpose of this generator is to create procedures from templates for both **Backup** and **Restore** operations to use in automated way.
Backup made using generated procedure should be possible to restore with a restore procedure in automated way.

**The generator is having two output formats:**
- shell script
- Kubernetes-like `kind: Job` and `kind: Pod`


Usage concept
-------------

1. User prepares YAML file as input definition **for both backup & restore**

```yaml
Params:
    hostname: postgres.db.svc.cluster.local
    port: 5432
    db: rkc-test
    user: riotkit
    password: "${DB_PASSWORD}" # injects a shell-syntax, put your password in a `kind: Secret` and mount as environment variable. You can also use $(cat /mnt/secret) syntax, be aware of newlines!

Repository:
    url: "https://example.org"
    token: "${BR_TOKEN}"
    encryptionKeyPath: "/var/lib/backup-repository/encryption.key"
    passphrase: "${GPG_PASSPHRASE}"
    recipient: "your-gpg@email.org"
    collectionId: "111-222-333-444"

```

2. Next, User runs a generation command e.g. `rkc backups generate backup --kubernetes --cron '*/30 * * *`

In result the User gets prepared Kubernetes manifests that could be applied to the cluster.
