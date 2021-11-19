Feature: IMAP move messages by append and delete (without MOVE support, e.g., Outlook)
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is IMAP client "source" logged in as "user"
    And there is IMAP client "target" logged in as "user"

  Scenario Outline: Move message from <srcMailbox> to <dstMailbox> by <order>
    Given there are messages in mailbox "<srcMailbox>" for "user"
      | id | from        | to          | subject | body  |
      | 1  | sndr1@pm.me | rcvr1@pm.me | subj1   | body1 |
      | 2  | sndr2@pm.me | rcvr2@pm.me | subj2   | body2 |
    And there is IMAP client "source" selected in "<srcMailbox>"
    And there is IMAP client "target" selected in "<dstMailbox>"
    When IMAP clients "source" and "target" move message seq "2" of "user" to "<dstMailbox>" by <order>
    Then IMAP response to "source" is "OK"
    Then IMAP response to "target" is "OK"
    And mailbox "<dstMailbox>" for "user" has 1 messages
    And mailbox "<dstMailbox>" for "user" has messages
      | from        | to          | subject |
      | sndr2@pm.me | rcvr2@pm.me | subj2   |
    And mailbox "<srcMailbox>" for "user" has 1 messages
    And mailbox "<srcMailbox>" for "user" has messages
      | from        | to          | subject |
      | sndr1@pm.me | rcvr1@pm.me | subj1   |
    Examples:
      | srcMailbox | dstMailbox   | order                 |
      | Trash      | INBOX        | APPEND DELETE EXPUNGE |
      | Spam       | INBOX        | APPEND DELETE EXPUNGE |
      | INBOX      | Archive      | APPEND DELETE EXPUNGE |
      | INBOX      | Folders/mbox | APPEND DELETE EXPUNGE |
      | INBOX      | Spam         | APPEND DELETE EXPUNGE |
      | INBOX      | Trash        | APPEND DELETE EXPUNGE |
      | Trash      | INBOX        | DELETE APPEND EXPUNGE |
      | Spam       | INBOX        | DELETE APPEND EXPUNGE |
      | INBOX      | Archive      | DELETE APPEND EXPUNGE |
      | INBOX      | Folders/mbox | DELETE APPEND EXPUNGE |
      | INBOX      | Spam         | DELETE APPEND EXPUNGE |
      | INBOX      | Trash        | DELETE APPEND EXPUNGE |
      | Trash      | INBOX        | DELETE EXPUNGE APPEND |
      | Spam       | INBOX        | DELETE EXPUNGE APPEND |
      | INBOX      | Archive      | DELETE EXPUNGE APPEND |
      | INBOX      | Folders/mbox | DELETE EXPUNGE APPEND |
      | INBOX      | Spam         | DELETE EXPUNGE APPEND |
      | INBOX      | Trash        | DELETE EXPUNGE APPEND |

  Scenario Outline: Move message from <mailbox> to All Mail by <order>
    Given there are messages in mailbox "<mailbox>" for "user"
      | id | from              | to         | subject | body  |
      | 1  | john.doe@mail.com | user@pm.me | subj1   | body1 |
      | 2  | john.doe@mail.com | name@pm.me | subj2   | body2 |
    And there is IMAP client "source" selected in "<mailbox>"
    And there is IMAP client "target" selected in "All Mail"
    When IMAP clients "source" and "target" move message seq "2" of "user" to "All Mail" by <order>
    Then IMAP response to "source" is "OK"
    Then IMAP response to "target" is "OK"
    And mailbox "<mailbox>" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | subj1   |
    And mailbox "All Mail" for "user" has messages
       | from              | to         | subject |
       | john.doe@mail.com | user@pm.me | subj1   |
       | john.doe@mail.com | name@pm.me | subj2   |
    Examples:
      | mailbox | order                 |
      | INBOX   | APPEND DELETE EXPUNGE |
      | Archive | APPEND DELETE EXPUNGE |
      | Trash   | APPEND DELETE EXPUNGE |
      | Spam    | APPEND DELETE EXPUNGE |
      | INBOX   | DELETE APPEND EXPUNGE |
      | Archive | DELETE APPEND EXPUNGE |
      | Trash   | DELETE APPEND EXPUNGE |
      | Spam    | DELETE APPEND EXPUNGE |
      | INBOX   | DELETE EXPUNGE APPEND |
      | Archive | DELETE EXPUNGE APPEND |
      | Trash   | DELETE EXPUNGE APPEND |
      | Spam    | DELETE EXPUNGE APPEND |
