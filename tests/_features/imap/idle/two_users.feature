Feature: IMAP IDLE with two users
  Scenario: IDLE statements are not leaked to other account
    Given there is connected user "user"
    And there are 10 messages in mailbox "INBOX" for "user"
    And there is connected user "userMoreAddresses"
    And there is IMAP client "active" logged in as "user"
    And there is IMAP client "active" selected in "INBOX"
    And there is IMAP client "idling" logged in as "userMoreAddresses"
    And there is IMAP client "idling" selected in "INBOX"
    When IMAP client "idling" starts IDLE-ing
    And IMAP client "active" marks message seq "1" as read
    Then IMAP client "idling" does not receive update for message seq "1" within 5 seconds

  Scenario: IDLE statements are not leaked to other alias
    Given there is connected user "userMoreAddresses"
    And there is "userMoreAddresses" in "combined" address mode
    And there are messages in mailbox "INBOX" for "userMoreAddresses"
      | from              | to          | subject |
      | john.doe@mail.com | [primary]   | foo     |
      | jane.doe@mail.com | [secondary] | bar     |
    And there is IMAP client "active" logged in as "userMoreAddresses" with address "primary"
    And there is IMAP client "active" selected in "INBOX"
    And there is IMAP client "idling" logged in as "userMoreAddresses" with address "secondary"
    And there is IMAP client "idling" selected in "INBOX"
    When IMAP client "idling" starts IDLE-ing
    And IMAP client "active" marks message seq "1" as read
    Then IMAP client "idling" does not receive update for message seq "1" within 5 seconds
