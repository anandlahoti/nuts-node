/*
 * Nuts node
 * Copyright (C) 2021. Nuts community
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

package crypto

import (
	"crypto"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/nuts-foundation/nuts-node/crypto/test"
	"github.com/nuts-foundation/nuts-node/test/io"

	"github.com/nuts-foundation/nuts-node/core"
	"github.com/nuts-foundation/nuts-node/crypto/storage"
	"github.com/stretchr/testify/assert"
)

func TestCryptoBackend(t *testing.T) {
	t.Run("Instance always returns same instance", func(t *testing.T) {
		client := Instance()
		client2 := Instance()

		if client != client2 {
			t.Error("Expected instances to be the same")
		}
	})
}

func TestCrypto_PublicKey(t *testing.T) {
	client := createCrypto(t)

	kid := "kid"
	ec := test.GenerateECKey()

	now := time.Now()
	period := core.Period{Begin: now}
	client.SavePublicKey(kid, ec.Public(), period)

	t.Run("Public key is returned from storage", func(t *testing.T) {
		pub, err := client.GetPublicKey(kid, now)

		assert.Nil(t, err)
		assert.NotEmpty(t, pub)
	})

	t.Run("error - unknown", func(t *testing.T) {
		_, err := client.GetPublicKey("unknown", now)

		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, storage.ErrNotFound))
		}
	})

	t.Run("error - kid not valid at time", func(t *testing.T) {
		_, err := client.GetPublicKey(kid, now.Add(-1))

		if assert.Error(t, err) {
			assert.True(t, errors.Is(err, storage.ErrNotFound))
		}
	})

	t.Run("error - saving an empty key", func(t *testing.T) {
		err := client.SavePublicKey(kid, nil, period)
		assert.Error(t, err)
	})
}

func TestCrypto_KeyExistsFor(t *testing.T) {
	client := createCrypto(t)

	kid := "kid"
	client.New(StringNamingFunc(kid))

	t.Run("returns true for existing key", func(t *testing.T) {
		assert.True(t, client.PrivateKeyExists(kid))
	})

	t.Run("returns false for non-existing key", func(t *testing.T) {
		assert.False(t, client.PrivateKeyExists("unknown"))
	})
}

func TestCrypto_New(t *testing.T) {
	client := createCrypto(t)

	t.Run("ok", func(t *testing.T) {
		kid := "kid"
		publicKey, returnKid, err := client.New(StringNamingFunc(kid))
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
		assert.Equal(t, kid, returnKid)
	})

	t.Run("error - NamingFunction returns err", func(t *testing.T) {
		errorNamingFunc := func(key crypto.PublicKey) (string, error) {
			return "", errors.New("b00m!")
		}
		_, _, err := client.New(errorNamingFunc)
		assert.Error(t, err)
	})
}

func TestCrypto_doConfigure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		e := createCrypto(t)
		err := e.doConfigure()
		assert.NoError(t, err)
	})
	t.Run("ok - default = fs backend", func(t *testing.T) {
		client := createCrypto(t)
		err := client.doConfigure()
		if !assert.NoError(t, err) {
			return
		}
		storageType := reflect.TypeOf(client.Storage).String()
		assert.Equal(t, "*storage.fileSystemBackend", storageType)
	})
	t.Run("error - unknown backend", func(t *testing.T) {
		client := createCrypto(t)
		client.Config.Storage = "unknown"
		err := client.doConfigure()
		assert.EqualErrorf(t, err, "only fs backend available for now", "expected error")
	})
}

func TestCrypto_Configure(t *testing.T) {
	createCrypto(t)

	t.Run("ok - configOnce", func(t *testing.T) {
		e := createCrypto(t)
		assert.False(t, e.configDone)
		err := e.Configure()
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, e.configDone)
		err = e.Configure()
		if !assert.NoError(t, err) {
			return
		}
		assert.True(t, e.configDone)
	})
}

func TestCryptoConfig_getFsPath(t *testing.T) {
	t.Run("no path configured returns defaultPath", func(t *testing.T) {
		c := Config{
			Fspath: "",
		}
		assert.Equal(t, "./", c.getFSPath())
	})
}

func createCrypto(t *testing.T) *Crypto {
	dir := io.TestDirectory(t)
	backend, _ := storage.NewFileSystemBackend(dir)
	crypto := Crypto{
		Storage: backend,
		Config:  TestCryptoConfig(dir),
	}

	return &crypto
}