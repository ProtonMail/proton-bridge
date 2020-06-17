Feature: IMAP auth
  Scenario: Authenticates successfully
    Given there is connected user "user"
    When IMAP client authenticates "user"
    Then IMAP response is "OK"

  Scenario: Authenticates with bad password
    Given there is connected user "user"
    When IMAP client authenticates "user" with bad password
    Then IMAP response is "IMAP error: NO backend/credentials: incorrect password"

  Scenario: Authenticates with disconnected user
    Given there is disconnected user "user"
    When IMAP client authenticates "user"
    Then IMAP response is "IMAP error: NO account is logged out, use the app to login again"

  Scenario: Authenticates with connected user that was loaded without internet
    Given there is connected user "user"
    And there is no internet connection
    When bridge starts
    And the internet connection is restored
    And the event loop of "user" loops once
    And IMAP client authenticates "user"
    # Problems during IMAP auth could lead to the user being disconnected.
    # This could take a few milliseconds because it happens async in separate goroutines.
    # We wait enough time for that to happen, then check that it didn't happen (user should remain connected).
    And 2 seconds pass
    Then "user" is connected

  Scenario: Authenticates with freshly logged-out user
    Given there is connected user "user"
    When "user" logs out
    And IMAP client authenticates "user"
    Then IMAP response is "IMAP error: NO account is logged out, use the app to login again"

  Scenario: Authenticates user which was re-logged in
    Given there is connected user "user"
    When "user" logs out
    And IMAP client authenticates "user"
    Then IMAP response is "IMAP error: NO account is logged out, use the app to login again"
    When "user" logs in
    And IMAP client authenticates "user"
    Then IMAP response is "OK"
    When IMAP client selects "INBOX"
    Then IMAP response is "OK"

  Scenario: Authenticates with no user
    When IMAP client authenticates with username "user@pm.me" and password "bridgepassword"
    Then IMAP response is "IMAP error: NO user user@pm.me not found"

  Scenario: Authenticates with capital letter
    Given there is connected user "userAddressWithCapitalLetter"
    When IMAP client authenticates "userAddressWithCapitalLetter"
    Then IMAP response is "OK"

  Scenario: Authenticates with more addresses - primary one
    Given there is connected user "userMoreAddresses"
    When IMAP client authenticates "userMoreAddresses" with address "primary"
    Then IMAP response is "OK"

  Scenario: Authenticates with more addresses - secondary one
    Given there is connected user "userMoreAddresses"
    When IMAP client authenticates "userMoreAddresses" with address "secondary"
    Then IMAP response is "OK"

  Scenario: Authenticates with more addresses - disabled address
    Given there is connected user "userMoreAddresses"
    When IMAP client authenticates "userMoreAddresses" with address "disabled"
    Then IMAP response is "IMAP error: NO user .* not found"

  @ignore-live
  Scenario: Authenticates with secondary address of account with disabled primary address
    Given there is connected user "userDisabledPrimaryAddress"
    When IMAP client authenticates "userDisabledPrimaryAddress" with address "secondary"
    Then IMAP response is "OK"

  Scenario: Authenticates two users
    Given there is connected user "user"
    And there is connected user "userMoreAddresses"
    When IMAP client "imap1" authenticates "user"
    Then IMAP response to "imap1" is "OK"
    When IMAP client "imap2" authenticates "userMoreAddresses" with address "primary"
    Then IMAP response to "imap2" is "OK"

  Scenario: Logs out user
    Given there is connected user "user"
    When IMAP client logs out
    Then IMAP response is "OK"
