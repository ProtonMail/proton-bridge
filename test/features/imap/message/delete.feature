Feature: IMAP delete messages
  Background:
    Given there is connected user "user"
    And there is "user" with mailbox "Folders/mbox"
    And there is "user" with mailbox "Labels/label"

  # https://gitlab.protontech.ch/ProtonMail/Slim-API/issues/1420
  @ignore-live
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
      | Drafts       |
      | Trash        |

  # https://gitlab.protontech.ch/ProtonMail/Slim-API/issues/1420
  @ignore-live
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
      | Drafts       |
      | Trash        |
