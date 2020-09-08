Feature: Re-login
  Scenario: Re-login with connected user
    Given there is connected user "user"
    When "user" logs in
    Then last response is "failed to finish login: user is already connected"
    And "user" is connected

  Scenario: Re-login with disconnected user
    Given there is disconnected user "user"
    When "user" logs in
    Then last response is "OK"
    And "user" is connected
