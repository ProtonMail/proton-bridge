Feature: The user reports a problem
  Background:
    Given bridge starts
    And it succeeds
  
  Scenario: The user wants to report a problem
    Then the description "Setup Outlook Windows" provides the following KB suggestions:
      """
      [
        {
          "url": "https://proton.me/support/protonmail-bridge-configure-client",
          "title": "How to configure your email client for Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-install",
          "title": "How to install Proton Mail Bridge"
        },
        {
          "url": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019",
          "title": "Proton Mail Bridge Microsoft Outlook for Windows 2019 setup guide"
        }
      ]
      """
