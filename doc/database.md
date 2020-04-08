# Database

Bridge needs to have a small database to pair our IDs with IMAP UIDs and indexes. IMAP protocol
requires every message to have an unique UID in mailbox. In this context, mailbox is not an account,
but a folder or label. This means that one message can have more UIDs, one for each mailbox (folder), 
and that two messages can have the same UID, but each for different mailbox (folder).

IMAP index is just an index. Look at it like to an array: `["UID1", "UID2", "UID3"]`. We can access
message by UID or index; for example index 2 and UID `UID2`. When this message is deleted, we need
to re-index all following messages. The array will look now like `["UID1", "UID3"]` and the last
message can be accessed by index 2 or UID `UID3`.

See RFCs for more information:

* https://tools.ietf.org/html/rfc822
* https://tools.ietf.org/html/rfc3501

Our database is currently built on BBolt and have those buckets (key-value storage):

* Message metadata bucket:

    * `[metadataBucket][API_ID] -> pmapi.Message{subject, from, to, size, other headers...}` (without body or attachment)

* Mapping buckets

    * `[mailboxesBucket][addressID-mailboxID][api_ids][API_ID] -> UID`
    * `[mailboxesBucket][addressID-mailboxID][imap_ids][UID] -> API_ID`
