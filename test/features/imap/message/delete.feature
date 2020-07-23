Feature: IMAP delete messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"

  Scenario Outline: Delete message
    Given there are 10 messages in mailbox "<mailbox>" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client deletes messages "1"
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 9 messages

    Examples:
      | mailbox      |
      | INBOX        |
      | Folders/mbox |
      | Labels/label |
      | Trash        |

  Scenario Outline: Delete all messages
    Given there are 10 messages in mailbox "<mailbox>" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "<mailbox>"
    When IMAP client deletes messages "1:*"
    Then IMAP response is "OK"
    And mailbox "<mailbox>" for "user" has 0 messages

    Examples:
      | mailbox      |
      | INBOX        |
      | Folders/mbox |
      | Labels/label |
      | Trash        |

  Scenario: Delete message by setting flags
    Given there are 1 messages in mailbox "INBOX" for "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "INBOX"
    When IMAP client marks message "1" with "\Deleted"
    Then IMAP response is "OK"
    And mailbox "INBOX" for "user" has 0 messages
    # Unread because we set flags without \Seen.
    And message "1" in "Trash" for "user" is marked as unread
