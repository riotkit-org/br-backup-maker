#!/usr/bin/env python3
import unittest
import subprocess

class BackupMakerProcedureGeneratorTest(unittest.TestCase):
    def test_tar_backup_and_restore(self):
        token = "asd"
        definition = f"""
---
Params:
    path: ./cmd

Repository:
    url: "http://127.0.0.1:30100"
    token: "{token}"
    encryptionKeyPath: "resources/test/gp-key.asc"
    passphrase: "riotkit"
    recipient: "test@riotkit.org"
    collectionId: "iwa-ait"
        """

        with open(".build/definition.yaml", "w") as f:
            f.write(definition)

        subprocess.check_call([
            "./.build/bmg", "backup", "--definition", ".build/definition.yaml", "--output-dir", ".build/",
            "--template", "tar"
        ])

        # todo:
        # execute backup.sh
        # remove files in ./cmd directory
        # restore backup
        # check if files exists
