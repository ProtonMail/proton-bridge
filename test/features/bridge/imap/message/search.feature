Feature: IMAP search messages
  Background:
    Given there is connected user "user"
    Given there are messages in mailbox "INBOX" for "user"
      | from               | to         | cc         | subject | read  | starred | deleted | body  |
      | john.doe@email.com | user@pm.me |            | foo     | false | false   | false   | hello |
      | jane.doe@email.com | user@pm.me | name@pm.me | bar     | true  | true    | false   | world |
      | jane.doe@email.com | name@pm.me |            | baz     | true  | false   | true    | bye   |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Search by Sequence numbers
    When IMAP client searches for "1"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1[^0-9]*$"

  Scenario: Search by UID
    When IMAP client searches for "UID 2"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 2[^0-9]*$"

  Scenario: Search by Sequence numbers and UID
    When IMAP client searches for "1 UID 1"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1[^0-9]*$"

  Scenario: Search by Sequence numbers and UID without match
    When IMAP client searches for "1 UID 2"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH[^0-9]*$"

  Scenario: Search by Subject
    When IMAP client searches for "SUBJECT foo"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1[^0-9]*$"

  Scenario: Search by From
    When IMAP client searches for "FROM jane.doe@email.com"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 2 3[^0-9]*$"

  Scenario: Search by To
    When IMAP client searches for "TO user@pm.me"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1 2[^0-9]*$"

  Scenario: Search by CC
    When IMAP client searches for "CC name@pm.me"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 2[^0-9]*$"

  Scenario: Search flagged messages
    When IMAP client searches for "FLAGGED"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 2[^0-9]*$"

  Scenario: Search not flagged messages
    When IMAP client searches for "UNFLAGGED"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1 3[^0-9]*$"

  Scenario: Search seen messages
    When IMAP client searches for "SEEN"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 2 3[^0-9]*$"

  Scenario: Search unseen messages
    When IMAP client searches for "UNSEEN"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1[^0-9]*$"

  Scenario: Search deleted messages
    When IMAP client searches for "DELETED"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 3[^0-9]*$"

  Scenario: Search undeleted messages
    When IMAP client searches for "UNDELETED"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1 2[^0-9]*$"

  Scenario: Search recent messages
    When IMAP client searches for "RECENT"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 1 2 3[^0-9]*$"

  Scenario: Search by more criterias
    When IMAP client searches for "SUBJECT baz TO name@pm.me SEEN UNFLAGGED"
    Then IMAP response is "OK"
    And IMAP response contains "SEARCH 3[^0-9]*$"
