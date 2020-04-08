Feature: IMAP create mailbox
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"

  Scenario: Create folder
    When IMAP client creates mailbox "Folders/mbox"
    Then IMAP response is "OK"
    And "user" has mailbox "Folders/mbox"
    And "user" does not have mailbox "Labels/mbox"

  Scenario: Create label
    When IMAP client creates mailbox "Labels/mbox"
    Then IMAP response is "OK"
    And "user" does not have mailbox "Folders/mbox"
    And "user" has mailbox "Labels/mbox"

  Scenario: Creating system mailbox is not possible
    When IMAP client creates mailbox "INBOX"
    Then IMAP response is "IMAP error: NO mailbox INBOX already exists"

  Scenario: Creating mailbox without prefix is not possible
    When IMAP client creates mailbox "mbox"
    Then IMAP response is "OK"
    And "user" does not have mailbox "mbox"
