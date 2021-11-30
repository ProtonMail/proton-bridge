Feature: IMAP operations with Drafts
  Background:
    Given there is connected user "user"
    And there are messages in mailbox "Drafts" for "user"
      | id   | from                              | subject                              | body |
      | msg1 | Lionel Richie <lionel@richie.com> | RE: Hello, is it me you looking for? | Nope |
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Drafts"


  Scenario: Draft subject updated locally

  Scenario: Draft recipient updated locally

  Scenario: Draft body updated locally

  @ignore-live
  Scenario: Draft subject updated on server side

  @ignore-live
  Scenario: Draft recipient updated on server side

  @ignore-live
  Scenario: Draft body and size updated on server side
    When IMAP client fetches body of UID "1"
    Then IMAP response is "OK"
    Then IMAP response contains "Nope"
    Given the body of draft "msg1" for "user" has changed to "Yes I am"
    And the event loop of "user" loops once
    And mailbox "Drafts" for "user" has 1 messages
    When IMAP client fetches body of UID "2"
    Then IMAP response is "OK"
    Then IMAP response contains "Yes I am"
    Then IMAP response does not contain "Nope"

