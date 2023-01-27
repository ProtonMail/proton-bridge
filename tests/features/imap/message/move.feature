Feature: IMAP move messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name   | type   |
      | mbox   | folder |
      | label  | label  |
      | label2 | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Labels/label2":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | baz     | false  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Sent":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | bax     | false  |
    And bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"

  Scenario: Move message from folder to label (keeps in folder)
    When IMAP client "1" moves the message with subject "foo" from "INBOX" to "Labels/label"
    And it succeeds
    And IMAP client "1" sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "1" sees the following messages in "Labels/label":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |

  # This test covers a limitation of Gluon where we are not able to ensure the messages moved to a label via append
  # expunge are preserved in the original folder.
  Scenario: Move message from folder to label with append expunge does not keep message in origin folder
    When user "[user:user]" connects and authenticates IMAP client "source"
    And user "[user:user]" connects and authenticates IMAP client "target"
    And IMAP client "source" selects "INBOX"
    And IMAP client "target" selects "Labels/label"
    And IMAP clients "source" and "target" move message with subject "foo" of "[user:user]" to "Labels/label" by APPEND DELETE EXPUNGE
    And it succeeds
    Then IMAP client "source" sees the following messages in "INBOX":
      | from              | to                   | subject | unread |
      | jane.doe@mail.com | name@[domain]        | bar     | true   |
    And IMAP client "target" sees the following messages in "Labels/label":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | foo     | false  |

  Scenario: Move message from label to folder
    When IMAP client "1" moves the message with subject "baz" from "Labels/label2" to "Folders/mbox"
    And it succeeds
    And IMAP client "1" sees the following messages in "Folders/mbox":
      | from              | to                   | subject | unread |
      | john.doe@mail.com | [user:user]@[domain] | baz     | false  |
    And IMAP client "1" sees 0 messages in "Labels/label2"

   Scenario: Move message from label to label
     When IMAP client "1" moves the message with subject "baz" from "Labels/label2" to "Labels/label"
     And it succeeds
     And IMAP client "1" sees the following messages in "Labels/label":
       | from              | to                   | subject | unread |
       | john.doe@mail.com | [user:user]@[domain] | baz     | false  |
     And IMAP client "1" sees 0 messages in "Labels/label2"

   Scenario: Move message from All Mail is not possible
     When IMAP client "1" moves the message with subject "baz" from "All Mail" to "Folders/folder"
     Then it fails
     And IMAP client "1" sees the following messages in "All Mail":
       | from              | to                   | subject | unread |
       | john.doe@mail.com | [user:user]@[domain] | foo     | false  |
       | jane.doe@mail.com | name@[domain]        | bar     | true   |
       | john.doe@mail.com | [user:user]@[domain] | baz     | false  |
       | john.doe@mail.com | [user:user]@[domain] | bax     | false  |

   Scenario: Move message from Inbox to Sent is not possible
     Given test skips reporter checks
     When IMAP client "1" moves the message with subject "bar" from "Inbox" to "Sent"
     Then it fails

   Scenario: Move message from Sent to Inbox is not possible
     Given test skips reporter checks
     When IMAP client "1" moves the message with subject "bax" from "Sent" to "Inbox"
     Then it fails