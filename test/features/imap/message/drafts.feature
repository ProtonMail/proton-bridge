Feature: IMAP operations with Drafts
  Background:
    Given there is connected user "user"
    And there is IMAP client logged in as "user"
    And there is IMAP client selected in "Drafts"
    And IMAP client imports message to "Drafts"
      """
      To: Lionel Richie <lionel@richie.com>
      Subject: RE: Hello, is it me you looking for?

      Nope.

      """
    And IMAP response is "OK"
    And API mailbox "<mailbox>" for "user" has 1 message


  Scenario: Draft subject updated on locally

  Scenario: Draft recipient updated on locally

  Scenario: Draft body updated on locally

  @ignore-live
  Scenario: Draft subject updated on server side

  @ignore-live
  Scenario: Draft recipient updated on server side

  @ignore-live
  Scenario: Draft body and size updated on server side

