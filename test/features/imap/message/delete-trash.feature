Feature: IMAP remove messages from Trash
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"

  Scenario Outline: Delete messages from Trash/Spam does not remove from All Mail
    Given there are messages in mailbox "<mailbox>" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client copies messages "2" to "Labels/label"
    Then IMAP response is "OK"
    When IMAP client marks message "2" as deleted
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 2 messages
    And mailbox "All Mail" for "user" has 2 messages
    And mailbox "Labels/label" for "user" has 1 messages
    When IMAP client sends expunge
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 1 messages
    And mailbox "All Mail" for "user" has 2 messages
    And mailbox "Labels/label" for "user" has 1 messages

    Examples:
      | mailbox |
      | Spam    |
      | Trash   |


  Scenario Outline: Delete messages from Trash/Spamm removes from All Mail
    Given there are messages in mailbox "<mailbox>" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client marks message "2" as deleted
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 2 messages
    And mailbox "All Mail" for "user" has 2 messages
    When IMAP client sends expunge
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 1 messages
    And mailbox "All Mail" for "user" has 1 messages

    Examples:
      | mailbox |
      | Spam    |
      | Trash   |
