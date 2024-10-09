Feature: IMAP import messages

  Background:
    Given there exists an account with username "[user:user]" and password "password"
    And the account "[user:user]" has additional address "[alias:secondary]@[domain]"
    And the account "[user:user]" has additional disabled address "[alias:disabled]@[domain]"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    And user "[user:user]" connects and authenticates IMAP client "1"
    Then it succeeds

  @skip-black
  Scenario: Messages imported with default address as sender are encrypted with the default address key
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <[user:user]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[user:user]@[domain]" was used to import

  @skip-black
  Scenario: Messages imported with alias as sender are encrypted with secondary address key
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <[alias:secondary]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[alias:secondary]@[domain]" was used to import

  @skip-black
  Scenario: Messages imported with a disabled alias as sender are encrypted with the disabled address key
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <[alias:disabled]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[alias:disabled]@[domain]" was used to import

  @skip-black
  Scenario: Messages imported with an unknown address as sender are encrypted with primary address key
    When IMAP client "1" appends the following message to "INBOX":
      """
      From: Bridge Test <bridgeqa@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Received: by 2002:0:0:0:0:0:0:0 with SMTP id 0123456789abcdef; Wed, 30 Dec 2020 01:23:45 0000
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[user:user]@[domain]" was used to import

  @skip-black
  Scenario: Drafts imported with default address as sender are encrypted with the default address key
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: Bridge Test <[user:user]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[user:user]@[domain]" was used to create draft

  @skip-black
  Scenario: Drafts imported with alias as sender are encrypted with secondary key
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: Bridge Test <[alias:secondary]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[alias:secondary]@[domain]" was used to create draft

  @skip-black
  Scenario: Drafts imported with a disabled alias as sender are encrypted with the disabled address key
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: Bridge Test <[alias:disabled]@[domain]>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[user:user]@[domain]" was used to create drafts

  @skip-black
  Scenario: Drafts imported with an unknown address as sender are encrypted with primary address key
    When IMAP client "1" appends the following message to "Drafts":
      """
      From: Bridge Test <bridgeqa@example.com>
      Date: 01 Jan 1980 00:00:00 +0000
      To: Internal Bridge <bridgetest@example.com>
      Subject: Basic text/plain message
      Content-Type: text/plain

      Hello
      """
    Then it succeeds
    And the key for address "[user:user]@[domain]" was used to create draft
