# Copyright (c) 2022 Proton AG
#
# This file is part of Proton Mail Bridge.
#
# Proton Mail Bridge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Proton Mail Bridge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.


FINGERPRINT=7E8B0DD8EC8EAAFAD9BE0115927CCDB2C9C9CFD5
PASSPHRASE=$(jq -r '.mailboxPasswords.user' < ./test/accounts/fake.json)


export_key_from_testdata(){
    jq -r '.[0].PrivateKey' < ./test/testdata/user_key.json
}

import_tmp_key() {
    echo "$PASSPHRASE" |  gpg --batch --yes --passphrase-fd 0 --import "$1"
}

delete_test_keys_from_keyring() {
    gpg --delete-secret-key ${FINGERPRINT}
}

update_key_expiration(){
    echo -n "$PASSPHRASE" | xclip
    echo "RUN:
    key 0
    expire
    2y
    y
    ${PASSPHRASE} (should be in clipboard)
    key 1
    expire
    2y
    y
    save

    "
    gpg --edit-key ${FINGERPRINT}
}

export_new_key_to_armor() {
    echo -n "$PASSPHRASE" | xclip
    echo "passphrase '${PASSPHRASE}' is in your clipboard" >&2
    gpg --export-secret-keys --armor ${FINGERPRINT}
}

replace_testdata_keys() {
    jq ".[0].PrivateKey=\"$(cat "$1")\""  < "$2" > tmp
    mv tmp "$2"
}

update_test_keys(){
    tmpkey=tmp_key.asc

    export_key_from_testdata > ${tmpkey}
    import_tmp_key ${tmpkey}

    update_key_expiration

    export_new_key_to_armor > ${tmpkey}
    replace_testdata_keys ${tmpkey} ./test/testdata/user_key.json
    replace_testdata_keys ${tmpkey} ./test/testdata/address_key.json

    delete_test_keys_from_keyring
    rm -f ${tmpkey}
}

update_test_keys
