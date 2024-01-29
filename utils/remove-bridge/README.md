# Clean Bridge

## Description

This repository contains scripts with which just by running them Bridge will quit, uninstall, and remove all leftover resources from your device. On Windows it will delete the entry(ies) in the Credential Manager, on macOS it will delete the entries from the Keychain Access, and on Linux distros it will try to delete the credentials both from `gnome-keyring` and `pass`.
There's a PowerShell script for Windows, and a Shell script for Linux & macOS.

---

### --WARNING--

These scripts are made with the assumption that there isn't another process with the `bridge` name in it. Be careful when using it on your own devices, it might kill another process with `bridge` in it's name.

The Shell script [remove_bridge](/remove_bridge) does not quit the process of Bridge V2. You'll need to manually quit it before executing this script. In the current implementation the process that the script quits is `bridge-gui` because when it was `bridge` previously, on Ubuntu the script closed itself not Bridge.

---

## Installation

There's no installation needed. Just download the script relevant for your Operating System, or clone the whole repo.

## Prerequisites

The `remove_bridge` script requires `zsh` to be executed.

## Usage

Run the script on your device when you need it.

On Linux distros it needs to be ran with `sudo` so it can uninstall Bridge.

### Linux & macOS

Recommendation for Linux & macOS is to place it in you `$PATH` location so you can run it immediately from the terminal when needed.

### Windows

To use the script without the need to input the full path of it, you can place it in a specific directory and add that directory to the PowerShell profile ([wiki link](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_profiles?view=powershell-7.3)) of the PowerShell version you are using so it's loaded in each session. Then you can just enter the name of the script to run it no matter the active directory in the session.

The `$PROFILE` automatic variable stores the paths to the PowerShell profiles that are available in the current session.
To view a profile path, display the value of the `$PROFILE` variable.

```PowerShell
PS C:\Users\Gjorgji> Write-Output $PROFILE
C:\Users\Gjorgji\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1
PS C:\Users\Gjorgji>
```

The `$PROFILE` variable stores the path to the "Current User, Current Host" profile. The other profiles are saved in note properties of the $PROFILE variable.
For example, the `$PROFILE` variable has the following values in the Windows PowerShell console.

- Current User, Current Host - $PROFILE
- Current User, Current Host - $PROFILE.CurrentUserCurrentHost
- Current User, All Hosts - $PROFILE.CurrentUserAllHosts
- All Users, Current Host - $PROFILE.AllUsersCurrentHost
- All Users, All Hosts - $PROFILE.AllUsersAllHosts

The script folder can be added to the profile at `$PROFILE.AllUsersAllHosts`, but where you place it it's up to you. The guide continues by just using `$PROFILE` for "Current User, Current Host".

#### **Create the Profile file**

By default, the Profile file is not created so you'll need to create it yourself before adding the script directory in it. If you have the Profile file already created skip to the [next step](#adding-the-script-directory-to-the-profile).

To create the file, open a PowerShell terminal and input:

```PowerShell
if (!(Test-Path -Path <profile-name>)) {
  New-Item -ItemType File -Path <profile-name> -Force
}
```

For example, to create a profile for the current user in the current PowerShell host application, use the following command:

```PowerShell
if (!(Test-Path -Path $PROFILE)) {
  New-Item -ItemType File -Path $PROFILE -Force
}
```

#### **Adding the script directory to the Profile**

To edit the $PROFILE file, open it with your favorite text editor, or just open your PowerShell terminal and input:

```PowerShell
PS C:\Users\Gjorgji> notepad $PROFILE
```

This will open the Profile file with the default Windows Notepad.

Once the file is opened, add the following and save the file:

```PowerShell
# Load scripts from the following locations
$env:Path += ";<path to script directory>"
```

As an example, if the script is placed in `C:\Users\Gjorgji\Documents\PowerShell\Scripts`, this line will be in the Profile:

```PowerShell
# Load scripts from the following location
$env:Path += ";$HOME\Documents\PowerShell\Scripts"
```

#### **Finally**

After doing all the above, restart your PowerShell terminal so the changes take effect, and whenever needed just run `Remove-Bridge`

```PowerShell
╭─    ~\bridge    devel ≡  ?1 ~1                                                    ✔  08:43:49  ─╮
╰─ Remove-Bridge                                                                                                     ─╯
All Bridge resource folders deleted!

CMDKEY: Credential deleted successfully.
╭─    ~\bridge    devel ≡  ?1 ~1                                        37.152s   ✔  08:45:43  ─╮
╰─                                                                                                                   ─╯
```
