Feature: Import-Export app
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/Foo"
    And there is "user" with mailbox "Folders/Bar"

  Scenario: EML -> PM -> EML
    Given there are EML files
      | file              | from               | to                        | subject | time                |
      | Inbox/hello.eml   | bridgetest@pm.test | test@protonmail.com       | hello   | 2020-01-01T12:00:00 |
      | Foo/one.eml       | foo@example.com    | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | Foo/two.eml       | bar@example.com    | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | Sub/Foo/three.eml | bar@example.com    | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |

    When user "user" imports local files
    Then progress result is "OK"
    And transfer failed for 0 messages
    And transfer imported 4 messages

    When user "user" exports to EML files
    Then progress result is "OK"
    And transfer failed for 0 messages
    # Every message is also in All Mail.
    And transfer imported 8 messages

    And exported messages match the original ones

  Scenario: MBOX -> PM -> MBOX
    Given there is MBOX file "Inbox.mbox" with messages
      | from               | to                  | subject | time                |
      | bridgetest@pm.test | test@protonmail.com | hello   | 2020-01-01T12:00:00 |
    And there is MBOX file "Foo.mbox" with messages
      | from            | to                        | subject | time                |
      | foo@example.com | bridgetest@protonmail.com | one     | 2020-01-01T12:00:00 |
      | bar@example.com | bridgetest@protonmail.com | two     | 2020-01-01T13:00:00 |
      | bar@example.com | bridgetest@protonmail.com | three   | 2020-01-01T12:30:00 |

    When user "user" imports local files
    Then progress result is "OK"
    And transfer failed for 0 messages
    And transfer imported 4 messages

    When user "user" exports to MBOX files
    Then progress result is "OK"
    And transfer failed for 0 messages
    # Every message is also in All Mail.
    And transfer imported 8 messages

    And exported messages match the original ones
