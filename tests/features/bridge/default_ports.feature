Feature: Bridge picks default ports wisely

  Scenario: bridge picks ports for IMAP and SMTP using default values.
    When bridge starts
    Then bridge IMAP port is 1143
    Then bridge SMTP port is 1025

  Scenario: bridge picks ports for IMAP wisely when default port is busy.
    When the network port 1143 is busy
    And bridge starts
    Then bridge IMAP port is 1144
    Then bridge SMTP port is 1025

  Scenario: bridge picks ports for SMTP wisely when default port is busy.
    When the network port range 1025-1030 is busy
    And bridge starts
    Then bridge IMAP port is 1143
    Then bridge SMTP port is 1031

  Scenario: bridge picks ports for IMAP SMTP wisely when default ports are busy.
    When the network port range 1025-1200 is busy
    And bridge starts
    Then bridge IMAP port is 1201
    Then bridge SMTP port is 1202
