Feature: Import from MBOX files
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/Foo"
    And there is "user" with mailbox "Folders/Bar"
    And there is MBOX file "Foo.mbox" with messages
      | from            | to                        | subject | time                |
      | foo@example.com | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | bar@example.com | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
    And there is MBOX file "Sub/Foo.mbox" with messages
      | from            | to                        | subject | time                |
      | bar@example.com | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |
    And there is MBOX file "Inbox.mbox"
      """
      From bridgetest@pm.test Thu Feb 20 20:20:20 2020
      Subject: hello
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>

      hello

      """

  Scenario: Import all
    When user "user" imports local files
    Then progress result is "OK"
    And transfer exported 4 messages
    And transfer imported 4 messages
    And transfer failed for 0 messages
    And API mailbox "INBOX" for "user" has messages
      | from               | to                  | subject |
      | bridgetest@pm.test | test@protonmail.com | hello   |
    And API mailbox "Folders/Foo" for "user" has messages
      | from            | to                        | subject |
      | foo@example.com | bridgetest@protonmail.com | one     |
      | bar@example.com | bridgetest@protonmail.com | two     |
      | bar@example.com | bridgetest@protonmail.com | three   |

  Scenario: Import only Foo to Bar with time limit
    When user "user" imports local files with rules
      | source | target | from                | to                  |
      | Foo    | Bar    | 2020-01-01T12:10:00 | 2020-01-01T13:00:00 |
    Then progress result is "OK"
    And transfer exported 2 messages
    And transfer imported 2 messages
    And transfer failed for 0 messages
    And API mailbox "Folders/Bar" for "user" has messages
      | from            | to                        | subject |
      | bar@example.com | bridgetest@protonmail.com | two     |
      | bar@example.com | bridgetest@protonmail.com | three   |

  Scenario: Import broken message
    Given there is MBOX file "Broken.mbox"
      """
      From bridgetest@pm.test Thu Feb 20 20:20:20 2020
      Content-type: image/png
      """
    When user "user" imports local files with rules
      | source | target |
      | Broken | Foo    |
    Then progress result is "OK"
    And transfer exported 1 messages
    And transfer imported 0 messages
    And transfer failed for 1 messages
