// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import (
	"encoding/base64"
	"errors"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// Draft actions
const (
	DraftActionReply    = 0
	DraftActionReplyAll = 1
	DraftActionForward  = 2
)

// PackageFlag for send message package types
type PackageFlag int

func (p *PackageFlag) Has(flag PackageFlag) bool { return iHasFlag(int(*p), int(flag)) }
func (p *PackageFlag) HasAtLeastOne(flag PackageFlag) bool {
	return iHasAtLeastOneFlag(int(*p), int(flag))
}
func (p *PackageFlag) Is(flag PackageFlag) bool    { return iIsFlag(int(*p), int(flag)) }
func (p *PackageFlag) HasNo(flag PackageFlag) bool { return iHasNoneOfFlag(int(*p), int(flag)) }

// Send message package types.
const (
	InternalPackage         = PackageFlag(1)
	EncryptedOutsidePackage = PackageFlag(2)
	ClearPackage            = PackageFlag(4)
	PGPInlinePackage        = PackageFlag(8)
	PGPMIMEPackage          = PackageFlag(16)
	ClearMIMEPackage        = PackageFlag(32)
)

// SignatureFlag for send signature types.
type SignatureFlag int

func (p *SignatureFlag) Is(flag SignatureFlag) bool    { return iIsFlag(int(*p), int(flag)) }
func (p *SignatureFlag) Has(flag SignatureFlag) bool   { return iHasFlag(int(*p), int(flag)) }
func (p *SignatureFlag) HasNo(flag SignatureFlag) bool { return iHasNoneOfFlag(int(*p), int(flag)) }

// Send signature types.
const (
	SignatureNone            = SignatureFlag(0)
	SignatureDetached        = SignatureFlag(1)
	SignatureAttachedArmored = SignatureFlag(2)
)

// DraftReq defines paylod for creating drafts
type DraftReq struct {
	Message              *Message
	ParentID             string `json:",omitempty"`
	Action               int
	AttachmentKeyPackets []string
}

func (c *client) CreateDraft(m *Message, parent string, action int) (created *Message, err error) {
	createReq := &DraftReq{Message: m, ParentID: parent, Action: action, AttachmentKeyPackets: []string{}}

	req, err := c.NewJSONRequest("POST", "/mail/v4/messages", createReq)
	if err != nil {
		return
	}

	var res MessageRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	created, err = res.Message, res.Err()
	return
}

type AlgoKey struct {
	Key       string
	Algorithm string
}

type MessageAddress struct {
	Type                          PackageFlag
	EncryptedBodyKeyPacket        string `json:"BodyKeyPacket"` // base64-encoded key packet.
	Signature                     SignatureFlag
	EncryptedAttachmentKeyPackets map[string]string `json:"AttachmentKeyPackets"`
}

type MessagePackage struct {
	Addresses               map[string]*MessageAddress
	Type                    PackageFlag
	MIMEType                string
	EncryptedBody           string             `json:"Body"`           // base64-encoded encrypted data packet.
	DecryptedBodyKey        AlgoKey            `json:"BodyKey"`        // base64-encoded session key (only if cleartext recipients).
	DecryptedAttachmentKeys map[string]AlgoKey `json:"AttachmentKeys"` // Only include if cleartext & attachments.
}

func newMessagePackage(
	send sendData,
	attKeys map[string]AlgoKey,
) (pkg *MessagePackage) {
	pkg = &MessagePackage{
		EncryptedBody: base64.StdEncoding.EncodeToString(send.ciphertext),
		Addresses:     send.addressMap,
		MIMEType:      send.contentType,
		Type:          send.sharedScheme,
	}

	if send.sharedScheme.HasAtLeastOne(ClearPackage | ClearMIMEPackage) {
		pkg.DecryptedBodyKey.Key = send.decryptedBodyKey.GetBase64Key()
		pkg.DecryptedBodyKey.Algorithm = send.decryptedBodyKey.Algo
	}

	if len(attKeys) != 0 && send.sharedScheme.Has(ClearPackage) {
		pkg.DecryptedAttachmentKeys = attKeys
	}

	return pkg
}

