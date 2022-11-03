Feature: Frontend events
  Scenario: Frontend starts and stops
    Given bridge is version "2.3.0" and the latest available version is "2.3.0" reachable from "2.3.0"
    When bridge starts
    Then frontend sees that bridge is version "2.3.0"
