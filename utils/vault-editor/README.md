# Vault Editor

Bridge uses an encrypted vault to store persistent data. This is a tool for reading and writing this vault.

Example usage:
```bash
$ ./vault-editor read
{
  "Settings": {
    "GluonDir": "/Users/james/Library/Caches/protonmail/bridge/gluon",
    "IMAPPort": 1143,
    "SMTPPort": 1025,
    "IMAPSSL": false,
    "SMTPSSL": false,
    "UpdateChannel": "stable",
    "UpdateRollout": 0.6046602879796196,
    "ColorScheme": "",
    "ProxyAllowed": true,
    "ShowAllMail": true,
    "Autostart": false,
    "AutoUpdate": true,
    "LastVersion": "2.4.1+git",
    "FirstStart": true,
    "FirstStartGUI": true
  },
  "Users": null,
  "Cookies": ...
  "Certs": {
    "Bridge": {
      "Cert": ...
      "Key": ...
    },
    "Installed": true
  }
}

$ ./vault-editor read > vault.json          # export the vault as JSON

$ vim vault.json                            # modify the exported vault somehow

$ cat vault.json|./vault-editor write       # import the modified vault

$ ./vault-editor read                       # the vault should have been modified
{
  "Settings": {
    "GluonDir": "/Users/james/Library/Caches/protonmail/bridge/gluon",
    "IMAPPort": 1144,
    "SMTPPort": 1026,
    "IMAPSSL": true,
    "SMTPSSL": true,
    "UpdateChannel": "early",
    "UpdateRollout": 0.6046602879796196,
    "ColorScheme": "",
    "ProxyAllowed": true,
    "ShowAllMail": true,
    "Autostart": false,
    "AutoUpdate": true,
    "LastVersion": "2.4.1+git",
    "FirstStart": true,
    "FirstStartGUI": true
  },
  "Users": null,
  "Cookies": ...
  "Certs": {
    "Bridge": {
      "Cert": ...
      "Key": ...
    },
    "Installed": true
  }
}
```