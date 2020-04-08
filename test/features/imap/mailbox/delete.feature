Feature: IMAP delete mailbox
  Background:
    Given there is connected user "user"

  Scenario: Delete folder
    Given there is "user" with mailbox "Folders/mbox"
    And there is IMAP client logged in as "user"
    When IMAP client deletes mailbox "Folders/mbox"
    Then IMAP response is "OK"
    And "user" does not have mailbox "Folders/mbox"

  Scenario: Delete label
    Given there is "user" with mailbox "Labels/mbox"
    And there is IMAP client logged in as "user"
    When IMAP client deletes mailbox "Labels/mbox"
    Then IMAP response is "OK"
    And "user" does not have mailbox "Labels/mbox"

  Scenario: Empty Trash by deleting it
    Given there are 10 messages in mailbox "Trash" for "user"
    And there is IMAP client logged in as "user"
    When IMAP client deletes mailbox "Trash"
    Then IMAP response is "OK"
    And mailbox "Trash" for "user" has 0 messages

  Scenario: Deleting system mailbox is not possible
    Given there is IMAP client logged in as "user"
    When IMAP client deletes mailbox "INBOX"
    Then IMAP response is "IMAP error: NO cannot empty mailbox 0"
