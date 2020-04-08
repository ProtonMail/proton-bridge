Feature: IMAP update messages
  Background:
    Given there is connected user "user"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | starred |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Mark message as read
    When IMAP client marks message "2" as read
    Then IMAP response is "OK"
    And message "1" in "INBOX" for "user" is marked as read
    And message "1" in "INBOX" for "user" is marked as unstarred

  Scenario: Mark message as unread
    When IMAP client marks message "1" as unread
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as unread
    And message "2" in "INBOX" for "user" is marked as starred

  Scenario: Mark message as starred
    Then message "1" in "INBOX" for "user" is marked as unread
    And message "1" in "INBOX" for "user" is marked as unstarred
    When IMAP client marks message "2" as starred
    Then IMAP response is "OK"
    And message "1" in "INBOX" for "user" is marked as unread
    And message "1" in "INBOX" for "user" is marked as starred

  Scenario: Mark message as unstarred
    When IMAP client marks message "1" as unstarred
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as read
    And message "2" in "INBOX" for "user" is marked as unstarred
