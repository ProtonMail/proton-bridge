Feature: IMAP create messages
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And the account "user@pm.me" has additional address "alias@pm.me"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"

  Scenario: Creates message to user's primary address
    When IMAP client "1" appends the following messages to "INBOX":
      | from               | to         | subject | body |
      | john.doe@email.com | user@pm.me | foo     | bar  |
    Then it succeeds
    And IMAP client "1" sees the following messages in "INBOX":
      | from               | to         | subject | body |
      | john.doe@email.com | user@pm.me | foo     | bar  |
    And IMAP client "1" sees the following messages in "All Mail":
      | from               | to         | subject | body |
      | john.doe@email.com | user@pm.me | foo     | bar  |

  Scenario: Creates draft
    When IMAP client "1" appends the following messages to "Drafts":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |

  Scenario: Creates message sent from user's primary address
    When IMAP client "1" appends the following messages to "Sent":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" sees the following messages in "Sent":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |
    And IMAP client "1" sees the following messages in "All Mail":
      | from       | to                 | subject | body |
      | user@pm.me | john.doe@email.com | foo     | bar  |

  Scenario: Creates message sent from user's secondary address
    When IMAP client "1" appends the following messages to "Sent":
      | from        | to                 | subject | body |
      | alias@pm.me | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" sees the following messages in "Sent":
      | from        | to                 | subject | body |
      | alias@pm.me | john.doe@email.com | foo     | bar  |
    And IMAP client "1" sees the following messages in "All Mail":
      | from        | to                 | subject | body |
      | alias@pm.me | john.doe@email.com | foo     | bar  |

  Scenario: Imports an unrelated message to inbox
    When IMAP client "1" appends the following messages to "INBOX":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |
    Then it succeeds
    And IMAP client "1" sees the following messages in "INBOX":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |
    And IMAP client "1" sees the following messages in "All Mail":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |

  Scenario: Imports an unrelated message to sent
    When IMAP client "1" appends the following messages to "Sent":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |
    Then it succeeds
    And IMAP client "1" sees the following messages in "Sent":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |
    And IMAP client "1" sees the following messages in "All Mail":
      | from               | to              | subject | body |
      | john.doe@email.com | john.doe2@pm.me | foo     | bar  |
