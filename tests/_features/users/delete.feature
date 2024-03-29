Feature: Delete user
  @ignore-live-auth
  Scenario: Deleting connected user
    Given there is user "user" which just logged in
    When user deletes "user"
    Then last response is "OK"
    And "user" has database file

  @ignore-live-auth
  Scenario: Deleting connected user with cache
    Given there is user "user" which just logged in
    When user deletes "user" with cache
    Then last response is "OK"
    And "user" does not have database file

  @ignore-live-auth
  Scenario: Deleting connected user without database file
    Given there is user "user" which just logged in
    And there is no database file for "user"
    When user deletes "user" with cache
    Then last response is "OK"

  # THIS IS BLOCKED BY ANTI-ABUSE
  @ignore-live
  Scenario: Deleting disconnected user
    Given there is disconnected user "user"
    When user deletes "user"
    Then last response is "OK"
    And "user" has database file

  # THIS IS BLOCKED BY ANTI-ABUSE
  @ignore-live
  Scenario: Deleting disconnected user with cache
    Given there is disconnected user "user"
    When user deletes "user" with cache
    Then last response is "OK"
    And "user" does not have database file

  # THIS IS BLOCKED BY ANTI-ABUSE
  @ignore-live
  Scenario: Deleting disconnected user without database file
    Given there is disconnected user "user"
    And there is no database file for "user"
    When user deletes "user" with cache
    Then last response is "OK"
