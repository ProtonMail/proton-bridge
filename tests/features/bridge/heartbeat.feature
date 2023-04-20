Feature: Send Telemetry Heartbeat
  Background:
    Given there exists an account with username "[user:user1]" and password "password"
    And bridge starts


  Scenario: Send at first start - one user
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge eventually sends the following heartbeat:
      """
      {
        "MeasurementGroup": "bridge.any.usage",
        "Event": "bridge_heartbeat",
        "Values": {
          "nb_account": 1
        },
        "Dimensions": {
          "auto_update": "on",
          "auto_start": "on",
          "beta": "off",
          "doh": "off",
          "split_mode": "off",
          "show_all_mail": "on",
          "imap_connection_mode": "starttls",
          "smtp_connection_mode": "starttls",
          "imap_port": "default",
          "smtp_port": "default",
          "cache_location": "default",
          "keychain_pref": "default",
          "prev_version": "0.0.0",
          "rollout": 42
        }
      }
      """
    And bridge do not need to send heartbeat

  Scenario: GroupMeasurement rejected by API
    Given heartbeat is not whitelisted
    Then bridge telemetry feature is enabled
    And bridge needs to send heartbeat
    When the user logs in with username "[user:user1]" and password "password"
    And user "[user:user1]" finishes syncing
    Then bridge needs to send heartbeat

