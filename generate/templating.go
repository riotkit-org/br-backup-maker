package generate

import (
    "bytes"
    "errors"
    "fmt"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "os/user"
    "path/filepath"
    "strings"
    "text/template"
)

type Templating struct{}

// RenderTemplate renders a Go-formatted template in order from (stops on first found):
//              1. ~/.rkc/backups/templates/{backup,restore}/{Name}.tmpl
//              2. ~/.rkc/backups/templates/base/{backup,restore}/{Name}.tmpl
//
//              Templates in first directory are replaced only if the user has not modified them.
func (t *Templating) RenderTemplate(name string, operation string, variables interface{}) (string, error) {
    // load raw template content
    content, templatePath, err := t.loadTemplate(name, operation)
    if err != nil {
        return "", errors.New(fmt.Sprintf("cannot render template: %s", err))
    }

    // parse
    tpl := template.New(name).Option("missingkey=error")
    parsed, parseErr := tpl.Parse(string(content))
    if parseErr != nil {
        return "", errors.New(fmt.Sprintf("cannot render template: %s", err))
    }

    // render
    textBuffer := bytes.NewBufferString("")
    if err := parsed.Execute(textBuffer, variables); err != nil {
        return "", errors.New(fmt.Sprintf("cannot render template, execution failed. Error: %s, Template: %s", err, templatePath))
    }
    return textBuffer.String(), nil
}

// loadTemplate is reading templates in order from (stops on first found):
//              1. ~/.rkc/backups/templates/{backup,restore}/{Name}.tmpl
//              2. ~/.rkc/backups/templates/base/{backup,restore}/{Name}.tmpl
//
//              Templates in first directory are replaced only if the user has not modified them.
func (t *Templating) loadTemplate(name string, operation string) ([]byte, string, error) {
    paths := []string{
        "./cmd/backups/generate/templates/" + operation + "/" + name + ".tmpl", // only in testing
        "~/.rkc/backups/templates/" + operation + "/" + name + ".tmpl",
        "~/.rkc/backups/templates/base/" + operation + "/" + name + ".tmpl",
    }

    for _, path := range paths {
        path, expandErr := expandPath(path)
        if expandErr != nil {
            logrus.Warnf("Cannot expand path: %s", path)
        }
        if _, err := os.Stat(path); os.IsNotExist(err) {
            continue
        }

        b, err := ioutil.ReadFile(path)
        return b, path, err
    }

    return []byte(""), "", errors.New(fmt.Sprintf("template not found, looked in those paths: %s, use --template/-t to select a template", strings.Join(paths, ",")))
}

func (t *Templating) LoadVariables(path string) (map[string]interface{}, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, errors.New(fmt.Sprintf("cannot find file '%s'", path))
    }

    content, readErr := ioutil.ReadFile(path)
    if readErr != nil {
        return nil, errors.New(fmt.Sprintf("cannot read variables from file: '%s', error: %s", path, readErr))
    }

    var result map[string]interface{}
    if err := yaml.Unmarshal(content, &result); err != nil {
        return nil, errors.New(fmt.Sprintf("cannot read variables from file: '%s', error: '%s'", path, err))
    }

    return result, nil
}

// loadHelmValuesOverride should be called always after LoadVariables()
func (t *Templating) loadHelmValuesOverride(path string) (map[interface{}]interface{}, error) {
    content, _ := ioutil.ReadFile(path)

    var result map[interface{}]interface{}
    if err := yaml.Unmarshal(content, &result); err != nil {
        return map[interface{}]interface{}{}, errors.New(fmt.Sprintf("cannot read variables from file: '%s', error: '%s'", path, err))
    }

    // no ".HelmValues" defined
    if _, ok := result["HelmValues"]; !ok {
        return map[interface{}]interface{}{}, nil
    }

    return result["HelmValues"].(map[interface{}]interface{}), nil
}

func expandPath(path string) (string, error) {
    if len(path) == 0 || path[0] != '~' {
        return path, nil
    }

    usr, err := user.Current()
    if err != nil {
        return "", err
    }
    return filepath.Join(usr.HomeDir, path[1:]), nil
}
