Feature: IMAP Draft messages
  Background:
    Given there exists an account with username "user@pm.me" and password "password"
    And bridge starts
    And the user logs in with username "user@pm.me" and password "password"
    And user "user@pm.me" finishes syncing
    And user "user@pm.me" connects and authenticates IMAP client "1"
    And IMAP client "1" selects "Drafts"
    When IMAP client "1" appends the following message to "Drafts":
      """

      This is a dra
      """
    Then IMAP client "1" eventually sees the following messages in "Drafts":
      | body          |
      | This is a dra |
    And IMAP client "1" sees 1 messages in "Drafts"

  Scenario: Draft edited locally
    When IMAP client "1" marks message 1 as deleted
    And IMAP client "1" expunges
    And IMAP client "1" appends the following message to "Drafts":
      """
      Subject: Basic Draft
      Content-Type: text/plain
      To: someone@proton.me

      This is a draft, but longer
      """
    Then it succeeds
    And IMAP client "1" eventually sees the following messages in "Drafts":
      | to                | subject     | body                        |
      | someone@proton.me | Basic Draft | This is a draft, but longer |
    And IMAP client "1" sees 1 messages in "Drafts"

  Scenario: Draft edited remotely
    When the following fields where changed in draft 1 for address "user@pm.me" of account "user@pm.me":
      | to                | subject     | body                             |
      | someone@proton.me | Basic Draft | This is a draft body, but longer |
    Then IMAP client "1" eventually sees the following messages in "Drafts":
      | to                | subject     | body                             |
      | someone@proton.me | Basic Draft | This is a draft body, but longer |
    And IMAP client "1" sees 1 messages in "Drafts"

