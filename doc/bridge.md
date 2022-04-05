# Bridge

## Main blocks

This is basic overview of the main bridge blocks.

Note connection between IMAP/SMTP and PMAPI. IMAP and SMTP packages are in the queue to be refactored
and we would like to try to have functionality in bridge core or bridge utilities (such as messages)
than direct usage of PMAPI from IMAP or SMTP. Also database (BoltDB) should be moved to bridge core.

```mermaid
graph LR
    S[Server]
    C[Client]
    U[User]

    subgraph "Bridge app"
        Core[Bridge core]
        API[PMAPI]
        Store
        DB[BoltDB]
        Frontend["Qt / CLI"]
        IMAP
        SMTP

        IMAP --> Store
        IMAP --> Core
        SMTP --> Core
        SMTP --> API
        Core --> API
        Core --> Store
        Store --> API
        Store --> DB
        Frontend --> Core

    end

    C --> IMAP
    C --> SMTP
    U --> Frontend
    API --> S
```

## Code structure

More detailed graph of main types used in Bridge app and connection between them. Here is already
communication to PMAPI only from bridge core which is not true, yet. IMAP and SMTP are still calling
PMAPI directly.

```mermaid
graph TD

    C["Client (e.g. Thunderbird)"]
    PM[Proton Mail Server]

    subgraph "Bridge app"
        subgraph "Bridge core"
            B[Bridge]
            U[User]

            B --> U
        end

        subgraph Store
            StoreU[Store User]
            StoreA[Address]
            StoreM[Mailbox]

            StoreU --> StoreA
            StoreA --> StoreM
        end

        subgraph Credentials
            CredStore[Store]
            Creds[Credentials]

            CredStore --> Creds
        end

        subgraph Frontend
            CLI
            Qt
        end

        subgraph IMAP
            IB[IMAP backend]
            IA[IMAP address]
            IM[IMAP mailbox]

            IB --> B
            IB --> IA
            IA --> IM
            IA --> U
            IA --> StoreA
            IM --> StoreM
        end

        subgraph SMTP
            SB[SMTP backend]
            SS[SMTP session]

            SB --> B
            SB --> SS
            SS --> U
        end
    end

    subgraph PMAPI
        AC[Client]
    end

    C --> IB
    C --> SB

    CLI --> B
    Qt --> B

    U --> CredStore
    U --> Creds

    U --> StoreU

    StoreU --> AC
    StoreA --> AC
    StoreM --> AC

    B --> AC
    U --> AC

    AC --> PM
```

## How to debug

Run `make run-debug` which starts [Delve](https://github.com/go-delve/delve).
