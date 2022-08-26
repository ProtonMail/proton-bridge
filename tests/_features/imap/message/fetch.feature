Feature: IMAP fetch messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"

  Scenario: Fetch of inbox
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  Scenario: Fetch of inbox by UID
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "1:*"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  Scenario: Fetch first few messages of inbox
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "1:5"
    Then IMAP response is "OK"
    And IMAP response has 5 messages

  Scenario: Fetch first few messages of inbox by UID
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "1:5"
    Then IMAP response is "OK"
    And IMAP response has 5 messages

  Scenario: Fetch last few messages of inbox using wildcard
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "6:*"
    Then IMAP response is "OK"
    And IMAP response has 5 messages

  Scenario: Fetch last few messages of inbox using wildcard by UID
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "6:*"
    Then IMAP response is "OK"
    And IMAP response has 5 messages

  Scenario: Fetch last message of inbox using wildcard
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "*"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Fetch last message of inbox using wildcard by UID
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "*"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Fetch backwards range using wildcard
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "*:1"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  Scenario: Fetch backwards range using wildcard by UID
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "*:1"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  Scenario: Fetch overshot range using wildcard returns last message
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "20:*"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Fetch overshot range using wildcard by UID returns last message
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches by UID "20:*"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Fetch of custom mailbox
    Given there are 10 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  # This test is wrong! RFC says it should return "BAD" (GODT-1153).
  Scenario Outline: Fetch range of empty mailbox
    Given there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "<range>"
    Then IMAP response is "OK"
    And IMAP response has 0 messages
    When IMAP client fetches by UID "<range>"
    Then IMAP response is "OK"
    And IMAP response has 0 messages

    Examples:
      | range |
      | 1     |
      | 1,5,6 |
      | 1:*   |
      | *     |

  @ignore-live
  Scenario: Fetch of big mailbox
    Given there are 100 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 100 messages

  Scenario: Fetch of big mailbox by UID
    Given there are 100 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches by UID "1:*"
    Then IMAP response is "OK"
    And IMAP response has 100 messages

  Scenario: Fetch returns also messages that are marked as deleted
    Given there are messages in mailbox "Folders/mbox" for "user"
      | from              | to         | subject | body  | read  | starred | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 2 message

  Scenario: Fetch by UID returns also messages that are marked as deleted
    Given there are messages in mailbox "Folders/mbox" for "user"
      | from              | to         | subject | body  | read  | starred | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches by UID "1:*"
    Then IMAP response is "OK"
    And IMAP response has 2 message

  Scenario: Fetch of very old message sent from the moon succeeds with modified date
    Given there are messages in mailbox "Folders/mbox" for "user"
      | from              | to         | subject        | time                |
      | john.doe@mail.com | user@pm.me | Very old email | 1969-07-20T00:00:00 |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client sends command "FETCH 1:* rfc822"
    Then IMAP response is "OK"
    And IMAP response contains "\nDate: Fri, 13 Aug 1982"
    And IMAP response contains "\nX-Pm-Date: Thu, 01 Jan 1970"
    And IMAP response contains "\nX-Original-Date: Sun, 20 Jul 1969"
    # We had bug to incorectly set empty date, so let's make sure
    # there is no reference anywhere in the response.
    And IMAP response does not contain "\nDate: Thu, 01 Jan 1970"

  Scenario: Fetch of message which was deleted without event processed
    Given there are 10 messages in mailbox "INBOX" for "user"
    And message "5" was deleted forever without event processed for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches bodies "1:*"
    Then IMAP response is "NO"
    When IMAP client fetches bodies "1:*"
    Then IMAP response is "OK"
    And IMAP response has 9 messages
