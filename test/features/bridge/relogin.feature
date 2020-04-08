Feature: Re-login to bridge
  Scenario: Re-login with connected user and database file
    Given there is connected user "user"
    And there is database file for "user"
    When "user" logs in to bridge
    Then bridge response is "failed to finish login: user is already logged in"
    And "user" is connected
    And "user" has running event loop

  @ignore
  Scenario: Re-login with connected user and no database file
    Given there is connected user "user"
    And there is no database file for "user"
    When "user" logs in to bridge
    Then bridge response is "failed to finish login: user is already logged in"
    And "user" is connected
    And "user" has database file
    And "user" has running event loop

  Scenario: Re-login with disconnected user and database file
    Given there is disconnected user "user"
    And there is database file for "user"
    When "user" logs in to bridge
    Then bridge response is "OK"
    And "user" is connected
    And "user" has running event loop

  Scenario: Re-login with disconnected user and no database file
    Given there is disconnected user "user"
    And there is no database file for "user"
    When "user" logs in to bridge
    Then bridge response is "OK"
    And "user" is connected
    And "user" has database file
    And "user" has running event loop
