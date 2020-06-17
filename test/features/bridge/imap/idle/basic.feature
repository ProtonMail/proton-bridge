Feature: IMAP IDLE
  Background:
    Given there is connected user "user"
    And there are 10 messages in mailbox "INBOX" for "user"

  @ignore
  Scenario Outline: Mark as read
    Given there is IMAP client "active" logged in as "user"
    And there is IMAP client "active" selected in "INBOX"
    And there is IMAP client "idling" logged in as "user"
    And there is IMAP client "idling" selected in "INBOX"
    When IMAP client "idling" starts IDLE-ing
    And IMAP client "active" marks message "<message>" as read
    Then IMAP client "idling" receives update marking message "<message>" as read within <seconds> seconds
    Then message "<message>" in "INBOX" for "user" is marked as read

    Examples:
      | message | seconds |
      | 1       | 2       |
      | 1:5     | 2       |
      | 1:10    | 5       |

  @ignore
  Scenario Outline: Mark as unread
    Given there is IMAP client "active" logged in as "user"
    And there is IMAP client "active" selected in "INBOX"
    And there is IMAP client "idling" logged in as "user"
    And there is IMAP client "idling" selected in "INBOX"
    When IMAP client "idling" starts IDLE-ing
    And IMAP client "active" marks message "<message>" as unread
    Then IMAP client "idling" receives update marking message "<message>" as unread within <seconds> seconds
    And message "<message>" in "INBOX" for "user" is marked as unread

    Examples:
      | message | seconds |
      | 1       | 2       |
      | 1:5     | 2       |
      | 1:10    | 5       |

  @ignore
  Scenario Outline: Three IDLEing
    Given there is IMAP client "active" logged in as "user"
    And there is IMAP client "active" selected in "INBOX"
    And there is IMAP client "idling1" logged in as "user"
    And there is IMAP client "idling1" selected in "INBOX"
    And there is IMAP client "idling2" logged in as "user"
    And there is IMAP client "idling2" selected in "INBOX"
    And there is IMAP client "idling3" logged in as "user"
    And there is IMAP client "idling3" selected in "INBOX"
    When IMAP client "idling1" starts IDLE-ing
    And IMAP client "idling2" starts IDLE-ing
    And IMAP client "idling3" starts IDLE-ing
    And IMAP client "active" marks message "<message>" as read
    Then IMAP client "idling1" receives update marking message "<message>" as read within <seconds> seconds
    Then IMAP client "idling2" receives update marking message "<message>" as read within <seconds> seconds
    Then IMAP client "idling3" receives update marking message "<message>" as read within <seconds> seconds

    Examples:
      | message | seconds |
      | 1       | 2       |
      | 1:5     | 2       |
      | 1:10    | 5       |
