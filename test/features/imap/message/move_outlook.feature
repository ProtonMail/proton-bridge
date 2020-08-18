Feature: IMAP move messages like Outlook
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"

  Scenario: Move message from INBOX to mailbox like outlook
    When IMAP client moves messages "2" to "<mailbox>" like Outlook for "user"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "<mailbox>" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |

    Examples:
      | mailbox      |
      | Archive      |
      | Folders/mbox |
      | Spam         |
      | Trash        |



  #Scenario: Move message from Trash/Spam to INBOX like outlook
