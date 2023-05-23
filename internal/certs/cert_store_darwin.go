// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package certs

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework Security
#import <Foundation/Foundation.h>
#import <Security/Security.h>


int installTrustedCert(char const *bytes, unsigned long long length) {
	if (length == 0) {
		return errSecInvalidData;
	}

	NSData *der = [NSData dataWithBytes:bytes length:length];

	// Step 1. Import the certificate in the keychain.
	SecCertificateRef cert = SecCertificateCreateWithData(NULL, (CFDataRef) der);
	NSDictionary* addQuery = @{
		(id)kSecValueRef: (__bridge id) cert,
		(id)kSecClass: (id)kSecClassCertificate,
	};

	OSStatus status = SecItemAdd((__bridge CFDictionaryRef) addQuery, NULL);
	if ((errSecSuccess != status) && (errSecDuplicateItem != status)) {
		CFRelease(cert);
		return status;
	}

	// Step 2. Set the trust for the certificate.
	SecPolicyRef policy = SecPolicyCreateSSL(true, NULL); // we limit our trust to SSL
	NSDictionary *trustSettings = @{
		(id)kSecTrustSettingsResult: [NSNumber numberWithInt:kSecTrustSettingsResultTrustRoot],
		(id)kSecTrustSettingsPolicy: (__bridge id) policy,
	};
	status = SecTrustSettingsSetTrustSettings(cert, kSecTrustSettingsDomainUser, (__bridge CFTypeRef)(trustSettings));
	CFRelease(policy);
	CFRelease(cert);

	return status;
}


int removeTrustedCert(char const *bytes, unsigned long long length) {
	if (0 == length) {
		return errSecInvalidData;
	}

	NSData *der = [NSData dataWithBytes: bytes length: length];
	SecCertificateRef cert = SecCertificateCreateWithData(NULL, (CFDataRef) der);

	// Step 1. Unset the trust for the certificate.
	SecPolicyRef policy = SecPolicyCreateSSL(true, NULL);
	NSDictionary * trustSettings = @{
		(id)kSecTrustSettingsResult: [NSNumber numberWithInt:kSecTrustSettingsResultUnspecified],
		(id)kSecTrustSettingsPolicy: (__bridge id) policy,
	};
	OSStatus status = SecTrustSettingsSetTrustSettings(cert, kSecTrustSettingsDomainUser, (__bridge CFTypeRef)(trustSettings));
	CFRelease(policy);
	if (errSecSuccess != status) {
		CFRelease(cert);
		return status;
	}

	// Step 2. Remove the certificate from the keychain.
	NSDictionary *query = @{ (id)kSecClass: (id)kSecClassCertificate,
							 (id)kSecMatchItemList: @[(__bridge id)cert],
							 (id)kSecMatchLimit: (id)kSecMatchLimitOne,
						   };
	status = SecItemDelete((__bridge CFDictionaryRef) query);

	CFRelease(cert);
	return status;
}
*/
import "C"

import (
	"encoding/pem"
	"errors"
	"fmt"
	"unsafe"
)

// some of the error codes returned by Apple's Security framework.
const (
	errSecSuccess            = 0
	errAuthorizationCanceled = -60006
)

// certPEMToDER converts a certificate in PEM format to DER format, which is the format required by Apple's Security framework.
func certPEMToDER(certPEM []byte) ([]byte, error) {
	block, left := pem.Decode(certPEM)
	if block == nil {
		return []byte{}, errors.New("invalid PEM certificate")
	}

	if len(left) > 0 {
		return []byte{}, errors.New("trailing data found at the end of a PEM certificate")
	}

	return block.Bytes, nil
}

func installCert(certPEM []byte) error {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return err
	}

	p := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(p)) //nolint:unconvert

	errCode := C.installTrustedCert((*C.char)(p), (C.ulonglong)(len(certDER)))
	switch errCode {
	case errSecSuccess:
		return nil
	case errAuthorizationCanceled:
		return fmt.Errorf("the user cancelled the authorization dialog")
	default:
		return fmt.Errorf("could not install certification into keychain (error %v)", errCode)
	}
}

func uninstallCert(certPEM []byte) error {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return err
	}

	p := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(p)) //nolint:unconvert

	if errCode := C.removeTrustedCert((*C.char)(p), (C.ulonglong)(len(certDER))); errCode != 0 {
		return fmt.Errorf("could not install certificate from keychain (error %v)", errCode)
	}

	return nil
}