type sendData struct {
	decryptedBodyKey *crypto.SessionKey //body session key
	addressMap       map[string]*MessageAddress
	sharedScheme     PackageFlag
	ciphertext       []byte
	cleartext        string
	contentType      string
}

type SendMessageReq struct {
	ExpirationTime int64 `json:",omitempty"`
	// AutoSaveContacts int `json:",omitempty"`

	// Data for encrypted recipients.
	Packages []*MessagePackage

	mime, plain, rich sendData
	attKeys           map[string]*crypto.SessionKey
	kr                *crypto.KeyRing
}

func NewSendMessageReq(
	kr *crypto.KeyRing,
	mimeBody, plainBody, richBody string,
	attKeys map[string]*crypto.SessionKey,
) *SendMessageReq {
	req := &SendMessageReq{}

	req.mime.addressMap = make(map[string]*MessageAddress)
	req.plain.addressMap = make(map[string]*MessageAddress)
	req.rich.addressMap = make(map[string]*MessageAddress)

	req.mime.cleartext = mimeBody
	req.plain.cleartext = plainBody
	req.rich.cleartext = richBody

	req.attKeys = attKeys
	req.kr = kr

	return req
}

var (
	errUnknownContentType           = errors.New("unknown content type")
	errMultipartInNonMIME           = errors.New("multipart mixed not allowed in this scheme")
	errAttSignNotSupported          = errors.New("attached signature not supported")
	errEncryptMustSign              = errors.New("encrypted package must be signed")
	errEncryptedOutsideNotSupported = errors.New("encrypted outside is not supported")
	errWrongSendScheme              = errors.New("wrong send scheme")
	errInternalMustEncrypt          = errors.New("internal package must be encrypted")
	errInlineMustEncrypt            = errors.New("PGP Inline package must be encrypted")
	errInlineMustBePlain            = errors.New("PGP Inline package must be plain text")
	errMissingPubkey                = errors.New("cannot encrypt body key packet: missing pubkey")
	errSignMustBeMultipart          = errors.New("clear signed html packet must be multipart")
	errMIMEMustBeMultipart          = errors.New("MIME packet must be multipart")
)

func (req *SendMessageReq) AddRecipient(
	email string, sendScheme PackageFlag,
	pubkey *crypto.KeyRing, signature SignatureFlag,
	contentType string, doEncrypt bool,
) (err error) {
	if signature.Has(SignatureAttachedArmored) {
		return errAttSignNotSupported
	}

	if doEncrypt && signature.HasNo(SignatureDetached) {
		return errEncryptMustSign
	}

	switch sendScheme {
	case PGPMIMEPackage, ClearMIMEPackage:
		if contentType != ContentTypeMultipartMixed {
			return errMIMEMustBeMultipart
		}
		return req.addMIMERecipient(email, sendScheme, pubkey, signature)
	case InternalPackage, ClearPackage, PGPInlinePackage:
		return req.addNonMIMERecipient(email, sendScheme, pubkey, signature, contentType, doEncrypt)
	case EncryptedOutsidePackage:
		return errEncryptedOutsideNotSupported
	default:
		return errWrongSendScheme
	}
}

