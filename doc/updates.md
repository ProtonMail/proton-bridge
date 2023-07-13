# Update mechanism of Bridge

There are multiple options how to change version of application:
* Automatic in-app update
* Manual in-app update
* Manual install

In-app update ends with restarting bridge into new version. Automatic in-app
update is downloading, verifying and installing the new version immediately
without user confirmation. For manual in-app update user needs to confirm first.
Update is done from special update file published on website.

The manual installation requires user to download, verify and install manually
using installer for given OS.

The bridge is installed and executed differently for given OS:

* Windows and Linux apps are using launcher mechanism:
    * There is system protected installation path which is created on first
      install. It contains bridge exe and launcher exe. When users starts
      bridge the launcher is executed first. It will check update path compare
      version with installed one. The newer version then is then executed.
    * Update mechanism means to replace files in update folder which is located
      in user space.

* macOS app does not use launcher
    * No launcher, only one executable
    * In-App update replaces the bridge files in installation path directly


```mermaid
flowchart LR
    subgraph Frontend
    U[User requests<br>version check]
    ManIns((Notify user about<br>manual install<br>is needed))
    R((Notify user<br>about restart))
    ManUp((Notify user about<br>manual update))
    NF((Notify user about<br>force update))

    ManUp -->|Install| InstFront[Install]
    InstFront -->|Ok| R
    InstFront -->|Error| ManIns

    U --> CheckFront[Check online]
    CheckFront -->|Ok| IAFront{Is new version<br>and applicable?}
    CheckFront -->|Error| ManIns

    IAFront -->|No| Latest((Notify user<br>has latest version))
    IAFront -->|Yes| CanInstall{Can update?}
    CanInstall -->|No| ManIns
    CanInstall -->|Yes| NotifOrInstall{Is automatic<br>update enabled?}
    NotifOrInstall -->|Manual| ManUp
    end


    subgraph Backend
    W[Wait for next check]

    W --> Check[Check online]

    Check --> NV{Has new<br>version?}
    Check -->|Error| W
    NV -->|No new version| W
    IA{Is install<br>applicable?}
    NV -->|New version<br>available| IA
    IA -->|Local rollout<br>not enough| W
    IA -->|Yes| AU{Is automatic\nupdate enabled?}

    AU -->|Yes| CanUp{Can update?}
    CanUp -->|No| ManIns

    CanUp -->|Yes| Ins[Install]
    Ins -->|Error| ManIns
    Ins -->|Ok| R

    AU -->|No| ManUp
    ManUp -->|Ignore| W


    F[Force update]
    F --> NF
    end

    ManIns --> Web[Open web page]
    NF --> Web
    ManUp --> Web
    R --> Re[Restart]
    NF --> Q[Quit bridge]
    NotifOrInstall -->|Automatic| W
```


The non-trivial is to combine the update with setting change:
* turn off/on automatic in-app updates
* change from stable to beta or back

_TODO fill flow chart details_


We are not support downgrade functionality. Only some circumstances can lead to
downgrading the app version.

_TODO fill flow chart details_
