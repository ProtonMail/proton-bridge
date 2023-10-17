Feature: Account settings

  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts

  Scenario: Check account default settings
  the account "[user:user]" has default draft format "HTML"