package generate

import (
    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"
    "io/ioutil"
)

type SnippetGenerationCommand struct {
    Template       string
    DefinitionFile string
    IsKubernetes   bool
    KeyPath        string
    OutputDir      string
    Schedule       string
    JobName        string
    Image          string
    Operation      string
    Namespace      string
    // todo: Helm templates path?
}

func (c *SnippetGenerationCommand) Run() error {
    t := Templating{}
    variables, loadErr := t.LoadVariables(c.DefinitionFile)
    if loadErr != nil {
        return loadErr
    }

    // todo: JSON schema

    rendered, err := t.RenderTemplate(c.Template, c.Operation, variables)
    if err != nil {
        return err
    }

    // GPG
    logrus.Infof("Loading GPG key from '%s'", c.KeyPath)
    gpgKeyContent, err := ioutil.ReadFile(c.KeyPath)
    if err != nil {
        return errors.Wrapf(err, "cannot read gpg key from path '%s'", c.KeyPath)
    }

    // output format: Kubernetes
    if c.IsKubernetes {
        helmValuesOverride, err := t.loadHelmValuesOverride(c.DefinitionFile)
        if err != nil {
            return errors.Wrap(err, "cannot read helm values override from yaml file")
        }

        renderedChart, helmErr := t.RenderChart(rendered, string(gpgKeyContent), c.Schedule, c.JobName, c.Image, helmValuesOverride, c.Namespace, c.Operation)
        if helmErr != nil {
            return helmErr
        }
        writeErr := writeFiles([]targetFile{
            {Name: c.OutputDir + "/" + c.Operation + ".yaml", Content: renderedChart},
        })
        if writeErr != nil {
            return errors.Wrap(writeErr, "cannot write files")
        }
        return nil
    }

    // output format: default/plain
    writeErr := writeFiles([]targetFile{
        {Name: c.OutputDir + "/" + c.Operation + ".sh", Content: rendered},
        {Name: c.OutputDir + "/gpg-key", Content: string(gpgKeyContent)},
    })
    if writeErr != nil {
        return errors.Wrap(writeErr, "cannot write files")
    }

    return nil
}

type targetFile struct {
    Name    string
    Content string
}

func writeFiles(files []targetFile) error {
    for _, file := range files {
        contentAsByte := []byte(file.Content)
        logrus.Infof("Writing file '%s'", file.Name)
        if err := ioutil.WriteFile(file.Name, contentAsByte, 0755); err != nil {
            return err
        }
    }
    return nil
}
