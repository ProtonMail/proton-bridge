Feature: IMAP update messages
  Background:
    Given there is connected user "user"
    # Messages are inserted in opposite way to keep increasing ID.
    # Sequence numbers are then opposite than listed above.
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | starred | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    | false   |
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

  Scenario: Mark message as read and starred
    When IMAP client marks message "2" with "\Seen \Flagged"
    Then IMAP response is "OK"
    And message "1" in "INBOX" for "user" is marked as read
    And message "1" in "INBOX" for "user" is marked as starred

  Scenario: Mark message as read only
    When IMAP client marks message "1" with "\Seen"
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as read
    # Unstarred because we set flags without \Starred.
    And message "2" in "INBOX" for "user" is marked as unstarred

  Scenario: Mark message as spam only
    When IMAP client marks message "1" with "Junk"
    Then IMAP response is "OK"
    # Unread and unstarred because we set flags without \Seen and \Starred.
    And message "1" in "Spam" for "user" is marked as unread
    And message "1" in "Spam" for "user" is marked as unstarred

  Scenario: Mark message as deleted
    When IMAP client marks message "2" as deleted
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as read
    And message "2" in "INBOX" for "user" is marked as starred
    And message "2" in "INBOX" for "user" is marked as deleted

  Scenario: Mark message as undeleted
    When IMAP client marks message "2" as undeleted
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as read
    And message "2" in "INBOX" for "user" is marked as starred
    And message "2" in "INBOX" for "user" is marked as undeleted

  Scenario: Mark message as deleted only
    When IMAP client marks message "2" with "\Deleted"
    Then IMAP response is "OK"
    And message "2" in "INBOX" for "user" is marked as unread
    And message "2" in "INBOX" for "user" is marked as unstarred
    And message "2" in "INBOX" for "user" is marked as undeleted

