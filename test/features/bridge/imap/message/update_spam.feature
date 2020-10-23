Feature: IMAP update messages in Spam folder
  Background:
    Given there is connected user "user"
    # Messages are inserted in opposite way to keep increasing ID.
    # Sequence numbers are then opposite than listed above.
    And there are messages in mailbox "Spam" for "user"
      | from              | to         | subject | body  | read  | starred | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    | false   |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Spam"

  Scenario: Mark message as read only
    When IMAP client marks message "2" with "\Seen"
    Then IMAP response is "OK"
    And message "1" in "Spam" for "user" is marked as read
    And message "1" in "Spam" for "user" is marked as unstarred
    And API mailbox "Spam" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |
