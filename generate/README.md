Backup & Restore procedure Generator
------------------------------------

Purpose of this generator is to create procedures from templates for both **Backup** and **Restore** operations to use in automated way.
Backup made using generated procedure should be possible to restore with a restore procedure in automated way.

**The generator is having two output formats:**
- shell script
- Kubernetes-like `kind: Job` and `kind: Pod`

Commands
--------

## *bmg backup*

Generates a bash script **ready for automated usage** with crontab or **Kubernetes** `kind: Cronjob`.


**Example usage with Kubernetes:**

```bash
# for definition.yaml example see yaml file in Usage Concept
bmg backup \
		--definition=./definition.yaml \ 
		--template postgres \
		--kubernetes \
		--gpg-key-path valid-sealed-secret.yaml \
		--output-dir=backup.yaml
```

Will create a `backup.yaml` with Kubernetes resources:
- `kind: Cronjob`
- `kind: Secret`
- `kind: SealedSecret`
- `kind: ConfigMap`

### SealedSecrets

Fully GitOps support in Kubernetes-native way - generated set of Kubernetes resources can be committed to Git repository in a secure way
using SealedSecrets integration.

When `--gpg-key-path` commandline switch points to a valid SealedSecrets yaml file instead of a plaintext GPG key then a SealedSecret will be used.

**Example of valid file:**

```yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
    name: mysecret
    # namespace name must match --k8s-namespace commandline switch value
    # namespace: ... 
spec:
    encryptedData:
        gpg-key: AgBy3i4OJSWK+PiTySYZZA9rO43cGDEq.....

        # there you can place environment variables passed to the container, for example:
        GPG_PASSPHRASE: ....some-encoded-passphrase-with-sealed-secrets-mechanism...
        DB_PASSWORD: ...
        BR_TOKEN: ...
        # all those ENVs will be available to use in section "Params" and "Repository" in definition.yaml configuration file (--definition)
```

### Reference

- Used Kubernetes templates: [chart](./chart)
- Standard available service templates: [service templates](./templates/backup)

Usage concept
-------------

1. User prepares YAML file as input definition **for both backup & restore**

```yaml
# System-specific variables, in this case specific to PostgreSQL
# ${...} and $(...) syntax will be evaluated in target environment e.g. Kubernetes POD
Params:
    hostname: postgres.db.svc.cluster.local
    port: 5432
    db: rkc-test
    user: riotkit
    password: "${DB_PASSWORD}" # injects a shell-syntax, put your password in a `kind: Secret` and mount as environment variable. You can also use $(cat /mnt/secret) syntax, be aware of newlines!

# Generic repository access details. Everything here will land AS IS into the bash script.
# This means that any ${...} and $(...) will be executed in target environment e.g. inside Kubernetes POD
Repository:
    url: "https://example.org"
    token: "${BR_TOKEN}"
    encryptionKeyPath: "/var/lib/backup-repository/encryption.key"
    passphrase: "${GPG_PASSPHRASE}"
    recipient: "your-gpg@email.org"
    collectionId: "111-222-333-444"

# Generic values for Helm used to generate jobs/pods. Those values will overwrite others.
# Notice: Environment variables with '${...}' and '$(...)' will be evaluated in LOCAL SHELL DURING BUILD
HelmValues:
    name: "hello-world"
    env: {}
        # if specified, then will be added to `kind: Secret` and injected into POD as environment
        # the value from ${GPG_PASSPHRASE} will be retrieved from the SHELL DURING THE BUILD
        #GPG_PASSPHRASE: "${GPG_PASSPHRASE}"

        # most secure way for Kubernetes is to not provide secrets there, but define them as environment variables
        # inside SealedSecrets - all encryptedData keys will be accessible as environment variables inside container
```

2. Next, User runs a generation command e.g. `rkc backups generate backup --kubernetes --cron '*/30 * * *`

In result the User gets prepared Kubernetes manifests that could be applied to the cluster.
