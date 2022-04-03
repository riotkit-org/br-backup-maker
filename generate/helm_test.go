package generate_test

import (
    "github.com/riotkit-org/backup-maker/generate"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestRenderChart_InvalidSealedSecretMissingGpgEntry(t *testing.T) {
    sealedSecret := `
    apiVersion: bitnami.com/v1alpha1
    kind: SealedSecret
    metadata:
        name: mysecret
    spec:
        encryptedData:
            invalid-key-name: AgBy3i4OJSWK+PiTySYZZA9rO43cGDEq.....
    `

    templating := generate.Templating{}
    _, err := templating.RenderChart(
        "#!/bin/bash\necho 'Hello libertarian world!';",
        sealedSecret,
        "", // no schedule
        "priama-akcia",
        "alpine:3.14",
        map[interface{}]interface{}{}, // no helm values overridden
        "priama-akcia",
        "backup",
    )

    assert.Contains(t, err.Error(), "SealedSecret is invalid: missing .Spec.EncryptedData.gpg-key")
}

func TestRenderChart_ValidWithSealedSecret(t *testing.T) {
    sealedSecret := `
    apiVersion: bitnami.com/v1alpha1
    kind: SealedSecret
    metadata:
        name: mysecret
        namespace: priama-akcia
    spec:
        encryptedData:
            gpg-key: AgBy3i4OJSWK+PiTySYZZA9rO43cGDEq.....
    `

    templating := generate.Templating{}
    rendered, err := templating.RenderChart(
        "#!/bin/bash\necho 'Hello libertarian world!';",
        sealedSecret,
        "", // no schedule
        "priama-akcia",
        "alpine:3.14",
        map[interface{}]interface{}{}, // no helm values overridden
        "priama-akcia",
        "backup",
    )

    // contains all documents
    assert.Contains(t, rendered, "kind: CronJob")
    assert.Contains(t, rendered, "kind: SealedSecret")
    assert.Contains(t, rendered, "kind: Secret")
    assert.Contains(t, rendered, "kind: ConfigMap")

    // some minor things that should be
    assert.Contains(t, rendered, "image: alpine:3.14")
    assert.Contains(t, rendered, "restartPolicy: Never")

    assert.Nil(t, err)
}
