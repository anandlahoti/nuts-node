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
 *
 */

package core

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

// Routable enables connecting a REST API to the echo server. The API wrappers should implement this interface
type Routable interface {
	// Routes configures the HTTP routes on the given router
	Routes(router EchoRouter)
}

// NewSystem creates a new, empty System.
func NewSystem() *System {
	serverCfg := NewServerConfig()
	result := &System{
		engines: []Engine{},
		Config:  serverCfg,
		Routers: []Routable{},
	}
	return result
}

// System is the control structure where engines are registered.
type System struct {
	// engines is the slice of all registered engines
	engines []Engine
	// Config holds the global and raw config
	Config *ServerConfig
	// Routers is used to connect API handlers to the echo server
	Routers []Routable
	// Context is cancelled when the system shuts down.
	Context context.Context
	// ContextCancel is a function to signal the system should shut down.
	ContextCancel context.CancelFunc
}

var coreLogger = logrus.StandardLogger().WithField(LogFieldModule, "core")

// Load loads the config and injects config values into engines
func (system *System) Load(flags *pflag.FlagSet) error {
	if err := system.Config.Load(flags); err != nil {
		return err
	}

	// visit each engine and inject the config
	return system.VisitEnginesE(func(engine Engine) error {
		if m, ok := engine.(Injectable); ok {
			return system.Config.InjectIntoEngine(m)
		}

		return nil
	})
}

// Diagnostics returns the compound diagnostics for all engines.
func (system *System) Diagnostics() []DiagnosticResult {
	result := make([]DiagnosticResult, 0)
	system.VisitEngines(func(engine Engine) {
		if m, ok := engine.(Diagnosable); ok {
			result = append(result, m.Diagnostics()...)
		}
	})
	return result
}

// Start starts all engines in the system.
func (system *System) Start() error {
	var err error
	return system.VisitEnginesE(func(engine Engine) error {
		if m, ok := engine.(Runnable); ok {
			coreLogger.Infof("Starting %s...", getEngineName(engine))
			err = m.Start()
			coreLogger.Infof("Started %s", getEngineName(engine))
		}
		return err
	})
}

// Shutdown shuts down all engines in the system.
func (system *System) Shutdown() error {
	var engines []Runnable
	system.VisitEngines(func(engine Engine) {
		if m, ok := engine.(Runnable); ok {
			engines = append(engines, m)
		}
	})
	for i := len(engines) - 1; i >= 0; i-- {
		curr := engines[i]
		coreLogger.Infof("Stopping %s...", getEngineName(curr))
		if err := curr.Shutdown(); err != nil {
			return err
		}
		coreLogger.Infof("Stopped %s", getEngineName(curr))
	}
	return nil
}

// Configure configures all engines in the system.
func (system *System) Configure() error {
	coreLogger.Debugf("Creating datadir: %s", system.Config.Datadir)
	var err error
	if err = os.MkdirAll(system.Config.Datadir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create datadir (dir=%s): %w", system.Config.Datadir, err)
	}
	return system.VisitEnginesE(func(engine Engine) error {
		// only if Engine is dynamically configurable
		if m, ok := engine.(Configurable); ok {
			coreLogger.Debugf("Configuring %s", getEngineName(engine))
			err = m.Configure(*system.Config)
			coreLogger.Debugf("Configured %s", getEngineName(engine))
		}
		return err
	})
}

// Migrate migrates data structures in an engine if needed.
func (system *System) Migrate() error {
	var err error
	return system.VisitEnginesE(func(engine Engine) error {
		// only if Engine is migratable
		if m, ok := engine.(Migratable); ok {
			coreLogger.Debugf("Migrating %s", getEngineName(engine))
			err = m.Migrate()
			coreLogger.Debugf("Migrated %s", getEngineName(engine))
		}
		return err
	})
}

// VisitEngines applies the given function on all engines in the system.
// It visits the engines in order they were registered.
func (system *System) VisitEngines(visitor func(engine Engine)) {
	_ = system.VisitEnginesE(func(engine Engine) error {
		visitor(engine)
		return nil
	})
}

// VisitEnginesE applies the given function on all engines in the system, stopping when an error is returned. The error
// is passed through.
// It visits the engines in order they were registered.
func (system *System) VisitEnginesE(visitor func(engine Engine) error) error {
	for _, e := range system.engines {
		if err := visitor(e); err != nil {
			return err
		}
	}

	return nil
}

// RegisterEngine is a helper func to add an engine to the list of engines from a different lib/pkg
func (system *System) RegisterEngine(engine Engine) {
	system.engines = append(system.engines, engine)
}

// RegisterRoutes is a helper func to register API routers so they can be linked to the echo server
func (system *System) RegisterRoutes(router Routable) {
	system.Routers = append(system.Routers, router)
}

// Runnable is the interface that groups the Start and Shutdown methods.
// When an engine implements these they will be called on startup and shutdown.
// Start and Shutdown should not be called more than once
type Runnable interface {
	Start() error
	Shutdown() error
}

// Migratable is the interface that defines if an engine is migratable.
// If an engine is migratable, Migrate is called between Configure() and Start().
// Migrations may require their own DB connection, they are closed before Start() is called.
type Migratable interface {
	Migrate() error
}

// Configurable is the interface that contains the Configure method.
// When an engine implements the Configurable interface, it will be called before startup.
// Configure should only be called once per engine instance
type Configurable interface {
	Configure(config ServerConfig) error
}

// ViewableDiagnostics is used for engines that display diagnostics in an interface
type ViewableDiagnostics interface {
	Named
	Diagnosable
}

// Diagnosable allows the implementer, mostly engines, to return diagnostics.
type Diagnosable interface {
	Diagnostics() []DiagnosticResult
}

// Engine is the base interface for a modular design
type Engine interface{}

// Named is the interface for all engines that have a name
type Named interface {
	// Name returns the name of the engine
	Name() string
}

// Injectable marks an engine capable of Config injection
type Injectable interface {
	Named
	// Config returns a pointer to the struct that holds the Config.
	Config() interface{}
}

func getEngineName(engine Engine) string {
	var name string
	if named, ok := engine.(Named); ok {
		name = named.Name()
	} else {
		name = fmt.Sprintf("%T", engine)
	}
	return name
}
