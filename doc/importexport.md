# Import-Export app

## Main blocks

This is basic overview of the main Import-Export blocks.

```mermaid
graph LR
    S[ProtonMail server]
    U[User]

    subgraph "Import-Export app"
        Users
        Frontend["Qt / CLI"]
        ImportExport
        Transfer

        Frontend --> ImportExport
        Frontend --> Transfer
        ImportExport --> Users
        ImportExport --> Transfer
    end

    EML --> Transfer
    MBOX --> Transfer
    IMAP --> Transfer
    S --> Transfer

    Transfer --> EML
    Transfer --> MBOX
    Transfer --> S

    U --> Frontend
```

## Code structure

More detailed graph of main types used in Import-Export app and connection between them.

```mermaid
graph TD
    PM[ProtonMail Server]
    EML[EML]
    MBOX[MBOX]
    IMAP[IMAP]

    subgraph "Import-Export app"
        subgraph "pkg users"
            subgraph "pkg credentials"
                CredStore[Store]
                Creds[Credentials]

                CredStore --> Creds
            end

            US[Users]
            U[User]

            US --> U
        end

        subgraph "pkg frontend"
            CLI
            Qt
        end

        subgraph "pkg importExport"
            IE[ImportExport]
        end

        subgraph "pkg transfer"
            Transfer
            Rules
            Progress

            Provider
            LocalProvider
            EMLProvider
            MBOXProvider
            IMAPProvider
            PMAPIProvider

            Mailbox
            Message

            Transfer --> |source|Provider
            Transfer --> |target|Provider
            Transfer --> Rules
            Transfer --> Progress

            Provider --> LocalProvider
            Provider --> EMLProvider
            Provider --> MBOXProvider
            Provider --> IMAPProvider
            Provider --> PMAPIProvider

            LocalProvider --> EMLProvider
            LocalProvider --> MBOXProvider

            Provider --> Mailbox
            Provider --> Message

        end

        subgraph PMAPI
            APIM[ClientManager]
            APIC[Client]

            APIM --> APIC
        end
    end

    CLI --> IE
    CLI --> Transfer
    CLI --> Progress
    Qt --> IE
    Qt --> Transfer
    Qt --> Progress

    U --> CredStore
    U --> Creds

    US --> APIM
    U --> APIM

    PMAPIProvider --> APIM
    EMLProvider --> EML
    MBOXProvider --> MBOX
    IMAPProvider --> IMAP

    IE --> US
    IE --> Transfer

    APIC --> PM
```
