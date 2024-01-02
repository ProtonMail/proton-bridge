// Copyright (c) 2024 Proton AG
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

// Memory management rules:
// Foundation object (Objective-C prefixed with `NS`) get ARC (Automatic Reference Counting), and do not need to be released manually.
// Core Foundation objects (C), prefixed with need to be released manually using CFRelease() unless:
// - They're obtained using a CF method containing the word Get (a.k.a. the Get Rule).
// - They're obtained using toll-free bridging from a Foundation Object (using the __bridge keyword).

//****************************************************************************************************************************************************
/// \brief Create a certificate object from DER-encoded data.
///
/// \return The certifcation. The caller is responsible for releasing the object using CFRelease.
/// \return NULL if data is not a valid DER-encoded certificate.
//****************************************************************************************************************************************************
SecCertificateRef certFromData(char const* data, uint64_t length) {
    NSData *der = [NSData dataWithBytes:data length:length];
    return SecCertificateCreateWithData(NULL, (__bridge CFDataRef)der);
}


//****************************************************************************************************************************************************
/// \brief Check if a certificate is in the user's keychain.
///
/// \param[in] cert The certificate.
/// \return true iff the certificate is in the user's keychain.
//****************************************************************************************************************************************************
bool _isCertificateInKeychain(SecCertificateRef const cert) {
    NSDictionary *attrs = @{
        (id)kSecMatchItemList: @[(__bridge id)cert],
        (id)kSecClass: (id)kSecClassCertificate,
        (id)kSecReturnData: @YES
    };
    return errSecSuccess == SecItemCopyMatching((__bridge CFDictionaryRef)attrs, NULL);
}

//****************************************************************************************************************************************************
/// \brief Check if a certificate is in the user's keychain.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return true iff the certificate is in the user's keychain.
//****************************************************************************************************************************************************
bool isCertificateInKeychain(char const* certData, uint64_t certSize) {
    return _isCertificateInKeychain(certFromData(certData, certSize));
}


//****************************************************************************************************************************************************
/// \brief Add a certificate to the user's keychain.
///
/// \param[in] cert The certificate.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus _addCertificateToKeychain(SecCertificateRef const cert) {
    NSDictionary* addQuery = @{
        (id)kSecValueRef: (__bridge id) cert,
        (id)kSecClass: (id)kSecClassCertificate,
    };
    return SecItemAdd((__bridge CFDictionaryRef) addQuery, NULL);
}

//****************************************************************************************************************************************************
/// \brief Add a certificate to the user's keychain.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus addCertificateToKeychain(char const* certData, uint64_t certSize) {
    return _addCertificateToKeychain(certFromData(certData, certSize));
}

//****************************************************************************************************************************************************
/// \brief Add a certificate to the user's keychain.
///
/// \param[in] cert The certificate.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus _removeCertificateFromKeychain(SecCertificateRef const cert) {
        NSDictionary *query = @{ (id)kSecClass: (id)kSecClassCertificate,
                                 (id)kSecMatchItemList: @[(__bridge id)cert],
                                 (id)kSecMatchLimit: (id)kSecMatchLimitOne,
        };
        return SecItemDelete((__bridge CFDictionaryRef) query);
}

//****************************************************************************************************************************************************
/// \brief Add a certificate to the user's keychain.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus removeCertificateFromKeychain(char const* certData, uint64_t certSize) {
    return _removeCertificateFromKeychain(certFromData(certData, certSize));
}

//****************************************************************************************************************************************************
/// \brief Check if a certificate is trusted in the user's keychain.
///
/// \param[in] cert The certificate.
/// \return true iff the certificate is trusted in the user's keychain.
//****************************************************************************************************************************************************
bool _isCertificateTrusted(SecCertificateRef const cert) {
    CFArrayRef trustSettings = NULL;
    OSStatus status = SecTrustSettingsCopyTrustSettings(cert, kSecTrustSettingsDomainUser, &trustSettings);
    if (status != errSecSuccess) {
        return false;
    }
    CFIndex count = CFArrayGetCount(trustSettings);
    bool result = false;
    for (CFIndex index = 0; index < count; ++index) {
        CFDictionaryRef dict = (CFDictionaryRef)CFArrayGetValueAtIndex(trustSettings, index);
        if (!dict) {
            continue;
        }
        CFNumberRef num = (CFNumberRef)CFDictionaryGetValue(dict, kSecTrustSettingsResult);
        int value;
        if (num && CFNumberGetValue(num, kCFNumberSInt32Type, &value) && (value == kSecTrustSettingsResultTrustRoot)) {
            result = true;
            break;
        }
    }
    CFRelease(trustSettings);
    return result;
}

