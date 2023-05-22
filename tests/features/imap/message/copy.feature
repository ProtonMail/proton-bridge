Feature: IMAP copy messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
      | label | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Copy message to label
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Labels/label"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" eventually sees the following messages in "Labels/label":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |

  Scenario: Copy all messages to label
    When IMAP client "1" copies all messages from "INBOX" to "Labels/label"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" eventually sees the following messages in "Labels/label":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |

  Scenario: Copy message to folder does move
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Folders/mbox"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to            | subject | unread |
      | jane.doe@mail.com | name@[domain] | bar     | true   |
    And IMAP client "1" eventually sees the following messages in "Folders/mbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |

  Scenario: Copy all messages to folder does move
    When IMAP client "1" copies all messages from "INBOX" to "Folders/mbox"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "Folders/mbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" eventually sees 0 messages in "INBOX"

  Scenario: Copy message from Inbox to Sent is not possible
    When IMAP client "1" copies the message with subject "foo" from "INBOX" to "Sent"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" eventually sees 0 messages in "Sent"

  Scenario: Copy message from All mail moves from the original location
    Given IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    When IMAP client "1" copies the message with subject "foo" from "All Mail" to "Folders/mbox"
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to            | subject | unread |
      | jane.doe@mail.com | name@[domain] | bar     | true   |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" eventually sees the following messages in "Folders/mbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |

