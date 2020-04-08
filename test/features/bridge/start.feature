Feature: Start bridge
  Scenario: Start with connected user, database file and internet connection
    Given there is connected user "user"
    And there is database file for "user"
    When bridge starts
    Then "user" is connected
    And "user" has loaded store
    And "user" has running event loop
    And "user" has API auth

  Scenario: Start with connected user, database file and no internet connection
    Given there is connected user "user"
    And there is database file for "user"
    And there is no internet connection
    When bridge starts
    Then "user" is connected
    And "user" has loaded store
    And "user" has running event loop
    And "user" does not have API auth

  @ignore
  Scenario: Start with connected user, no database file and internet connection
    Given there is connected user "user"
    And there is no database file for "user"
    When bridge starts
    Then "user" is connected
    And "user" has loaded store
    And "user" has running event loop
    And "user" has API auth

  @ignore
  Scenario: Start with connected user, no database file and no internet connection
    Given there is connected user "user"
    And there is no database file for "user"
    And there is no internet connection
    When bridge starts
    Then "user" is disconnected
    And "user" does not have API auth

  Scenario: Start with disconnected user, database file and internet connection
    Given there is disconnected user "user"
    And there is database file for "user"
    When bridge starts
    Then "user" is disconnected
    And "user" has loaded store
    And "user" does not have running event loop
    And "user" does not have API auth

  Scenario: Start with disconnected user, database file and no internet connection
    Given there is disconnected user "user"
    And there is database file for "user"
    And there is no internet connection
    When bridge starts
    Then "user" is disconnected
    And "user" has loaded store
    And "user" does not have running event loop
    And "user" does not have API auth

  @ignore
  Scenario: Start with disconnected user, no database file and internet connection
    Given there is disconnected user "user"
    And there is no database file for "user"
    When bridge starts
    Then "user" is disconnected
    And "user" does not have loaded store
    And "user" does not have running event loop
    And "user" does not have API auth

  @ignore
  Scenario: Start with disconnected user, no database file and no internet connection
    Given there is disconnected user "user"
    And there is no database file for "user"
    And there is no internet connection
    When bridge starts
    Then "user" is disconnected
    And "user" does not have loaded store
    And "user" does not have running event loop
    And "user" does not have API auth
