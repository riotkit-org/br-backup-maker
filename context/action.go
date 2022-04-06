package context

import "strings"

type Action struct {
    Url          string
    CollectionId string
    AuthToken    string
    Command      string
    Timeout      int64
    ActionType   string

    VersionToRestore string
    DownloadPath     string

    Gpg      GPGOperationContext
    LogLevel uint32
}

// GetCommand returns a valid command that includes GPG encryption/decryption (if using)
// this method is fully context aware, it understands if we are uploading or downloading a backup therefore
// if decryption or encryption is performed
func (that Action) GetCommand(custom string) string {
    cmd := that.Command

    if custom != "" {
        cmd = custom
    }

    if !that.Gpg.Enabled(that.ActionType) {
        return cmd
    }

    if that.ActionType == "make" {
        return cmd + " | " + that.Gpg.GetEncryptionCommand()
    }

    return that.Gpg.GetDecryptionCommand() + " | " + cmd
}

// GetPrintableCommand returns same command as in GetCommand(), but with erased credentials
// so the command could be logged or printed into the console
func (that Action) GetPrintableCommand(custom string) string {
    return strings.ReplaceAll(that.GetCommand(custom), that.Gpg.Passphrase, "***")
}

func (that Action) ShouldShowCommandsOutput() bool {
    return that.LogLevel >= 5
}