func (req *SendMessageReq) addNonMIMERecipient(
	email string, sendScheme PackageFlag,
	pubkey *crypto.KeyRing, signature SignatureFlag,
	contentType string, doEncrypt bool,
) (err error) {
	if sendScheme.Is(ClearPackage) &&
		signature.Is(SignatureDetached) &&
		contentType == ContentTypeHTML {
		return errSignMustBeMultipart
	}

	var send *sendData

	switch contentType {
	case ContentTypePlainText:
		send = &req.plain
		send.contentType = ContentTypePlainText
	case ContentTypeHTML, "":
		send = &req.rich
		send.contentType = ContentTypeHTML
	case ContentTypeMultipartMixed:
		return errMultipartInNonMIME
	default:
		return errUnknownContentType
	}

	if send.decryptedBodyKey == nil {
		if send.decryptedBodyKey, send.ciphertext, err = encryptSymmDecryptKey(req.kr, send.cleartext); err != nil {
			return err
		}
	}
	newAddress := &MessageAddress{Type: sendScheme, Signature: signature}

	if sendScheme.Is(PGPInlinePackage) && !doEncrypt {
		return errInlineMustEncrypt
	}
	if sendScheme.Is(PGPInlinePackage) && contentType == ContentTypeHTML {
		return errInlineMustBePlain
	}
	if sendScheme.Is(InternalPackage) && !doEncrypt {
		return errInternalMustEncrypt
	}
	if doEncrypt && pubkey == nil {
		return errMissingPubkey
	}

	if doEncrypt {
		newAddress.EncryptedBodyKeyPacket, newAddress.EncryptedAttachmentKeyPackets, err = encryptAndEncodeSessionKeys(pubkey, send.decryptedBodyKey, req.attKeys)
		if err != nil {
			return err
		}
	}
	send.addressMap[email] = newAddress
	send.sharedScheme |= sendScheme

	return nil
}

func (req *SendMessageReq) addMIMERecipient(
	email string, sendScheme PackageFlag,
	pubkey *crypto.KeyRing, signature SignatureFlag,
) (err error) {
	req.mime.contentType = ContentTypeMultipartMixed
	if req.mime.decryptedBodyKey == nil {
		if req.mime.decryptedBodyKey, req.mime.ciphertext, err = encryptSymmDecryptKey(req.kr, req.mime.cleartext); err != nil {
			return err
		}
	}

	if sendScheme.Is(PGPMIMEPackage) {
		if pubkey == nil {
			return errMissingPubkey
		}
		// Attachment keys are not needed because attachments are part
		// of MIME body and therefore attachments are encrypted with
		// body session key.
		mimeBodyPacket, _, err := encryptAndEncodeSessionKeys(pubkey, req.mime.decryptedBodyKey, map[string]*crypto.SessionKey{})
		if err != nil {
			return err
		}
		req.mime.addressMap[email] = &MessageAddress{Type: sendScheme, EncryptedBodyKeyPacket: mimeBodyPacket, Signature: signature}
	} else {
		req.mime.addressMap[email] = &MessageAddress{Type: sendScheme, Signature: signature}
	}
	req.mime.sharedScheme |= sendScheme

	return nil
}

func (req *SendMessageReq) PreparePackages() {
	attkeysEncoded := make(map[string]AlgoKey)
	for attID, attkey := range req.attKeys {
		attkeysEncoded[attID] = AlgoKey{
			Key:       attkey.GetBase64Key(),
			Algorithm: attkey.Algo,
		}
	}

	for _, send := range []sendData{req.mime, req.plain, req.rich} {
		if len(send.addressMap) == 0 {
			continue
		}
		req.Packages = append(req.Packages, newMessagePackage(send, attkeysEncoded))
	}
}

type SendMessageRes struct {
	Res

	Sent *Message

	// Parent is only present if the sent message has a parent (reply/reply all/forward).
	Parent *Message
}

func (c *client) SendMessage(id string, sendReq *SendMessageReq) (sent, parent *Message, err error) {
	if id == "" {
		err = errors.New("pmapi: cannot send message with an empty id")
		return
	}

	if sendReq.Packages == nil {
		sendReq.Packages = []*MessagePackage{}
	}

	req, err := c.NewJSONRequest("POST", "/mail/v4/messages/"+id, sendReq)
	if err != nil {
		return
	}

	var res SendMessageRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	sent, parent, err = res.Sent, res.Parent, res.Err()
	return
}
