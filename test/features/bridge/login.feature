Feature: Login to bridge for the first time
  Scenario: Normal bridge login
    Given there is user "user"
    When "user" logs in to bridge
    Then bridge response is "OK"
    And "user" is connected
    And "user" has database file
    And "user" has running event loop

  Scenario: Login with bad username
    When "user" logs in to bridge with bad password
    Then bridge response is "failed to login: Incorrect login credentials. Please try again"

  Scenario: Login with bad password
    Given there is user "user"
    When "user" logs in to bridge with bad password
    Then bridge response is "failed to login: Incorrect login credentials. Please try again"

  Scenario: Login without internet connection
    Given there is no internet connection
    When "user" logs in to bridge
    Then bridge response is "failed to login: cannot reach the server"

  @ignore-live
  Scenario: Login user with 2FA
    Given there is user "user2fa"
    When "user2fa" logs in to bridge
    Then bridge response is "OK"
    And "user2fa" is connected
    And "user2fa" has database file
    And "user2fa" has running event loop

  Scenario: Login user with capital letters in address
    Given there is user "userAddressWithCapitalLetter"
    When "userAddressWithCapitalLetter" logs in to bridge
    Then bridge response is "OK"
    And "userAddressWithCapitalLetter" is connected
    And "userAddressWithCapitalLetter" has database file
    And "userAddressWithCapitalLetter" has running event loop

  Scenario: Login user with more addresses
    Given there is user "userMoreAddresses"
    When "userMoreAddresses" logs in to bridge
    Then bridge response is "OK"
    And "userMoreAddresses" is connected
    And "userMoreAddresses" has database file
    And "userMoreAddresses" has running event loop

  @ignore-live
  Scenario: Login user with disabled primary address
    Given there is user "userDisabledPrimaryAddress"
    When "userDisabledPrimaryAddress" logs in to bridge
    Then bridge response is "OK"
    And "userDisabledPrimaryAddress" is connected
    And "userDisabledPrimaryAddress" has database file
    And "userDisabledPrimaryAddress" has running event loop

  Scenario: Login two users
    Given there is user "user"
    And there is user "userMoreAddresses"
    When "user" logs in to bridge
    Then bridge response is "OK"
    And "user" is connected
    When "userMoreAddresses" logs in to bridge
    Then bridge response is "OK"
    And "userMoreAddresses" is connected
