Feature: IMAP remove messages from Trash
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |

  Scenario Outline: Message in Trash or Spam and some other label is not permanently deleted
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "<mailbox>":
      | from              | to           | subject | body  |
      | john.doe@mail.com | user@pm.me   | foo     | hello |
      | jane.doe@mail.com | name@pm.me   | bar     | world |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "<mailbox>"
    When IMAP client "1" copies the message with subject "foo" from "<mailbox>" to "Labels/label"
    Then it succeeds
    When IMAP client "1" marks the message with subject "foo" as deleted
    Then it succeeds
    And IMAP client "1" sees 2 messages in "<mailbox>"
    And IMAP client "1" sees 2 messages in "All Mail"
    And IMAP client "1" sees 1 messages in "Labels/label"
    When IMAP client "1" expunges
    Then it succeeds
    And IMAP client "1" sees 1 messages in "<mailbox>"
    And IMAP client "1" sees 2 messages in "All Mail"
    And IMAP client "1" sees 1 messages in "Labels/label"

    Examples:
      | mailbox |
      | Spam    |
      | Trash   |

  Scenario Outline: Message in Trash or Spam only is permanently deleted
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "<mailbox>":
      | from              | to           | subject | body  |
      | john.doe@mail.com | user@pm.me   | foo     | hello |
      | jane.doe@mail.com | name@pm.me   | bar     | world |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "<mailbox>"
    When IMAP client "1" marks the message with subject "foo" as deleted
    Then it succeeds
    And IMAP client "1" sees 2 messages in "<mailbox>"
    And IMAP client "1" sees 2 messages in "All Mail"
    When IMAP client "1" expunges
    Then it succeeds
    And IMAP client "1" sees 1 messages in "<mailbox>"
    And IMAP client "1" eventually sees 1 messages in "All Mail"

    Examples:
      | mailbox |
      | Spam    |
      | Trash   |
