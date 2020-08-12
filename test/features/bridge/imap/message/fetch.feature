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

  Scenario: Fetch first few message of inbox
    Given there are 10 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client fetches "1:5"
    Then IMAP response is "OK"
    And IMAP response has 5 messages

  Scenario: Fetch of custom mailbox
    Given there are 10 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 10 messages

  Scenario: Fetch of emtpy mailbox
    Given there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 0 messages

  Scenario: Fetch of big mailbox
    Given there are 100 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches "1:*"
    Then IMAP response is "OK"
    And IMAP response has 100 messages

  Scenario: Fetch returns alsways latest messages
    Given there are 10 messages in mailbox "Folders/mbox" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches by UID "11:*"
    Then IMAP response is "OK"
    And IMAP response has 1 message

  Scenario: Fetch returns also messages that are marked as deleted
    Given there are messages in mailbox "Folders/mbox" for "user"
      | from              | to         | subject | body  | read  | starred | deleted |
      | john.doe@mail.com | user@pm.me | foo     | hello | false | false   | false   |
      | jane.doe@mail.com | name@pm.me | bar     | world | true  | true    | true    |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Folders/mbox"
    When IMAP client fetches by UID "1:*"
    Then IMAP response is "OK"
    And IMAP response has 2 message
