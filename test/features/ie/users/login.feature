Feature: Login for the first time
  Scenario: Normal login
    Given there is user "user"
    When "user" logs in
    Then last response is "OK"
    And "user" is connected

  @ignore-live
  Scenario: Login with bad username
    When "user" logs in with bad password
    Then last response is "failed to login: Incorrect login credentials. Please try again"

  @ignore-live
  Scenario: Login with bad password
    Given there is user "user"
    When "user" logs in with bad password
    Then last response is "failed to login: Incorrect login credentials. Please try again"

  Scenario: Login without internet connection
    Given there is no internet connection
    When "user" logs in
    Then last response is "failed to login: cannot reach the server"

  @ignore-live
  Scenario: Login user with 2FA
    Given there is user "user2fa"
    When "user2fa" logs in
    Then last response is "OK"
    And "user2fa" is connected

  Scenario: Login user with capital letters in address
    Given there is user "userAddressWithCapitalLetter"
    When "userAddressWithCapitalLetter" logs in
    Then last response is "OK"
    And "userAddressWithCapitalLetter" is connected

  Scenario: Login user with more addresses
    Given there is user "userMoreAddresses"
    When "userMoreAddresses" logs in
    Then last response is "OK"
    And "userMoreAddresses" is connected

  @ignore-live
  Scenario: Login user with disabled primary address
    Given there is user "userDisabledPrimaryAddress"
    When "userDisabledPrimaryAddress" logs in
    Then last response is "OK"
    And "userDisabledPrimaryAddress" is connected

  Scenario: Login two users
    Given there is user "user"
    And there is user "userMoreAddresses"
    When "user" logs in
    Then last response is "OK"
    And "user" is connected
    When "userMoreAddresses" logs in
    Then last response is "OK"
    And "userMoreAddresses" is connected
