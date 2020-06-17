Feature: IMAP list mailboxes
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox1"
    And there is "user" with mailbox "Labels/mbox2"
    And there is IMAP client logged in as "user"

  Scenario: List mailboxes
    When IMAP client lists mailboxes
    Then IMAP response contains "INBOX"
    Then IMAP response contains "Sent"
    Then IMAP response contains "Archive"
    Then IMAP response contains "Trash"
    Then IMAP response contains "All Mail"
    Then IMAP response contains "Folders/mbox1"
    Then IMAP response contains "Labels/mbox2"
