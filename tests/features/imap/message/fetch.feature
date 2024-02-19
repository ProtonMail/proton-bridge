Feature: IMAP Fetch
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[user:alias]@[domain]"
    And the account "[user:user]" has the following custom mailboxes:
      | name  | type   |
      | mbox  | folder |
    And the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Inbox":
      | from              | to                   | subject | date                          |
      | john.doe@mail.com | [user:user]@[domain] | foo     |  13 Jul 69 00:00 +0000        |
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  # The date returned from black is server time.. Black is probably correct we need to fix GPA server
  @skip-black
  Scenario: Fetch very old message
    Given IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | date                  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | 13 Aug 82 00:00 +0000 |
    Then IMAP client "1" sees header "X-Original-Date: Sun, 13 Jul 1969 00:00:00 +0000" in message with subject "foo" in "INBOX"


  # The date returned from black is server time.. Black is probably correct we need to fix GPA server
  @skip-black
  Scenario: Fetch from deleted cache
    When the user deletes the gluon cache
    Then IMAP client "1" eventually sees the following messages in "INBOX":
      | from              | to                   | subject | date                  |
      | john.doe@mail.com | [user:user]@[domain] | foo     | 13 Aug 82 00:00 +0000 |

  Scenario: Fetch messages sent from Web Client
    When the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Sent":
      | from                 | to                | subject |
      | [user:user]@[domain] | john.doe@mail.com | foo     |
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                | subject |
      | [user:user]@[domain] | john.doe@mail.com | foo     |

  @regression
  Scenario: Fetch multiple messages sent from Web Client
    When the address "[user:user]@[domain]" of account "[user:user]" has 5 messages in "Sent"
    Then IMAP client "1" eventually sees 5 messages in "Sent"

  @regression
  Scenario: Fetch messages sent from Web Client in Split mode
    When the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    And user "[user:user]" connects and authenticates IMAP client "2" with address "[user:alias]@[domain]"
    Then it succeeds
    When the address "[user:user]@[domain]" of account "[user:user]" has the following messages in "Sent":
      | from                 | to                | subject |
      | [user:user]@[domain] | john.doe@mail.com | foo     |
    Then IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                | subject |
      | [user:user]@[domain] | john.doe@mail.com | foo     |
    When the address "[user:alias]@[domain]" of account "[user:user]" has the following messages in "Sent":
      | from                 | to                 | subject |
      | [user:alias]@[domain] | john.doe@mail.com | bar     |
    Then IMAP client "2" eventually sees the following messages in "Sent":
      | from                 | to                 | subject |
      | [user:alias]@[domain] | john.doe@mail.com | bar     |

  @regression
  Scenario: Fetch multiple messages sent from Web Client in Split mode
    When the user sets the address mode of user "[user:user]" to "split"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1" with address "[user:user]@[domain]"
    And user "[user:user]" connects and authenticates IMAP client "2" with address "[user:alias]@[domain]"
    Then it succeeds
    When the address "[user:user]@[domain]" of account "[user:user]" has 5 messages in "Sent"
    And the address "[user:alias]@[domain]" of account "[user:user]" has 10 messages in "Sent"
    Then IMAP client "1" eventually sees 5 messages in "Sent"
    Then IMAP client "2" eventually sees 10 messages in "Sent"
