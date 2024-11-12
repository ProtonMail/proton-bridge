Feature: The IMAP ID is propagated to bridge
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then it succeeds

  Scenario: Initial user agent before an IMAP client connects
    Then the user agent is "NoClient/0.0.1 ([GOOS])"

  Scenario: User agent before an IMAP client announces its ID
    When user "[user:user]" connects and authenticates IMAP client "1"
    Then the user agent is "UnknownClient/0.0.1 ([GOOS])"

  Scenario: User agent after an IMAP client announces its ID
    When user "[user:user]" connects IMAP client "1"
    And IMAP client "1" announces its ID with name "name" and version "version"
    Then the user agent is "name/version ([GOOS])"

  Scenario: User agent is used for API calls
    When user "[user:user]" connects IMAP client "1"
    And IMAP client "1" announces its ID with name "name" and version "version"
    When the user reports a bug
    Then the header in the "POST" request to "/core/v4/reports/bug" has "User-Agent" set to "name/version ([GOOS])"

  Scenario: User agent re-announces a new ID to IMAP client
    When user "[user:user]" connects IMAP client "1"
    And  IMAP client "1" announces its ID with name "name" and version "version"
    Then the user agent is "name/version ([GOOS])"
    And  IMAP client "1" announces its ID with name "new_name" and version "new_version"
    Then the user agent is "new_name/new_version ([GOOS])"

  Scenario: User agent re-announces a new ID to IMAP client and new ID is used for API calls
    When user "[user:user]" connects IMAP client "1"
    And  IMAP client "1" announces its ID with name "name" and version "version"
    When the user reports a bug
    Then the header in the "POST" request to "/core/v4/reports/bug" has "User-Agent" set to "name/version ([GOOS])"
    When IMAP client "1" announces its ID with name "new_name" and version "new_version"
    Then the user agent is "new_name/new_version ([GOOS])"
    When the user reports a bug
    Then the header in the "POST" request to "/core/v4/reports/bug" has "User-Agent" set to "new_name/new_version ([GOOS])"

  Scenario: Apple Notes user agent is ignored after IMAP client announces its ID
    When user "[user:user]" connects IMAP client "1"
    And  IMAP client "1" announces its ID with name "name" and version "version"
    Then the user agent is "name/version ([GOOS])"
    When IMAP client "1" announces its ID with name "Mac OS X Notes" and version "4.11"
    Then the user agent is "name/version ([GOOS])"

