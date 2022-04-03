package generate_test

import (
    "github.com/riotkit-org/backup-maker/generate"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestRenderTemplate_WithValidParams(t *testing.T) {
    params := map[string]string{
        "hostname": "postgres.db.svc.cluster.local",
        "port":     "5432",
        "db":       "rkc-test",
        "user":     "riotkit",
        "password": "${DB_PASSWORD}",
    }

    repository := map[string]string{
        "url":               "https://example.org",
        "token":             "${BR_TOKEN}",
        "encryptionKeyPath": "/var/lib/backup-repository/encryption.key",
        "passphrase":        "${GPG_PASSPHRASE}",
        "recipient":         "your-gpg@email.org",
        "collectionId":      "111-222-333-444",
    }

    template := generate.Templating{}
    out, err := template.RenderTemplate("postgres", "backup", map[string]interface{}{
        "Params":     params,
        "Repository": repository,
    })

    // will contain:
    //   - plaintext value like 'riotkit' as user
    //   - not processed '${DB_PASSWORD}' as password
    assert.Contains(t, out, "pgbr db backup --password '${DB_PASSWORD}' --user 'riotkit' --db-name 'rkc-test' --port '5432'")
    assert.Contains(t, out, "exec backup-maker make") // the last command is backup-maker (may change within the template, just like line above)
    assert.Nil(t, err)
}

func TestRenderTemplate_FailsWhenAnyVariableIsMissing(t *testing.T) {
    params := map[string]string{
        "hostname": "postgres.db.svc.cluster.local",
        "port":     "5432",
        // NOTICE: a few variables there are missing: db, user, password
    }

    repository := map[string]string{
        "url":               "https://example.org",
        "token":             "${BR_TOKEN}",
        "encryptionKeyPath": "/var/lib/backup-repository/encryption.key",
        "passphrase":        "${GPG_PASSPHRASE}",
        "recipient":         "your-gpg@email.org",
        "collectionId":      "111-222-333-444",
    }

    template := generate.Templating{}
    _, err := template.RenderTemplate("postgres", "backup", map[string]interface{}{
        "Params":     params,
        "Repository": repository,
    })

    assert.NotNil(t, err)
    assert.Contains(t, err.Error(), "templates/backup/postgres.tmpl")
    assert.Contains(t, err.Error(), "map has no entry for key")
}
