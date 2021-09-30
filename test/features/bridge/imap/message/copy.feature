Feature: IMAP copy messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And there are messages in mailbox "Sent" for "user"
      | from              | to         | subject  | body  |
      | john.doe@mail.com | user@pm.me | response | hello |
    And there is IMAP client logged in as "user"

  Scenario: Copy message to label
    Given there is IMAP client selected in "INBOX"
    When IMAP client copies message seq "1" to "Labels/label"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And mailbox "Labels/label" for "user" has messages
      | from              | to         | subject | body  | read | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true | false   |

  Scenario: Copy all messages to label
    Given there is IMAP client selected in "INBOX"
    When IMAP client copies message seq "1:*" to "Labels/label"
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
    Given there is IMAP client selected in "INBOX"
    When IMAP client copies message seq "1" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject | body  | read | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true | false   |

  Scenario: Copy message from All mail creates a duplicate
    Given there is IMAP client selected in "All Mail"
    When IMAP client copies message seq "1" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 2 messages
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject  | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo      | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar      | world | false | true    |
    And mailbox "All Mail" for "user" has 3 messages
    And mailbox "All Mail" for "user" has messages
      | from              | to         | subject  | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo      | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar      | world | false | false   |
    And mailbox "Folders/mbox" for "user" has 1 messages
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject | body  | read | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true | false   |

  Scenario: Copy all messages to folder does move
    Given there is IMAP client selected in "INBOX"
    When IMAP client copies message seq "1:*" to "Folders/mbox"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 0 messages
    And mailbox "Folders/mbox" for "user" has messages
      | from              | to         | subject | body  | read  | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | true  | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | false | true    |

  Scenario: Copy message from Inbox to Sent is not possible
    Given there is IMAP client selected in "INBOX"
    When IMAP client copies message seq "1" to "Sent"
    Then IMAP response is "move from Inbox to Sent is not allowed"

  Scenario: Copy message from Sent to Inbox is not possible
    Given there is IMAP client selected in "Sent"
    When IMAP client copies message seq "1" to "INBOX"
    Then IMAP response is "move from Sent to Inbox is not allowed"
