Feature: IMAP move messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Move message
    When IMAP client moves message seq "1" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |

  Scenario: Move all messages
    When IMAP client moves message seq "1:*" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 0 messages
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |

  Scenario: Move message from All Mail is not possible
    When IMAP client moves message seq "1" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "All Mail" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
