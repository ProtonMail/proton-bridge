Feature: IMAP copy messages
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |
    And the account "user@pm.me" has the following messages in "Inbox":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Copy message to label
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Labels/label"
    Then IMAP client "1" sees the following messages in "INBOX":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And IMAP client "1" sees the following messages in "Labels/label":
      | sender            | recipient    | subject  | unread |
      | john.doe@mail.com | user@pm.me   | foo      | false  |

  Scenario: Copy all messages to label
    When IMAP client "1" copies all messages from "INBOX" to "Labels/label"
    Then IMAP client "1" sees the following messages in "INBOX":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And IMAP client "1" sees the following messages in "Labels/label":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |

  Scenario: Copy message to folder does move
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Folders/mbox"
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | sender            | recipient    | subject | unread |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And IMAP client "1" sees the following messages in "Folders/mbox":
      | sender            | recipient    | subject  | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |

  Scenario: Copy all messages to folder does move
    When IMAP client "1" copies all messages from "INBOX" to "Folders/mbox"
    Then IMAP client "1" sees the following messages in "Folders/mbox":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And IMAP client "1" eventually sees 0 messages in "INBOX"

  Scenario: Copy message from Inbox to Sent is not possible
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Sent"
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | sender            | recipient    | subject | unread |
      | john.doe@mail.com | user@pm.me   | foo     | false  |
      | jane.doe@mail.com | name@pm.me   | bar     | true   |
    And IMAP client "1" eventually sees 0 messages in "Sent"