//****************************************************************************************************************************************************
/// \brief Check if a certificate is trusted in the user's keychain.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return true iff the certificate is trusted in the user's keychain.
//****************************************************************************************************************************************************
bool isCertificateTrusted(char const* certData, uint64_t certSize) {
    return _isCertificateTrusted(certFromData(certData, certSize));
}

//****************************************************************************************************************************************************
/// \brief Set the trust level for a certificate in the user's keychain. This call will trigger a security prompt.
///
/// \param[in] cert The certificate.
/// \param[in] trustLevel The trust level.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus _setCertificateTrustLevel(SecCertificateRef const cert, int trustLevel) {
    SecPolicyRef policy = SecPolicyCreateSSL(true, NULL); // we limit our trust to SSL
    NSDictionary *trustSettings = @{
        (id)kSecTrustSettingsResult: [NSNumber numberWithInt:trustLevel],
        (id)kSecTrustSettingsPolicy: (__bridge id) policy,
    };
    OSStatus status = SecTrustSettingsSetTrustSettings(cert, kSecTrustSettingsDomainUser, (__bridge CFTypeRef)(trustSettings));
    CFRelease(policy);
    return status;
}

//****************************************************************************************************************************************************
/// \brief Set a certificate as trusted in the user's keychain. This call will trigger a security prompt.
///
/// \param[in] cert The certificate.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus _setCertificateTrusted(SecCertificateRef cert) {
    return _setCertificateTrustLevel(cert, kSecTrustSettingsResultTrustRoot);
}

//****************************************************************************************************************************************************
/// \brief Set a certificate as trusted in the user's keychain. This call will trigger a security prompt.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus setCertificateTrusted(char const* certData, uint64_t certSize) {
    return _setCertificateTrusted(certFromData(certData, certSize));
}

//****************************************************************************************************************************************************
/// \brief Remove the trust level of a certificate in the user's keychain.
///
/// \param[in] cert The certificate.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus _removeCertificateTrust(SecCertificateRef cert) {
    return _setCertificateTrustLevel(cert, kSecTrustSettingsResultUnspecified);
}

//****************************************************************************************************************************************************
/// \brief Remove the trust level of a certificate in the user's keychain.
///
/// \param[in] certData The certificate data in DER encoded format.
/// \param[in] certSize The size of the certData in bytes.
/// \return The status for the operation.
//****************************************************************************************************************************************************
OSStatus removeCertificateTrust(char const* certData, uint64_t certSize) {
    return _removeCertificateTrust(certFromData(certData, certSize));
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

// wrapCGoCertCallReturningBool wrap call to a CGo function returning a bool.
// if the certificate is invalid the call will return false.
func wrapCGoCertCallReturningBool(certPEM []byte, fn func(*C.char, C.ulonglong) bool) bool {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return false // error are ignored
	}

	buffer := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(buffer)) //nolint:unconvert

	return fn((*C.char)(buffer), C.ulonglong(len(certDER)))
}

// wrapCGoCertCallReturningBool wrap call to a CGo function returning an error
func wrapCGoCertCallReturningError(certPEM []byte, fn func(*C.char, C.ulonglong) error) error {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return err
	}

	buffer := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(buffer)) //nolint:unconvert

	return fn((*C.char)(buffer), C.ulonglong(len(certDER)))
}

// isCertInKeychain returns true if the given certificate is stored in the user's keychain.
func isCertInKeychain(certPEM []byte) bool {
	return wrapCGoCertCallReturningBool(certPEM, isCertInKeychainCGo)
}

func isCertInKeychainCGo(buffer *C.char, size C.ulonglong) bool {
	return bool(C.isCertificateInKeychain(buffer, size))
}

// addCertToKeychain adds a certificate to the user's keychain.
// Trying to add a certificate that is already in the keychain will result in an error.
func addCertToKeychain(certPEM []byte) error {
	return wrapCGoCertCallReturningError(certPEM, addCertToKeychainCGo)
}

