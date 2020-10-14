Feature: IMAP get mailbox info
  Background:
    Given there is connected user "user"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | starred |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    |
    And there is IMAP client logged in as "user"

  Scenario: Mailbox info contains mailbox name
    When IMAP client gets info of "INBOX"
    Then IMAP response contains "2 EXISTS"
    And IMAP response contains "UNSEEN 1"
    And IMAP response contains "UIDNEXT 3"
    And IMAP response contains "UIDVALIDITY"
