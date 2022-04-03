package generate

//
// Extracts templates from binary into user home directory
//
// Rules:
//   - .base directories are always overwritten when CLI application is launched
//   - user directory allows for user customization. Files will be overwritten by app util user changes at least one file
//
// Example structure:
//
//   ~/.bm
//   ├── chart
//   │         ├── .base
//   │         │         ├── configmap.yaml
//   │         │         ├── cronjob.yaml
//   │         │         └── secret.yaml
//   │         └── user
//   │             ├── configmap.yaml
//   │             ├── cronjob.yaml
//   │             └── secret.yaml
//   └── templates
//      ├── .base
//      │         ├── backup
//      │         │         ├── files.tmpl
//      │         │         └── postgres.tmpl
//      │         └── restore
//      │             ├── files.tmpl
//      │             └── postgres.tmpl
//      └── user
//          ├── backup
//          │         ├── files.tmpl
//          │         └── postgres.tmpl
//          └── restore
//              ├── files.tmpl
//              └── postgres.tmpl
//

import (
    "embed"
    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"
    "golang.org/x/mod/sumdb/dirhash"
    "io/ioutil"
    "os"
    "path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed chart/*
var chartFS embed.FS

func extract() error {
    templatesMainPath, _ := ExpandPath("~/.bm/templates")
    chartsMainPath, _ := ExpandPath("~/.bm/chart")

    templatesModifiedByUser := checkModifiedByUser(templatesMainPath)
    chartsModifiedByUser := checkModifiedByUser(chartsMainPath)

    for _, key := range []string{".base", "user"} {
        // templates
        if (!templatesModifiedByUser && key == "user") || key == ".base" {
            templatesPath, pathErr := ExpandPath("~/.bm/templates/" + key)
            if pathErr != nil {
                return pathErr
            }
            if err := extractRecursively(templatesFS, "templates", "templates", templatesPath); err != nil {
                return err
            }
        }

        // helm chart
        if (!chartsModifiedByUser && key == "user") || key == ".base" {
            chartPath, chartPathErr := ExpandPath("~/.bm/chart/" + key)
            if chartPathErr != nil {
                return chartPathErr
            }
            if err := extractRecursively(chartFS, "chart", "chart", chartPath); err != nil {
                return err
            }
        }
    }

    return nil
}

func extractRecursively(fs embed.FS, baseDirName string, path string, target string) error {
    files, err := fs.ReadDir(path)
    if err != nil {
        return err
    }

    for _, entry := range files {
        entryPath := path + "/" + entry.Name()

        if entry.IsDir() {
            if err := extractRecursively(fs, baseDirName, entryPath, target); err != nil {
                return err
            }
            continue
        }

        logrus.Debugf("Extracting file '%s'", entryPath)
        data, readErr := fs.ReadFile(entryPath)
        if readErr != nil {
            return errors.Wrap(readErr, "Cannot read file from go:embed")
        }

        entryPath = entryPath[len(baseDirName)+1:] // removes the redundant duplicate base directory created deeply twice e.g. templates/base/templates/...
        fileToBeCreated := target + "/" + entryPath

        if err := os.MkdirAll(filepath.Dir(fileToBeCreated), 0700); err != nil {
            return err
        }
        if err := ioutil.WriteFile(fileToBeCreated, data, 0700); err != nil {
            return errors.Wrap(err, "Cannot write destination file")
        }
    }

    return nil
}

func checkModifiedByUser(path string) bool {
    base, _ := dirhash.HashDir(path+"/.base", "", dirhash.Hash1)
    user, _ := dirhash.HashDir(path+"/user", "", dirhash.Hash1)

    return base != user
}
