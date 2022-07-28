Feature: IMAP list mailboxes
  Background:
    Given there is connected user "user"

  Scenario: List mailboxes
    Given there is "user" with mailbox "Folders/mbox1"
    And there is "user" with mailbox "Labels/mbox2"
    And there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"
    Then IMAP response contains "Folders/mbox1"
    Then IMAP response contains "Labels/mbox2"

  Scenario: List mailboxes without All Mail
    Given there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"
    When All mail mailbox is hidden
    And IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response doesn't contain "All Mail"
    When All mail mailbox is visible
    And IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"

  @ignore-live
  Scenario: List mailboxes with subfolders
    # Escaped slash in the name contains slash in the name.
    # Not-escaped slash in the name means tree structure.
    # We keep escaping in an IMAP communication so each mailbox is unique and
    # both mailboxes are accessible. The slash is visible in the IMAP client.
    Given there is "user" with mailbox "Folders/a\/b"
    And there is "user" with mailbox "Folders/a/b"
    And there is IMAP client logged in as "user"
    When IMAP client lists mailboxes
    Then IMAP response contains "Folders/a\/b"
    Then IMAP response contains "Folders/a/b"
