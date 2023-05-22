Feature: Bridge can fully sync an account
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
      | two   | folder |
      | three | label  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Folders/two":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |
    Then it succeeds
    When bridge starts
    Then it succeeds

  Scenario: The account is synced when the user logs in and persists across bridge restarts
    When the user logs in with username "[user:user]" and password "password"
    Then bridge sends sync started and finished events for user "[user:user]"
    When bridge restarts
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following mailbox info:
      | name         | total | unread |
      | INBOX        | 0     | 0      |
      | Drafts       | 0     | 0      |
      | Sent         | 0     | 0      |
      | Starred      | 0     | 0      |
      | Archive      | 0     | 0      |
      | Spam         | 0     | 0      |
      | Trash        | 0     | 0      |
      | All Mail     | 4     | 2      |
      | Folders      | 0     | 0      |
      | Folders/one  | 2     | 1      |
      | Folders/two  | 2     | 1      |
      | Labels       | 0     | 0      |
      | Labels/three | 0     | 0      |

  Scenario: If the gluon files are deleted, the account is synced again
    Given the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And bridge stops
    And the user deletes the gluon files
    And bridge starts
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following mailbox info:
      | name         | total | unread |
      | INBOX        | 0     | 0      |
      | Drafts       | 0     | 0      |
      | Sent         | 0     | 0      |
      | Starred      | 0     | 0      |
      | Archive      | 0     | 0      |
      | Spam         | 0     | 0      |
      | Trash        | 0     | 0      |
      | All Mail     | 4     | 2      |
      | Folders      | 0     | 0      |
      | Folders/one  | 2     | 1      |
      | Folders/two  | 2     | 1      |
      | Labels       | 0     | 0      |
      | Labels/three | 0     | 0      |

  Scenario: If an address has no keys, it does not break other addresses
    Given the account "[user:user]" has additional address "[alias:alias]@[domain]" without keys
    When the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then IMAP client "1" eventually sees the following messages in "Folders/one":
      | from       | to         | subject | unread |
      | a@[domain] | a@[domain] | one     | true   |
      | b@[domain] | b@[domain] | two     | false  |