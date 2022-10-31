find_program(QMAKE_EXE "qmake")
if (NOT QMAKE_EXE)
    message(FATAL_ERROR "Could not locate qmake executable, make sur you have Qt 6 installed in that qmake is in your PATH environment variable.")
endif()
message(STATUS "Found qmake at ${QMAKE_EXE}")
execute_process(COMMAND "${QMAKE_EXE}" -query QT_INSTALL_PREFIX OUTPUT_VARIABLE QT_DIR OUTPUT_STRIP_TRAILING_WHITESPACE)

set(CMAKE_PREFIX_PATH  ${QT_DIR} ${CMAKE_PREFIX_PATH})