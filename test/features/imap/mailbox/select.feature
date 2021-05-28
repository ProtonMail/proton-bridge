Feature: IMAP select into mailbox
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is IMAP client logged in as "user"

  Scenario: Select into inbox
    When IMAP client selects "INBOX"
    Then IMAP response is "OK"

  Scenario: Select into custom mailbox
    When IMAP client selects "Folders/mbox"
    Then IMAP response is "OK"

  Scenario: Select into non-existing mailbox
    When IMAP client selects "qwerty"
    Then IMAP response is "IMAP error: NO mailbox qwerty does not exist"
