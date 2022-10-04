Feature: Bridge can fully sync an account
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has the following custom mailboxes:
      | name  | type   |
      | one   | folder |
      | two   | folder |
      | three | label  |
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "one":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And the address "user@pm.me" of account "user@pm.me" has the following messages in "two":
      | from    | to        | subject | unread |
      | a@pm.me | a@pm.me   | one     | true  |
      | b@pm.me | b@pm.me   | two     | false |
    And bridge starts
    
  Scenario: The account is synced when the user logs in and persists across bridge restarts
    When the user logs in with username "user@pm.me" and password "password"
    Then bridge sends sync started and finished events for user "user@pm.me"
    When bridge restarts
    And user "user@pm.me" connects and authenticates IMAP client "1"
    Then IMAP client "1" sees the following mailbox info:
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
    Given the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And bridge stops
    And the user deletes the gluon files
    And bridge starts
    When user "user@pm.me" connects and authenticates IMAP client "1"
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