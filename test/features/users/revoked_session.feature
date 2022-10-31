Feature: Session deleted on API

  @ignore-live
  Scenario: Session revoked after start
    Given there is connected user "user"
    When session was revoked for "user"
    And the event loop of "user" loops once
    Then "user" is disconnected


  @ignore-live
  Scenario: Starting with revoked session
    Given there is user "user" which just logged in
    And session was revoked for "user"
    When bridge starts
    Then "user" is disconnected

