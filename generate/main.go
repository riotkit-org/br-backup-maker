package generate

import (
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

// NewBackupCommand generates a backup procedure command
func NewBackupCommand() *cobra.Command {
    app := &SnippetGenerationCommand{}

    command := &cobra.Command{
        Use:   "backup",
        Short: "Generates a backup procedure",
        Run: func(command *cobra.Command, args []string) {
            err := app.Run()

            if err != nil {
                logrus.Errorf(err.Error())
            }
        },
    }

    // command.Example = ""
    // command.Long = ""

    command.Flags().StringVarP(&app.Template, "template", "t", "", "Template Name e.g. 'postgres', 'mysql', 'gitea', 'redis', 'files', 'wordpress'")
    command.Flags().StringVarP(&app.DefinitionFile, "definition", "d", "./rkc-backup.yaml", "Backup & Restore definition in YAML format, see reference in docs")
    command.Flags().BoolVarP(&app.IsKubernetes, "kubernetes", "k", false, "Generate output in Kubernetes manifests format")
    command.Flags().StringVarP(&app.KeyPath, "gpg-key-path", "g", "gpg-key", "Path to the GPG key (private or public, recommended to use public key)")
    command.Flags().StringVarP(&app.OutputDir, "output-dir", "o", "./", "Path where to store output files")
    command.Flags().StringVarP(&app.Schedule, "k8s-job-schedule", "", "16 1 * * *", "Cronjob schedule (if using --kubernetes)")
    command.Flags().StringVarP(&app.JobName, "k8s-name", "", "my-backup-job", "Resources Name (if using --kubernetes)")
    command.Flags().StringVarP(&app.Image, "k8s-image", "", "ghcr.io/riotkit-org/backup-maker-env:latest", "Image (if using --kubernetes)")
    command.Flags().StringVarP(&app.Namespace, "k8s-namespace", "n", "", "Namespace (if using --kubernetes)")
    app.Operation = "backup"

    return command
}

func NewRestoreCommand() *cobra.Command {
    app := &SnippetGenerationCommand{}

    command := &cobra.Command{
        Use:   "restore",
        Short: "Generates a restore procedure",
        Run: func(command *cobra.Command, args []string) {
            err := app.Run()

            if err != nil {
                logrus.Errorf(err.Error())
            }
        },
    }

    command.Flags().StringVarP(&app.Template, "template", "t", "", "Template Name e.g. 'postgres', 'mysql', 'gitea', 'redis', 'files', 'wordpress'")
    command.Flags().StringVarP(&app.DefinitionFile, "definition", "d", "./rkc-backup.yaml", "Backup & Restore definition in YAML format, see reference in docs")
    command.Flags().BoolVarP(&app.IsKubernetes, "kubernetes", "k", false, "Generate output in Kubernetes manifests format")
    command.Flags().StringVarP(&app.KeyPath, "gpg-key-path", "g", "gpg-key", "Path to the GPG key (private or public, recommended to use public key)")
    command.Flags().StringVarP(&app.OutputDir, "output-dir", "o", "./", "Path where to store output files")
    command.Flags().StringVarP(&app.JobName, "k8s-name", "", "my-backup-job", "Resources Name (if using --kubernetes)")
    command.Flags().StringVarP(&app.Image, "k8s-image", "", "ghcr.io/riotkit-org/backup-maker-env:latest", "Image (if using --kubernetes)")
    command.Flags().StringVarP(&app.Namespace, "k8s-namespace", "n", "", "Namespace (if using --kubernetes)")
    app.Operation = "restore"
    app.Schedule = ""

    return command
}

// Main creates the new command
func Main() *cobra.Command {
    if err := extract(); err != nil {
        logrus.Fatal(err)
    }

    cmd := &cobra.Command{
        Use:   "bmg",
        Short: "",
        Run: func(cmd *cobra.Command, args []string) {
            err := cmd.Help()
            if err != nil {
                logrus.Errorf(err.Error())
            }
        },
    }
    cmd.AddCommand(NewBackupCommand())
    cmd.AddCommand(NewRestoreCommand())

    return cmd
}
