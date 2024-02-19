Feature: The user reports a problem
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    And user "[user:user]" finishes syncing
    Then it succeeds

  Scenario: User sends a problem report without logs attached
    When the user reports a bug
    Then the header in the "POST" multipart request to "/core/v4/reports/bug" has "Title" set to "[Bridge] Bug - title"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Description" set to "description"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Username" set to "[user:user]"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has no file "logs.zip"

  Scenario: User sends a problem report with logs attached
    When the user reports a bug with field "IncludeLogs" set to "true"
    Then it succeeds
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Title" set to "[Bridge] Bug - title"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Description" set to "description"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Username" set to "[user:user]"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has file "logs.zip"


  @regression
  Scenario: User sends a problem report while signed out of Bridge
    When user "[user:user]" logs out
    And the user reports a bug with field "Email" set to "[user:user]@[domain]"
    Then it succeeds
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Username" set to "[user:user]"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Email" set to "[user:user]@[domain]"

  @regression
  Scenario: User sends a problem report with changed Title
    When the user reports a bug with field "Title" set to "Testing title"
    Then the header in the "POST" multipart request to "/core/v4/reports/bug" has "Title" set to "[Bridge] Bug - Testing title"

  @regression
  Scenario: User sends a problem report with changed Description
    When the user reports a bug with field "Description" set to "There's an issue with my testing, please fix!"
    Then the header in the "POST" multipart request to "/core/v4/reports/bug" has "Description" set to "There's an issue with my testing, please fix!"

  @regression
  Scenario: User sends a problem report with multiple details changed
    When the user reports a bug with details:
      """
      {
        "Title": "Testing Title",
        "Description": "Testing Description",
        "Username": "[user:user]",
        "Email": "[user:user]@[domain]",
        "EmailClient": "Apple Mail",
        "IncludeLogs": true
      }
      """
    Then the header in the "POST" multipart request to "/core/v4/reports/bug" has "Title" set to "[Bridge] Bug - Testing Title"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "OS" set to "osType"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "OSVersion" set to "osVersion"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Description" set to "Testing Description"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Username" set to "[user:user]"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Email" set to "[user:user]@[domain]"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has "Client" set to "Apple Mail"
    And the header in the "POST" multipart request to "/core/v4/reports/bug" has file "logs.zip"
