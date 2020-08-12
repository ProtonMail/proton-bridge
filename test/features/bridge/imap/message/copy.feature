Feature: IMAP copy messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Copy message to label
    When IMAP client copies messages "2" to "Labels/label"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And mailbox "Labels/label" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |

  Scenario: Copy all messages to label
    When IMAP client copies messages "1:*" to "Labels/label"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And mailbox "Labels/label" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |

  Scenario: Copy message to folder does move
    When IMAP client copies messages "2" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |

  Scenario: Copy all messages to folder does move
    When IMAP client copies messages "1:*" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 0 messages
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
