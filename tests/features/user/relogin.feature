Feature: A logged out user can login again
  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts
    And the user logs in with username "[user:user]" and password "password"
    Then it succeeds

  Scenario: Login to disconnected account
    When user "[user:user]" logs out
    And bridge restarts
    And the user logs in with username "[user:user]" and password "password"
    Then user "[user:user]" is eventually listed and connected

  Scenario: Cannot login to removed account
    When user "[user:user]" is deleted
    Then user "[user:user]" is not listed

  Scenario: Bridge password persists after logout/login
    Given there exists an account with username "testUser" and password "password"
    And the user logs in with username "testUser" and password "password"
    And the bridge password of user "testUser" is changed to "YnJpZGdlcGFzc3dvcmQK"
    And user "testUser" is deleted
    And the user logs in with username "testUser" and password "password"
    Then user "testUser" is eventually listed and connected
    And the bridge password of user "testUser" is equal to "YnJpZGdlcGFzc3dvcmQK"
