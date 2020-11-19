Feature: Import from EML files
  Background:
    Given there is connected user "user"
    And there is EML file "Inbox/clear.eml"
      """
      Subject: clear
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>

			secret
      """
    And there is EML file "Inbox/encrypted.eml"
      """
      Subject: encrypted
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>

      -----BEGIN PGP MESSAGE-----
      
      hQEMA7hGUUsYs0fEAQgA10NwJSNTLm3vpxVtoYBaA9AjFI5H4hKuV3/f2NHbWb2s
      k5enK3tEIOYdFdrO+NLrV6weHq3Dgu4er3URTQ62tFwjSJyeXxmY0d9J+JdxibJg
      wqZubuSHYzQHpFqJgoUUWEVEsv0Ao8yQo8BE9iybCKoZf6KojROUuE748oxlxJnV
      m1XuaVIzgw4xN0GUA5sLLuWeL94b2dZe5SDDQE5POzDgueZ7faefX8U1pGErCRJ0
      sO6FSw3SF4NpvrxVESWgCmsG5pcuxE2JqB0UoHnNDcqsW8w1Q+GabAPo6UqHhgIg
      56MRCWeou2djHIIj9TMUIVFzSG/HvTYQWVS+i4Z7AdJJAXr53GgbZQznO80Qxwcb
      FFdlwOXHuaXqhqCb338jlQWnbcnUsuJWxBAxkHrlP/nluFqPdIBOglC38kdYSBed
      3YwuEB9sXV/fcw==
      =B05V
      -----END PGP MESSAGE-----
      """
    And there is EML file "Inbox/encrypted-mime.eml"
      """
      Subject: encrypted mime
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>
      MIME-Version: 1.0
      Content-Type: multipart/encrypted; protocol="application/pgp-encrypted"; boundary="WLjzd46aUAiOcuNXjWTJItBZonI56MuAk"

      --WLjzd46aUAiOcuNXjWTJItBZonI56MuAk
      Content-Type: application/pgp-encrypted
      Content-Description: PGP/MIME version identification


      --WLjzd46aUAiOcuNXjWTJItBZonI56MuAk
      Content-Type: application/octet-stream; name="encrypted.asc"
      Content-Description: OpenPGP encrypted message
      Content-Disposition: inline; filename="encrypted.asc"

      -----BEGIN PGP MESSAGE-----

      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
      =4M3l
      -----END PGP MESSAGE-----

      --WLjzd46aUAiOcuNXjWTJItBZonI56MuAk--
      """
    And there is EML file "Inbox/signed.eml"
      """
      Subject: signed
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>

      -----BEGIN PGP SIGNED MESSAGE-----
      Hash: SHA256
      
      secret
      -----BEGIN PGP SIGNATURE-----
      
      iQEzBAEBCAAdFiEENaE2ZPemenI4pZah/SJcGo7SJWIFAl+cCAgACgkQ/SJcGo7S
      JWKsOQf/YakNXkMNjZIu8Hf1WflxtiDXVzTugOicC05k5W64oIqSHt0xNaFKE37k
      //3eDMWbHvqHKFVdg7qcLsVPeVBaW3bdZUiexGM24OiGgyEitufnHQLOtEDTound
      JyH5nUeHpvpBKIIOJZNBDM0HsRYnwKwrOWk3N2VRwog4J8J3cmJ/f9bPWNI/0OPT
      qmtVGRVg6Ge83nZn51Vof//jFzkO4wGYCsE0aF0Ywc7nISZuyKQzmu/qgmwzDG50
      PjpvIQ/ygisRPNdRlylXEqyoIDCQ+v0AnxhhwX/5dbt6xMuMMOxBrFSC94Zce1Vj
      x2ssXlT4ONPnkI/YWwhtQPLU628IMg==
      =GiS3
      -----END PGP SIGNATURE-----
      """
    And there is EML file "Inbox/encrypted-signed.eml"
      """
      Subject: encrypted and signed
      From: Bridge Test <bridgetest@pm.test>
      To: Internal Bridge <test@protonmail.com>

      -----BEGIN PGP MESSAGE-----
      
      hQEMA7hGUUsYs0fEAQf/dppHciWIf+o4l0gEfHeyHV/HVhG4es0aVQYrwFQlSWVx
      estMuyLBSMfrsQXLago7Q9ZNo/XnKszzprCXxxYH52hAg64oAsjKB3jgRmVizs8b
      8lj0BRf003wUluS/0msV9SiEZBGeL8jGq6Te9vaM8OHHhIVzVjGnRdTSC0jBE6cS
      vy8IBHXYe0LfdZiPojPDPGQdSej+H3uu7eZGBvVHTDeQLPDel4k7Ykdr0qlNXs6O
      5XpM5YG4w+t0aG+YROPH+BUj8PpPojQ/lrv/yFISTRbHlEd8N50w8BNTnBet+9Vm
      oPcyvN+RQxBlvRuPpDjUmREvmtObKZV6+m6gocemx9LAzQEeVLcpjO/hJhl8gX72
      MNz3McU7aXf5sSoOPdHDNx8T2NON/2bwG5FE+PRMuVywTKhCB7o8VAsJpGMQ8xRM
      5WCNhow0AI7kni8yZA+GbvspnJWfit9tCTR5MIFHCSH9J3kJJnWkxQSN04GGpBcd
      n43GWn7O7ufA4lMMZiGXMdi/J1iV9waAsIfMPk29BMq6xK0/jJYdHqQS+vNsSnF5
      xL/Ir4RYq4SFFA06A/E7HpXr2ruZhBQCkzaIIdrVJR/Lp2VLJIVulTBQK8y2AFtj
      JeeKS0kIuC/7UPF2O624kwNr8dmIhDJYusFs6ZeED/nAKwDO/vP2CSwVC3sUjn3N
      u+sWqQUTxSmjhRVf9b0+VyTh0mXCovJQXomL6Zz6lxXuJqqzELIOfCxYD1z9GwTG
      cT08Aa2eEpf3agdLCTxvjO3iq9FksMHvIN+LSCQ6Pw+aTByjrk0oMmvGbANAogTk
      yrplG/iRVlmq0p/Cfl5UEjKqT/nt5j9zbpeuYXmhjiBT9SBE07oUVLY1VT7ihcY=
      =qYnL
      -----END PGP MESSAGE-----
      """

  Scenario: Import encrypted
		Given there is skip encrypted messages set to "false"
    When user "user" imports local files
    Then progress result is "OK"
    And transfer failed for 0 messages
    And transfer exported 5 messages
    And transfer imported 5 messages
		And transfer skipped 0 messages
    And API mailbox "INBOX" for "user" has messages
      | from               | to                  | subject              |
      | bridgetest@pm.test | test@protonmail.com | clear                |
      | bridgetest@pm.test | test@protonmail.com | encrypted            |
      | bridgetest@pm.test | test@protonmail.com | encrypted mime       |
      | bridgetest@pm.test | test@protonmail.com | signed               |
      | bridgetest@pm.test | test@protonmail.com | encrypted and signed |

  Scenario: Skip encrypted
		Given there is skip encrypted messages set to "true"
    When user "user" imports local files
    Then progress result is "OK"
    And transfer failed for 0 messages
    And transfer exported 5 messages
    And transfer imported 2 messages
		And transfer skipped 3 messages
    And API mailbox "INBOX" for "user" has messages
      | from               | to                  | subject |
      | bridgetest@pm.test | test@protonmail.com | clear   |
      | bridgetest@pm.test | test@protonmail.com | signed  |
