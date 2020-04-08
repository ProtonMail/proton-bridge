Feature: IMAP search messages
  Background:
    Given there is connected user "user"
    Given there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Search by subject
    When IMAP client searches for "SUBJECT foo"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Search by text
    When IMAP client searches for "TEXT world"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Search by from
    When IMAP client searches for "FROM jane.doe@email.com"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Search by to
    When IMAP client searches for "TO user@pm.me"
    Then IMAP response is "OK"
    And IMAP response has 1 message
