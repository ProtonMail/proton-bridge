Feature: Re-login
  @ignore-live-auth
  Scenario: Re-login with connected user and database file
    Given there is user "user" which just logged in
    And there is database file for "user"
    When "user" logs in
    Then last response is "failed to finish login: user is already connected"
    And "user" is connected
    And "user" has running event loop
    And "user" has non-zero space

  @ignore
  Scenario: Re-login with connected user and no database file
    Given there is user "user" which just logged in
    And there is no database file for "user"
    When "user" logs in
    Then last response is "failed to finish login: user is already connected"
    And "user" is connected
    And "user" has database file
    And "user" has running event loop

  @ignore-live-auth
  Scenario: Re-login with disconnected user and database file
    Given there is disconnected user "user"
    And there is database file for "user"
    When "user" logs in
    Then last response is "OK"
    And "user" is connected
    And "user" has running event loop
    And "user" has non-zero space

  @ignore-live-auth
  Scenario: Re-login with disconnected user and no database file
    Given there is disconnected user "user"
    And there is no database file for "user"
    When "user" logs in
    Then last response is "OK"
    And "user" is connected
    And "user" has database file
    And "user" has running event loop
    And "user" has non-zero space
