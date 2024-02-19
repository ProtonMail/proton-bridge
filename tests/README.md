# Bridge Integration tests

Tests defined in this folder are using `github.com/cucumber/godog` library to
define scenarios.

The scenarios are defined in `./features/` folder.
The step definition can be found in `./steps_test.go`.


# How to run
All features are run as sub-test of `TestFeatures` in `./bdd_test.go`.
The most simple way to execute is `make test-integration` from project source directory.


There are several environment variables which can be used to control the tests:


* `FEATURES` sets the path to folder / file / line in file to select which
  scenarios to run.

        FEATURES=${PWD}/tests/features/user/addressmode.feature:162

* `FEATURE_TEST_LOG_LEVEL` the logrus level for tests (affects also testing
  bridge instance)

        FEATURE_TEST_LOG_LEVEL=trace

* `BRIDGE_API_DEBUG` when enabled
  [GPA](https://github.com/ProtonMail/go-proton-api/)
  client used in testing bridge instance will log http communication and logrus
  is automatically set to `trace`

        BRIDGE_API_DEBUG=1

* `GO_PROTON_API_SERVER_LOGGER_ENABLED` GPA mock server will print log line per
  each request to stdout (not logrus)

        GO_PROTON_API_SERVER_LOGGER_ENABLED=1

* `FEATURE_API_DEBUG` when enabled GPA client for preparation of test
  condiditions (see `./ctx_helper_test.go`) will dump http communication to
  stdoout.

        FEATURE_API_DEBUG=1

* `FEATURE_TEST_LOG_IMAP` when enabled
  bridge will dump all (client and server) IMAP communication to logs
  and logrus is automatically set to `trace`

        FEATURE_TEST_LOG_IMAP=1

* `GLUON_LOG_IMAP_LINE_LIMIT` controls maximal number of lines (by default 1)
  which are printed into imap trace log (logrus).
  Needs `FEATURE_TEST_LOG_IMAP` enabled to take effect.

        GLUON_LOG_IMAP_LINE_LIMIT=1048576


* `FEATURE_TEST_LOG_SMTP` when enabled
  bridge will dump all SMTP communication to logs
  and logrus is automatically set to `trace`

        FEATURE_TEST_LOG_SMTP=1




