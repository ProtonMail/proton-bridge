Feature: IMAP move messages by append and delete (without MOVE support, e.g., Outlook)
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is IMAP client "source" logged in as "user"
    And there is IMAP client "target" logged in as "user"

  Scenario Outline: Move message from INBOX to mailbox by append and delete
    Given there are messages in mailbox "INBOX" for "user"
      | id | from              | to         | subject | body  |
      | 1  | john.doe@mail.com | user@pm.me | foo     | hello |
      | 2  | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client "source" selected in "INBOX"
    And there is IMAP client "target" selected in "<mailbox>"
    When IMAP clients "source" and "target" move message seq "2" of "user" from "INBOX" to "<mailbox>" by append and delete
    Then IMAP response to "source" is "OK"
    Then IMAP response to "target" is "OK"
    When IMAP client "source" sends expunge
    Then IMAP response to "source" is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
    And mailbox "<mailbox>" for "user" has messages
      | from              | to         | subject |
      | jane.doe@mail.com | name@pm.me | bar     |

    Examples:
      | mailbox      |
      | Archive      |
      | Folders/mbox |
      | Spam         |
      | Trash        |

  Scenario Outline: Move message from Trash/Spam to INBOX by append and delete
    Given there are messages in mailbox "<mailbox>" for "user"
      | id | from              | to         | subject | body  |
      | 1  | john.doe@mail.com | user@pm.me | foo     | hello |
      | 2  | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client "source" selected in "<mailbox>"
    And there is IMAP client "target" selected in "INBOX"
    When IMAP clients "source" and "target" move message seq "2" of "user" from "<mailbox>" to "INBOX" by append and delete
    Then IMAP response to "source" is "OK"
    Then IMAP response to "target" is "OK"
    When IMAP client "source" sends expunge
    Then IMAP response to "source" is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "<mailbox>" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |

    Examples:
      | mailbox |
      | Spam    |
      | Trash   |
