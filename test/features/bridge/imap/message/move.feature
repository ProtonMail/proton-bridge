Feature: IMAP move messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/folder"
    And there is "user" with mailbox "Labels/label"
    And there is "user" with mailbox "Labels/label2"
    And there are messages in mailbox "INBOX" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | foo     | hello |
      | jane.doe@mail.com | name@pm.me | bar     | world |
    And there are messages in mailbox "Sent" for "user"
      | from              | to         | subject  | body  |
      | john.doe@mail.com | user@pm.me | response | hello |
    And there are messages in mailbox "Labels/label2" for "user"
      | from              | to         | subject | body  |
      | john.doe@mail.com | user@pm.me | baz     | hello |
    And there is IMAP client logged in as "user"

  Scenario: Move message from inbox (folder) to folder
    Given there is IMAP client selected in "INBOX"
    When IMAP client moves message seq "1" to "Folders/folder"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "Folders/folder" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is not called

  Scenario: Move all messages from inbox to folder
    Given there is IMAP client selected in "INBOX"
    When IMAP client moves message seq "1:*" to "Folders/folder"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 0 messages
    And mailbox "Folders/folder" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is not called

  Scenario: Move message from folder to label (keeps in folder)
    Given there is IMAP client selected in "INBOX"
    When IMAP client moves message seq "1" to "Labels/label"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "Labels/label" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is not called

  Scenario: Move message from label to folder
    Given there is IMAP client selected in "Labels/label2"
    When IMAP client moves message seq "1" to "Folders/folder"
    Then IMAP response is "OK"
    And mailbox "Labels/label2" for "user" has 0 messages
    And mailbox "Folders/folder" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | baz     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is called

  Scenario: Move message from label to label
    Given there is IMAP client selected in "Labels/label2"
    When IMAP client moves message seq "1" to "Labels/label"
    Then IMAP response is "OK"
    And mailbox "Labels/label2" for "user" has 0 messages
    And mailbox "Labels/label" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | baz     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is called

  Scenario: Move message from All Mail is not possible
    Given there is IMAP client selected in "All Mail"
    When IMAP client moves message seq "1" to "Folders/folder"
    Then IMAP response is "OK"
    And mailbox "All Mail" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | foo     |
      | jane.doe@mail.com | name@pm.me | bar     |
    And mailbox "Folders/folder" for "user" has messages
      | from              | to         | subject |
      | john.doe@mail.com | user@pm.me | baz     |
    And API endpoint "PUT /mail/v4/messages/label" is called
    And API endpoint "PUT /mail/v4/messages/unlabel" is not called

  Scenario: Move message from Inbox to Sent is not possible
    Given there is IMAP client selected in "INBOX"
    When IMAP client moves message seq "1" to "Sent"
    Then IMAP response is "move from Inbox to Sent is not allowed"

  Scenario: Move message from Sent to Inbox is not possible
    Given there is IMAP client selected in "Sent"
    When IMAP client moves message seq "1" to "INBOX"
    Then IMAP response is "move from Sent to Inbox is not allowed"
