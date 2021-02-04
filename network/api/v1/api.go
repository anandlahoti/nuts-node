/*
 * Copyright (C) 2020. Nuts community
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
 *
 */

package v1

import (
	"github.com/labstack/echo/v4"
	hash2 "github.com/nuts-foundation/nuts-node/crypto/hash"
	"github.com/nuts-foundation/nuts-node/network"
	"github.com/nuts-foundation/nuts-node/network/log"
	"net/http"
)

// ServerImplementation returns a erver API implementation that uses the given network.
func ServerImplementation(network network.Network) ServerInterface {
	return &wrapper{Service: network}
}

type wrapper struct {
	Service network.Network
}

// ListDocuments lists all documents
func (a wrapper) ListDocuments(ctx echo.Context) error {
	documents, err := a.Service.ListDocuments()
	if err != nil {
		log.Logger().Errorf("Error while listing documents: %v", err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	results := make([]string, len(documents))
	for i, document := range documents {
		results[i] = string(document.Data())
	}
	return ctx.JSON(http.StatusOK, results)
}

// GetDocument returns a specific document
func (a wrapper) GetDocument(ctx echo.Context, hashAsString string) error {
	hash, err := hash2.ParseHex(hashAsString)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	document, err := a.Service.GetDocument(hash)
	if err != nil {
		log.Logger().Errorf("Error while retrieving document (hash=%s): %v", hash, err)
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if document == nil {
		return ctx.String(http.StatusNotFound, "document not found")
	}
	ctx.Response().Header().Set(echo.HeaderContentType, "application/jose")
	ctx.Response().WriteHeader(http.StatusOK)
	_, err = ctx.Response().Writer.Write(document.Data())
	return err
}

// GetDocumentPayload returns the payload of a specific document
func (a wrapper) GetDocumentPayload(ctx echo.Context, hashAsString string) error {
	hash, err := hash2.ParseHex(hashAsString)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	data, err := a.Service.GetDocumentPayload(hash)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	if data == nil {
		return ctx.String(http.StatusNotFound, "document or contents not found")
	}
	ctx.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	ctx.Response().WriteHeader(http.StatusOK)
	_, err = ctx.Response().Writer.Write(data)
	return err
}