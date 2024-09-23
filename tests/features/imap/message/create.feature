Feature: IMAP create messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[alias:alias]@[domain]"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  Scenario: Creates message to user's primary address
    When IMAP client "1" appends the following messages to "INBOX":
      | from               | to                   | subject | body |
      | john.doe@email.com | [user:user]@[domain] | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from               | to                   | subject | body |
      | john.doe@email.com | [user:user]@[domain] | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from               | to                   | subject | body |
      | john.doe@email.com | [user:user]@[domain] | foo     | bar  |

  Scenario: Creates draft
    When IMAP client "1" appends the following messages to "Drafts":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |
    # This fails now
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |

  Scenario: Creates message sent from user's primary address
    When IMAP client "1" appends the following messages to "Sent":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from                 | to                 | subject | body |
      | [user:user]@[domain] | john.doe@email.com | foo     | bar  |

  Scenario: Creates message sent from user's secondary address
    When IMAP client "1" appends the following messages to "Sent":
      | from                   | to                 | subject | body |
      | [alias:alias]@[domain] | john.doe@email.com | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from                   | to                 | subject | body |
      | [alias:alias]@[domain] | john.doe@email.com | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from                   | to                 | subject | body |
      | [alias:alias]@[domain] | john.doe@email.com | foo     | bar  |

  Scenario: Imports an unrelated message to inbox
    When IMAP client "1" appends the following messages to "INBOX":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "INBOX":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |

  Scenario: Imports an unrelated message to sent
    When IMAP client "1" appends the following messages to "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |

  Scenario: Imports a similar (duplicate) message to sent
    When IMAP client "1" appends the following messages to "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    And it succeeds
    And IMAP client "1" appends the following messages to "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | from               | to                 | subject | body |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
      | john.doe@email.com | john.doe2@[domain] | foo     | bar  |
