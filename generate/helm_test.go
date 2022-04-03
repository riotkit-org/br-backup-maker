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

func TestRenderChart_InvalidSealedSecretMismatchInNamespace(t *testing.T) {
    sealedSecret := `
    apiVersion: bitnami.com/v1alpha1
    kind: SealedSecret
    metadata:
        name: mysecret
        namespace: OTHER-MISMATCHED
    spec:
        encryptedData:
            gpg-key: AgBy3i4OJSWK+PiTySYZZA9rO43cGDEq.....
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

    assert.Contains(t, err.Error(), "SealedSecret is invalid: SealedSecret is in different Namespace (OTHER-MISMATCHED), expected to be in 'priama-akcia'")
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

func TestRenderChart_ValidPlainGPGKeyInSecret(t *testing.T) {
    gpgKey := `
    ------ gpg private key blah blah blah ------
    .... secret key Putin chuj, all politics are dickheads ....
    ------ end of blah blah blah ------
    `

    templating := generate.Templating{}
    rendered, err := templating.RenderChart(
        "#!/bin/bash\necho 'Hello libertarian world!';",
        gpgKey,
        "", // no schedule
        "priama-akcia",
        "alpine:3.14",
        map[interface{}]interface{}{}, // no helm values overridden
        "priama-akcia",
        "backup",
    )

    // contains all documents
    assert.Contains(t, rendered, "kind: CronJob")
    assert.Contains(t, rendered, "kind: Secret")
    assert.Contains(t, rendered, "kind: ConfigMap")

    // gpg
    assert.Contains(t, rendered, "all politics are dickheads") // contains gpg key in plan format (should be inside `kind: Secret`)
    assert.NotContains(t, rendered, "kind: SealedSecret")      // we use plain GPG key in this case!

    assert.Nil(t, err)
}

// covers processVariablesLocally() and evaluateShell()
func TestRenderChart_ValidWithEvaluatedHelmValuesInLocalShellAtBuildStage(t *testing.T) {
    gpgKey := `
    ------ gpg private key blah blah blah ------
    .... secret key Putin chuj, all politics are dickheads ....
    ------ end of blah blah blah ------
    `

    environment := map[interface{}]interface{}{
        "SOME_THING": "$(uname) is the real free OS", // should show operating system name "Linux"
    }
    variables := map[interface{}]interface{}{
        "env": environment,
    }

    templating := generate.Templating{}
    rendered, err := templating.RenderChart(
        "#!/bin/bash\necho 'Hello libertarian world!';",
        gpgKey,
        "", // no schedule
        "priama-akcia",
        "alpine:3.14",
        variables,
        "priama-akcia",
        "backup",
    )

    assert.Contains(t, rendered, "Linux is the real free OS") // somewhere inside `kind: Secret` there should be a variable defined with this value
    assert.Nil(t, err)
}
