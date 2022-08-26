Feature: IMAP get mailbox status
  Background:
    Given there is connected user "user"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | starred |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Mailbox status contains mailbox name
    When IMAP client gets status of "INBOX"
    Then IMAP response contains "INBOX"

  Scenario: Mailbox status contains counts and UIDs
    When IMAP client gets status of "INBOX"
    And IMAP response contains "MESSAGES 2"
    And IMAP response contains "UNSEEN 1"
    And IMAP response contains "UIDNEXT 3"
    And IMAP response contains "UIDVALIDITY"
