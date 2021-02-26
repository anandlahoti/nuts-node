/*
 * Nuts node
 * Copyright (C) 2021 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package dummy

import (
	"encoding/json"
	"testing"

	"github.com/nuts-foundation/nuts-node/auth/contract"
	"github.com/nuts-foundation/nuts-node/auth/services"
	"github.com/stretchr/testify/assert"
)

func TestDummy_StartSigningSession(t *testing.T) {
	t.Run("returns error when in strictMode", func(t *testing.T) {
		d := Dummy{
			InStrictMode: true,
		}

		_, err := d.StartSigningSession("")

		assert.Error(t, err)
		assert.Equal(t, errNotEnabled, err)
	})

	t.Run("ok - valid sessionID", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
			Sessions:     make(map[string]string, 0),
			Status:       make(map[string]string, 0),
		}

		s, err := d.StartSigningSession("")

		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Len(t, s.SessionID(), 32)
		assert.Equal(t, s.Payload(), []byte("dummy"))
	})

	t.Run("ok - values are stored", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
			Sessions:     make(map[string]string, 0),
			Status:       make(map[string]string, 0),
		}

		s, err := d.StartSigningSession("contract")

		assert.NoError(t, err)
		assert.Len(t, d.Sessions, 1)
		assert.Equal(t, d.Sessions[s.SessionID()], "contract")
		assert.Len(t, d.Status, 1)
		assert.Equal(t, d.Status[s.SessionID()], SessionCreated)
	})
}

func TestDummy_SigningSessionStatus(t *testing.T) {
	t.Run("returns error when in strictMode", func(t *testing.T) {
		d := Dummy{
			InStrictMode: true,
		}

		_, err := d.SigningSessionStatus("")

		assert.Error(t, err)
		assert.Equal(t, errNotEnabled, err)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
			Sessions:     make(map[string]string, 0),
			Status:       make(map[string]string, 0),
		}

		_, err := d.SigningSessionStatus("")

		assert.Error(t, err)
		assert.Equal(t, services.ErrSessionNotFound, err)
	})

	t.Run("ok - returns correct statuses", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
			Sessions:     make(map[string]string, 0),
			Status:       make(map[string]string, 0),
		}

		s, err := d.StartSigningSession("contract")
		assert.NoError(t, err)

		// created
		s1, err := d.SigningSessionStatus(s.SessionID())
		assert.NoError(t, err)
		assert.Equal(t, SessionCreated, s1.Status())

		// in progress
		s1, err = d.SigningSessionStatus(s.SessionID())
		assert.NoError(t, err)
		assert.Equal(t, SessionInProgress, s1.Status())

		// created
		s1, err = d.SigningSessionStatus(s.SessionID())
		assert.NoError(t, err)
		assert.Equal(t, SessionCompleted, s1.Status())
		assert.Len(t, d.Sessions, 0)
		assert.Len(t, d.Status, 0)

	})

	t.Run("ok - returns correct data", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
			Sessions:     make(map[string]string, 0),
			Status:       make(map[string]string, 0),
		}

		s, err := d.StartSigningSession("contract")
		assert.NoError(t, err)

		// created
		s1, err := d.SigningSessionStatus(s.SessionID())
		assert.NoError(t, err)
		s2 := s1.(signingSessionResult)
		assert.Equal(t, "contract", s2.Request)
		assert.Equal(t, SessionCreated, s2.State)
		assert.Equal(t, s.SessionID(), s2.ID)
	})
}

func TestDummy_VerifyVP(t *testing.T) {
	t.Run("error - strictMode", func(t *testing.T) {
		d := Dummy{
			InStrictMode: true,
		}

		_, err := d.VerifyVP([]byte{}, nil)

		assert.Error(t, err)
		assert.Equal(t, errNotEnabled, err)
	})

	t.Run("ok", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
		}

		p := Presentation{
			VerifiablePresentationBase: contract.VerifiablePresentationBase{
				Context: []string{contract.VerifiableCredentialContext},
				Type:    []contract.VPType{contract.VerifiablePresentationType, VerifiablePresentationType},
			},
			Proof: Proof{
				Type:      NoSignatureType,
				Initials:  "I",
				Lastname:  "Tester",
				Birthdate: "1980-01-01",
				Email:     "tester@example.com",
				Contract:  "EN:PractitionerLogin:v3 I hereby declare to act on behalf of care org. This declaration is valid from maandag 1 oktober 12:00:00 until maandag 1 oktober 13:00:00.",
			},
		}

		j, _ := json.Marshal(p)
		vr, err := d.VerifyVP(j, nil)

		assert.NoError(t, err)
		assert.Equal(t, contract.Valid, vr.Validity)
		assert.Equal(t, VerifiablePresentationType, vr.VPType)
	})

	t.Run("error - incorrect json", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
		}

		_, err := d.VerifyVP([]byte("not json"), nil)

		assert.Error(t, err)
	})

	t.Run("error - incorrect contract", func(t *testing.T) {
		d := Dummy{
			InStrictMode: false,
		}

		p := Presentation{
			Proof: Proof{
				Contract: "Not a contract",
			},
		}

		j, _ := json.Marshal(p)
		_, err := d.VerifyVP(j, nil)

		assert.Error(t, err)
	})
}

func TestSigningSessionResult_Status(t *testing.T) {
	ssr := signingSessionResult{
		State: "state",
	}

	assert.Equal(t, "state", ssr.Status())
}

func TestSigningSessionResult_VerifiablePresentation(t *testing.T) {
	t.Run("returns nil when not completed", func(t *testing.T) {
		ssr := signingSessionResult{
			State: "state",
		}
		vp, _ := ssr.VerifiablePresentation()
		assert.Nil(t, vp)
	})

	t.Run("ok - correct data", func(t *testing.T) {
		ssr := signingSessionResult{
			State: SessionCompleted,
		}
		vp, _ := ssr.VerifiablePresentation()
		dvp := vp.(Presentation)

		assert.Equal(t, []string{contract.VerifiableCredentialContext}, dvp.Context)
		assert.Equal(t, []contract.VPType{contract.VerifiablePresentationType, VerifiablePresentationType}, dvp.Type)
		assert.Equal(t, "", dvp.Proof.Contract)
		assert.Equal(t, "1980-01-01", dvp.Proof.Birthdate)
		assert.Equal(t, "tester@example.com", dvp.Proof.Email)
		assert.Equal(t, "I", dvp.Proof.Initials)
		assert.Equal(t, "Tester", dvp.Proof.Lastname)
		assert.Equal(t, "NoSignature", dvp.Proof.Type)
	})
}
