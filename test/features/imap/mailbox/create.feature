Feature: IMAP create mailbox
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"
    And "user" does not have mailbox "Folders/mbox"
    And "user" does not have mailbox "Labels/mbox"

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

  Scenario: Creating label with existing name is not possible
    Given there is "user" with mailbox "Folders/mbox"
    When IMAP client creates mailbox "Labels/mbox"
    Then IMAP response is "IMAP error: NO A label or folder with this name already exists"
    And "user" has mailbox "Folders/mbox"
    And "user" does not have mailbox "Labels/mbox"

  Scenario: Creating folder with existing name is not possible
    Given there is "user" with mailbox "Labels/mbox"
    When IMAP client creates mailbox "Folders/mbox"
    Then IMAP response is "IMAP error: NO A label or folder with this name already exists"
    And "user" has mailbox "Labels/mbox"
    And "user" does not have mailbox "Folders/mbox"

  Scenario: Creating system mailbox is not possible
    When IMAP client creates mailbox "INBOX"
    Then IMAP response is "IMAP error: NO mailbox INBOX already exists"
    When IMAP client creates mailbox "Folders/INBOX"
    Then IMAP response is "IMAP error: NO Invalid name"
    # API allows you to create custom folder with naem `All Mail`
    #When IMAP client creates mailbox "Folders/All mail"
    #Then IMAP response is "IMAP error: NO mailbox All Mail already exists"

  Scenario: Creating mailbox without prefix is not possible
    When IMAP client creates mailbox "mbox"
    Then IMAP response is "OK"
    And "user" does not have mailbox "mbox"
    When All mail mailbox is hidden
    And IMAP client creates mailbox "All mail"
    Then IMAP response is "OK"
    And "user" does not have mailbox "All mail"
