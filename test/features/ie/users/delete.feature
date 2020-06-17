Feature: Delete user
  Scenario: Deleting connected user
    Given there is connected user "user"
    When user deletes "user"
    Then last response is "OK"

  Scenario: Deleting connected user with cache
    Given there is connected user "user"
    When user deletes "user" with cache
    Then last response is "OK"

  Scenario: Deleting disconnected user
    Given there is disconnected user "user"
    When user deletes "user"
    Then last response is "OK"

  Scenario: Deleting disconnected user with cache
    Given there is disconnected user "user"
    When user deletes "user" with cache
    Then last response is "OK"
