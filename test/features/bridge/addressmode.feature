Feature: Address mode
  Background:
    Given there is connected user "userMoreAddresses"
    And there is "userMoreAddresses" with mailbox "Folders/mbox"
    And there are messages in mailbox "Folders/mbox" for "userMoreAddresses"
      | from              | to          | subject |
      | john.doe@mail.com | [primary]   | foo     |
      | jane.doe@mail.com | [secondary] | bar     |

  Scenario: All messages in one mailbox with combined mode
    Given there is "userMoreAddresses" in "combined" address mode
    Then mailbox "Folders/mbox" for address "primary" of "userMoreAddresses" has messages
      | from              | to          | subject |
      | john.doe@mail.com | [primary]   | foo     |
      | jane.doe@mail.com | [secondary] | bar     |

  Scenario: Messages separated in more mailboxes with split mode
    Given there is "userMoreAddresses" in "split" address mode
    Then mailbox "Folders/mbox" for address "primary" of "userMoreAddresses" has messages
      | from              | to        | subject |
      | john.doe@mail.com | [primary] | foo     |
    And mailbox "Folders/mbox" for address "secondary" of "userMoreAddresses" has messages
      | from              | to          | subject |
      | jane.doe@mail.com | [secondary] | bar     |

  Scenario: Switch address mode from combined to split mode
    Given there is "userMoreAddresses" in "combined" address mode
    When "userMoreAddresses" changes the address mode
    Then bridge response is "OK"
    And "userMoreAddresses" has address mode in "split" mode
    And mailbox "Folders/mbox" for address "primary" of "userMoreAddresses" has messages
      | from              | to        | subject |
      | john.doe@mail.com | [primary] | foo     |
    And mailbox "Folders/mbox" for address "secondary" of "userMoreAddresses" has messages
      | from              | to          | subject |
      | jane.doe@mail.com | [secondary] | bar     |

  Scenario: Switch address mode from split to combined mode
    Given there is "userMoreAddresses" in "split" address mode
    When "userMoreAddresses" changes the address mode
    Then bridge response is "OK"
    And "userMoreAddresses" has address mode in "combined" mode
    And mailbox "Folders/mbox" for address "primary" of "userMoreAddresses" has messages
      | from              | to          | subject |
      | john.doe@mail.com | [primary]   | foo     |
      | jane.doe@mail.com | [secondary] | bar     |

  Scenario: Make secondary address primary in combined mode
    Given there is "userMoreAddresses" in "combined" address mode
    When "userMoreAddresses" swaps address "primary" with address "secondary"
    And "userMoreAddresses" receives an address event
    Then mailbox "Folders/mbox" for address "primary" of "userMoreAddresses" has messages
      | from              | to          | subject |
      | john.doe@mail.com | [primary]   | foo     |
      | jane.doe@mail.com | [secondary] | bar     |
