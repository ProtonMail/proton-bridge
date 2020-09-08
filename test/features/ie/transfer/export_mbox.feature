Feature: Export to MBOX files
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/Foo"
    And there are messages in mailbox "INBOX" for "user"
      | from               | to                  | subject | time                |
      | bridgetest@pm.test | test@protonmail.com | hello   | 2020-01-01T12:00:00 |
    And there are messages in mailbox "Folders/Foo" for "user"
      | from            | to                        | subject | time                |
      | foo@example.com | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | bar@example.com | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | bar@example.com | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |

  Scenario: Export all
    When user "user" exports to MBOX files
    Then progress result is "OK"
    # Every message is also in All Mail.
    And transfer exported 8 messages
    And transfer imported 8 messages
    And transfer failed for 0 messages
    And transfer exported messages
      | folder   | from               | to                        | subject | time                |
      | Inbox    | bridgetest@pm.test | test@protonmail.com       | hello   | 2020-01-01T12:00:00 |
      | Foo      | foo@example.com    | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | Foo      | bar@example.com    | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | Foo      | bar@example.com    | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |
      | All Mail | bridgetest@pm.test | test@protonmail.com       | hello   | 2020-01-01T12:00:00 |
      | All Mail | foo@example.com    | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | All Mail | bar@example.com    | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | All Mail | bar@example.com    | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |

  Scenario: Export only Foo with time limit
    When user "user" exports to MBOX files with rules
      | source | target | from                | to                  |
      | Foo    |        | 2020-01-01T12:10:00 | 2020-01-01T13:00:00 |
    Then progress result is "OK"
    And transfer exported 2 messages
    And transfer imported 2 messages
    And transfer failed for 0 messages
    And transfer exported messages
      | folder | from            | to                        | subject | time                |
      | Foo    | bar@example.com | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | Foo    | bar@example.com | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |
