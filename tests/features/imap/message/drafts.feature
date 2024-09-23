Feature: IMAP Draft messages
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "Drafts"
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Date: 01 Jan 1980 00:00:00 +0000

      This is a dra
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | body          |
      | This is a dra |
    And IMAP client "1" eventually sees 1 messages in "Drafts"


  Scenario: Draft edited locally
    When IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    And IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft
      Content-Type: text/plain
      To: someone@example.com

      This is a draft, but longer
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject     | body                        |
      | someone@example.com | Basic Draft | This is a draft, but longer |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft" in "Drafts"

  # The draft event is received from black but it's not processed to IMAP
  @skip-black
  Scenario: Draft edited remotely
    When the following fields were changed in draft 1 for address "[user:user]@[domain]" of account "[user:user]":
      | to                  | subject     | body                             |
      | someone@example.com | Basic Draft | This is a draft body, but longer |
    Then IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject     | body                             |
      | someone@example.com | Basic Draft | This is a draft body, but longer |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft" in "Drafts"
  
  # The draft event is received from black but it's not processed to IMAP
  @skip-black
  @regression
  Scenario: Draft edited remotely and sent from client
    When IMAP client "1" selects "Drafts"
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    Then it succeeds
    When there exists an account with username "[user:to]" and password "password"
    And there exists an account with username "[user:cc]" and password "password"
    And IMAP client "1" appends the following message to "Drafts":
      """
      From: Bridge Test <[user:user]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft
      Content-Type: text/plain
      To: Internal Bridge <[user:to]@[domain]>

      This is a draft
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                 | subject       | body            |
      | [user:to]@[domain] | Basic Draft   | This is a draft |
    When the following fields were changed in draft 1 for address "[user:user]@[domain]" of account "[user:user]":
      | cc                 | subject             | body                        |
      | [user:cc]@[domain] | Basic Draft Updated | This is a draft, but longer |
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                 | cc                 | subject             | body                        |
      | [user:to]@[domain] | [user:cc]@[domain] | Basic Draft Updated | This is a draft, but longer |
    When user "[user:user]" connects and authenticates SMTP client "1"
    And SMTP client "1" sends the following message from "[user:user]@[domain]" to "[user:to]@[domain]":
      """
      From: Bridge Test <[user:user]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft Updated
      Content-Type: text/plain
      To: Internal Bridge <[user:to]@[domain]>
      CC: Additional Internal Bridge <[user:cc]@[domain]>

      This is a draft, but longer
      """
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Sent":
      | to                 | cc                 | subject             | body                        |
      | [user:to]@[domain] | [user:cc]@[domain] | Basic Draft Updated | This is a draft, but longer |
    Then IMAP client "1" selects "Drafts"
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    And IMAP client "1" eventually sees 0 messages in "Drafts"


  # The draft event is received from black but it's not processed to IMAP
  @skip-black
  Scenario: Draft moved to trash remotely
    When draft 1 for address "[user:user]@[domain]" of account "[user:user]" was moved to trash
    Then IMAP client "1" eventually sees the following messages in "Trash":
      | body          |
      | This is a dra |
    And IMAP client "1" eventually sees 0 messages in "Drafts"

  @regression
  Scenario: Draft moved to trash locally and expunged
    When IMAP client "1" moves all messages from "Drafts" to "Trash"
    And IMAP client "1" eventually sees the following messages in "Trash":
      | body          |
      | This is a dra |
    And IMAP client "1" marks message 1 as deleted
    Then IMAP client "1" eventually sees that message at row 1 has the flag "\Deleted"
    When IMAP client "1" expunges
    Then IMAP client "1" eventually sees 0 messages in "Drafts"

  Scenario: Draft saved without "Date" header
    When IMAP client "1" selects "Drafts"
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Subject: Draft without Date
      Content-Type: text/plain
      To: someone@example.com

      This is a Draft without Date in header
      """
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject            | body                                   |
      | someone@example.com | Draft without Date | This is a Draft without Date in header |

  Scenario: Draft saved without "From" header
    When IMAP client "1" selects "Drafts"
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" appends the following message to "Drafts":
      """
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Draft without From
      Content-Type: text/plain
      To: someone@example.com

      This is a Draft without From in header
      """
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject            | body                                   |
      | someone@example.com | Draft without From | This is a Draft without From in header |

  @regression
  Scenario: Only one draft in Drafts and All Mail after editing it locally multiple times
    Given there exists an account with username "[user:to]" and password "password"
    And there exists an account with username "[user:cc]" and password "password"
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft
      Content-Type: text/plain
      To: someone@example.com

      This is a draft, but longer
      """
    And it succeeds
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    Then IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject     | body                        |
      | someone@example.com | Basic Draft | This is a draft, but longer |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft" in "Drafts"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | to                  | subject     | body                        |
      | someone@example.com | Basic Draft | This is a draft, but longer |
    And IMAP client "1" eventually sees 1 messages in "All Mail"
    And IMAP client "1" selects "Drafts"
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft
      Content-Type: text/plain
      To: Internal Bridge <[user:to]@[domain]>

      This is a draft, but longer with changed recipient
      """
    Then it succeeds
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                 | subject     | body                                               |
      | [user:to]@[domain] | Basic Draft | This is a draft, but longer with changed recipient |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft" in "Drafts"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | to                 | subject     | body                                               |
      | [user:to]@[domain] | Basic Draft | This is a draft, but longer with changed recipient |
    And IMAP client "1" eventually sees 1 messages in "All Mail"
    And IMAP client "1" selects "Drafts"
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: foo@bar.com
      Date: 01 Jan 1980 00:00:00 +0000
      Subject: Basic Draft Updated
      Content-Type: text/plain
      To: Internal Bridge <[user:to]@[domain]>
      CC: Additional Internal Bridge <[user:cc]@[domain]>

      This is a draft, but longer with changed recipient, updated subject, and CC
      """
    Then it succeeds
    And IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                 | cc                 | subject             | body                                                                        |
      | [user:to]@[domain] | [user:cc]@[domain] | Basic Draft Updated | This is a draft, but longer with changed recipient, updated subject, and CC |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft updated" in "Drafts"
    And IMAP client "1" eventually sees the following messages in "All Mail":
      | to                 | cc                 | subject             | body                                                                        |
      | [user:to]@[domain] | [user:cc]@[domain] | Basic Draft Updated | This is a draft, but longer with changed recipient, updated subject, and CC |
    And IMAP client "1" eventually sees 1 messages in "All Mail"
