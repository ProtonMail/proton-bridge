Feature: User agent
  Background:
    Given there is connected user "user"

  Scenario: Get user agent
    Given there is IMAP client logged in as "user"
    When IMAP client sends ID with argument:
    """
    "name" "Foo" "version" "1.4.0"
    """
    Then API client manager user-agent is "Foo/1.4.0 ([GOOS])"

  Scenario: Update user agent
    Given there is IMAP client logged in as "user"
    When IMAP client sends ID with argument:
    """
    "name" "Foo" "version" "1.4.0"
    """
    Then API client manager user-agent is "Foo/1.4.0 ([GOOS])"
    When IMAP client sends ID with argument:
    """
    "name" "Bar" "version" "4.2.0"
    """
    Then API client manager user-agent is "Bar/4.2.0 ([GOOS])"