func addCertToKeychainCGo(buffer *C.char, size C.ulonglong) error {
	if errCode := C.addCertificateToKeychain(buffer, size); errCode != errSecSuccess {
		return fmt.Errorf("could not add certificate to keychain (error %v)", errCode)
	}

	return nil
}

// removeCertFromKeychain removes a certificate from the user's keychain.
// Trying to remove a certificate that is not in the keychain will result in an error.
func removeCertFromKeychain(certPEM []byte) error {
	return wrapCGoCertCallReturningError(certPEM, removeCertFromKeychainCGo)
}

func removeCertFromKeychainCGo(buffer *C.char, size C.ulonglong) error {
	if errCode := C.removeCertificateFromKeychain(buffer, size); errCode != errSecSuccess {
		return fmt.Errorf("could not remove certificate from keychain (error %v)", errCode)
	}
	return nil
}

// isCertTrusted check if a certificate is trusted in the user's keychain.
func isCertTrusted(certPEM []byte) bool {
	return wrapCGoCertCallReturningBool(certPEM, isCertTrustedCGo)
}

func isCertTrustedCGo(buffer *C.char, size C.ulonglong) bool {
	return bool(C.isCertificateTrusted(buffer, size))
}

// setCertTrusted sets a certificate as trusted in the user's keychain.
// This function will trigger a security prompt from the system.
func setCertTrusted(certPEM []byte) error {
	return wrapCGoCertCallReturningError(certPEM, setCertTrustedCGo)
}

func setCertTrustedCGo(buffer *C.char, size C.ulonglong) error {
	errCode := C.setCertificateTrusted(buffer, size)
	switch errCode {
	case errSecSuccess:
		return nil
	case errAuthorizationCanceled:
		return ErrUserCanceledCertificateInstall
	default:
		return fmt.Errorf("could not set certificate trust in keychain (error %v)", errCode)
	}
}

// removeCertTrust remove the trust level of the certificated from the user's keychain.
// This function will trigger a security prompt from the system.
func removeCertTrust(certPEM []byte) error {
	return wrapCGoCertCallReturningError(certPEM, removeCertTrustCGo)
}

func removeCertTrustCGo(buffer *C.char, size C.ulonglong) error {
	errCode := C.removeCertificateTrust(buffer, size)
	switch errCode {
	case errSecSuccess:
		return nil
	case errAuthorizationCanceled:
		return ErrUserCanceledCertificateInstall
	default:
		return fmt.Errorf("could not set certificate trust in keychain (error %v)", errCode)
	}
}

func osSupportCertInstall() bool {
	return true
}

// installCert installs a certificate in the keychain. The certificate is added to the keychain and it is set as trusted.
// This function will trigger a security prompt from the system, unless the certificate is already trusted in the user keychain.
func installCert(certPEM []byte) error {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return err
	}

	p := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(p)) //nolint:unconvert
	buffer := (*C.char)(p)
	size := C.ulonglong(len(certDER))

	if !isCertInKeychainCGo(buffer, size) {
		if err := addCertToKeychainCGo(buffer, size); err != nil {
			return err
		}
	}

	if !isCertTrustedCGo(buffer, size) {
		return setCertTrustedCGo(buffer, size)
	}

	return nil
}

// uninstallCert uninstalls a certificate in the keychain. The certificate trust is removed and the certificated is deleted from the keychain.
// This function will trigger a security prompt from the system, unless the certificate is not trusted in the user keychain.
func uninstallCert(certPEM []byte) error {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return err
	}

	p := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(p)) //nolint:unconvert
	buffer := (*C.char)(p)
	size := C.ulonglong(len(certDER))

	if isCertTrustedCGo(buffer, size) {
		if err := removeCertTrustCGo(buffer, size); err != nil {
			return err
		}
	}

	if isCertInKeychainCGo(buffer, size) {
		return removeCertFromKeychainCGo(buffer, size)
	}

	return nil
}

func isCertInstalled(certPEM []byte) bool {
	certDER, err := certPEMToDER(certPEM)
	if err != nil {
		return false
	}

	p := C.CBytes(certDER)
	defer C.free(unsafe.Pointer(p)) //nolint:unconvert
	buffer := (*C.char)(p)
	size := C.ulonglong(len(certDER))

	return isCertInKeychainCGo(buffer, size) && isCertTrustedCGo(buffer, size)
}
