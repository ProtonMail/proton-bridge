Select-String -Path (Join-Path $PSScriptRoot "../Makefile") -Pattern "^BRIDGE_APP_VERSION\?=(\S*)" |
        ForEach-Object {$_.Matches} | ForEach-Object { $_.Groups[1].Value }