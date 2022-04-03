Backup Maker
============

[![Test](https://github.com/riotkit-org/br-backup-maker/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/riotkit-org/br-backup-maker/actions/workflows/test.yaml)

Tiny backup client packed in a single binary. Interacts with a `Backup Repository` server to store files, uses GPG to secure your
backups even against the server administrator.

**Features:**
- Captures output from user-defined Backup/Restore commands
- Automated, optional GPG support enables easy to use E2E encryption
- Buffered upload of backup made on-the-fly requires no additional disk space to create backup
- Small, single binary, can be injected into container or distributed as a lightweight container

**Notice:** You need to have backup of your encryption private key. **Lost encryption key means your backups are unreadable!**

# Usage

## Getting backup-maker

Take a look at releases tab and pick a version suitable for your platform. We support Unix-like platforms, there is no support for Windows.

You can use [eget](https://github.com/zyedidia/eget) as a 'package manager' to install `backup-maker`

```bash
# for pre-release
eget --pre-release riotkit-org/br-backup-maker --to /usr/local/bin/backup-maker

# for latest stable release
eget riotkit-org/br-backup-maker --to /usr/local/bin/backup-maker
```

## Creating backup

```bash
# most of commandline switches can be replaced with environment variables, check the table in other section of documentation
export BM_AUTH_TOKEN="some-token"; \
export BM_COLLECTION_ID="111-222-333-444"; \
export BM_PASSPHRASE="riotkit"; \
backup-maker make --url https://example.org \
    -c "tar -zcvf - ./" \
    --key build/test/backup.key \
    --recipient test@riotkit.org \
    --log-level info
```

## Restoring a backup

```bash
# commandline switches could be there also replaced with environment variables
backup-maker restore --url $$(cat .build/test/domain.txt) \
    -i $$(cat .build/test/collection-id.txt) \
    -t $$(cat .build/test/auth-token.txt) \
    -c "cat - > /tmp/test" \
    --private-key .build/test/backup.key \
    --passphrase riotkit \
    --recipient test@riotkit.org \
    --log-level debug
```

## Backup - How it works?

This list of steps includes only steps that are done inside `Backup Maker`, to understand whole flow
please take a look at `Backup Controller` documentation.

**Note: GPG steps are optional**

1. `gpg` keyring is created in a temporary directory, keys are imported
2. Command specified in `--cmd` or in `-c` is executed
3. Result of the command, it's stdout is transferred to the `gpg` process
4. From `gpg` process the encoded data is buffered directly to the server
5. Feedback is returned
6. Temporary `gpg` keyring is deleted

## Restore - How it works?

It is very similar as in backup operation.

1. `gpg` keyring is created in a temporary directory, keys are imported
2. Command specified in `--cmd` or in `-c` is executed
3. `gpg` process is started
4. Backup download is starting
5. Backup is transmitted on the fly from server to `gpg` -> our shell command
6. Our shell `--cmd` / `-c` command is taking stdin and performing a restore action
7. Feedback is returned
8. Temporary `gpg` keyring is deleted

## Automated procedures

Our suggested approach is to maintain a community-driven repository of automation scripts templates
together with a tool that generates Backup & Restore procedures. Those procedures could be easily understood and be customized by the user.

### [Documentation for 'bmg' (Backup Maker procedure Generator)](./generate/README.md)

## Hints

- Skip `--private-key` and `--passphrase` to disable GPG
- Use `debug` log level to see GPG output and more verbose output at all


## Proposed usage

### Scenario 1: Standalone binary running from crontab

Just schedule a cronjob that would trigger `backup-maker make` with proper switches. Create a helper script to easily restore backup as a part
of a disaster recovery plan.

### Scenario 2: Dockerized applications, keep it inside application container

Pack `backup-maker` into docker image and trigger backups from internal or external crontab, jobber or other scheduler.

### Scenario 3: Kubernetes usage with plain `kind: Crojob` resources

Use [bmg](./generate/README.md) to generate Kubernetes resources that could be applied to cluster with `kubectl` or added to repository and applied by [FluxCD](https://fluxcd.io/) or [ArgoCD](https://argo-cd.readthedocs.io/en/stable/).

### Scenario 4: Kubernetes usage with Argo Workflows or Tekton

Create a definition of an [Argo Workflow](https://argoproj.github.io/argo-workflows/) or [Tekton Pipeline](https://tekton.dev/) that will spawn a Kubernetes job with defined token, collection id, command, GPG key.

Environment variables
---------------------

Environment variables are optional, if present will cover values of appropriate commandline switches.

| Type    | Name                | Description                                                                               |
|---------|---------------------|-------------------------------------------------------------------------------------------|
| path    | BM_PUBLIC_KEY_PATH  | Path to the public key used for encryption                                                |
| string  | BM_CMD              | Command used to encrypt or decrypt (depends on context)                                   |
| string  | BM_PASSPHRASE       | Passphrase for the GPG key                                                                |
| string  | BM_VERSION          | Version to restore (defaults to "latest"), e.g. v1                                        |
| email   | BM_RECIPIENT        | E-mail address of GPG recipient key                                                       |
| url     | BM_URL              | Backup Repository URL address e.g. https://example.org                                    |
| uuidv4  | BM_COLLECTION_ID    | Existing collection ID                                                                    |
| jwt     | BM_AUTH_TOKEN       | JSON Web Token generated in Backup Repository that allows to write to given collection id |
| integer | BM_TIMEOUT          | Connection and read timeouts in seconds                                                   |
| path    | BM_PRIVATE_KEY_PATH | GPG private key used to decrypt backup                                                    |
