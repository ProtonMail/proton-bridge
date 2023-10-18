Feature: Account settings

  Background:
    Given there exists an account with username "[user:user]" and password "password"
    Then it succeeds
    When bridge starts

  Scenario: Check account default settings
  Then the account "[user:user]" matches the following settings:
    | DraftMIMEType | AttachPublicKey | Sign    | PGPScheme |
    | text/html     | false           | 0       | 0         |
  When the account "[user:user]" has public key attachment "enabled"
  And the account "[user:user]" has sign external messages "enabled"
  And the account "[user:user]" has default draft format "plain"
  And the account "[user:user]" has default PGP schema "inline"
  Then the account "[user:user]" matches the following settings:
    | DraftMIMEType | AttachPublicKey | Sign    | PGPScheme |
    | text/plain    | true            | 1       | 8         |

