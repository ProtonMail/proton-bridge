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

  Scenario: Draft edited remotely
    When the following fields were changed in draft 1 for address "[user:user]@[domain]" of account "[user:user]":
      | to                  | subject     | body                             |
      | someone@example.com | Basic Draft | This is a draft body, but longer |
    Then IMAP client "1" eventually sees the following messages in "Drafts":
      | to                  | subject     | body                             |
      | someone@example.com | Basic Draft | This is a draft body, but longer |
    And IMAP client "1" eventually sees 1 messages in "Drafts"
    And IMAP client "1" does not see header "Reply-To" in message with subject "Basic Draft" in "Drafts"

  Scenario: Draft moved to trash remotely
    When draft 1 for address "[user:user]@[domain]" of account "[user:user]" was moved to trash
    Then IMAP client "1" eventually sees the following messages in "Trash":
      | body          |
      | This is a dra |
    And IMAP client "1" eventually sees 0 messages in "Drafts"

