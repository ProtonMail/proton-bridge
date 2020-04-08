Feature: IMAP mailbox rename
  Background:
    Given there is connected user "user"

  Scenario: Rename folder
    Given there is "user" with mailbox "Folders/mbox"
    And there is IMAP client logged in as "user"
    When IMAP client renames mailbox "Folders/mbox" to "Folders/mbox2"
    Then IMAP response is "OK"
    And "user" does not have mailbox "Folders/mbox"
    And "user" has mailbox "Folders/mbox2"

  Scenario: Rename label
    Given there is "user" with mailbox "Labels/mbox"
    And there is IMAP client logged in as "user"
    When IMAP client renames mailbox "Labels/mbox" to "Labels/mbox2"
    Then IMAP response is "OK"
    And "user" does not have mailbox "Labels/mbox"
    And "user" has mailbox "Labels/mbox2"

  Scenario: Renaming folder to label is not possible
    Given there is "user" with mailbox "Folders/mbox"
    And there is IMAP client logged in as "user"
    When IMAP client renames mailbox "Folders/mbox" to "Labels/mbox"
    Then IMAP response is "IMAP error: NO cannot rename folder to non-folder"

  Scenario: Renaming system folder is not possible
    Given there is IMAP client logged in as "user"
    When IMAP client renames mailbox "INBOX" to "Folders/mbox"
    Then IMAP response is "IMAP error: NO cannot rename system mailboxes"
