// MIT License
//
// Copyright (c) 2022 Spiral Scout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package logger

import (
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"go.uber.org/zap"
)

// PluginName declares plugin name.
const PluginName = "logs"

// Plugin manages zap logger.
type Plugin struct {
	base     *zap.Logger
	cfg      *Config
	channels ChannelConfig
	version  string
	cmd      string
}

func (p *Plugin) Init(cfg config.Configurer) error {
	p.version = cfg.GetVersion()
	p.cmd = cfg.GetCmd()

	const op = errors.Op("logs plugin init")
	var err error

	if cfg.Has(PluginName) {
		if err = cfg.UnmarshalKey(PluginName, &p.cfg); err != nil {
			return errors.E(op, errors.Init, err)
		}
		if err = cfg.UnmarshalKey(PluginName, &p.channels); err != nil {
			return errors.E(op, errors.Init, err)
		}
	} else {
		p.cfg = &Config{}
	}

	p.cfg.InitDefault()
	p.base, err = p.cfg.BuildLogger()
	if err != nil {
		return errors.E(op, errors.Init, err)
	}
	p.base = p.base.With(p.fields()...)

	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	_ = p.base.Sync()
	return nil
}

// ServiceLogger returns logger dedicated to the specific channel. Similar to Named() but also reads the core params.
func (p *Plugin) ServiceLogger(n endure.Named) (*zap.Logger, error) {
	return p.namedLogger(n.Name())
}

// Provides declares factory methods.
func (p *Plugin) Provides() []any {
	return []any{
		p.ServiceLogger,
	}
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

// namedLogger returns logger bound to the specific channel
func (p *Plugin) namedLogger(name string) (*zap.Logger, error) {
	if cfg, ok := p.channels.Channels[name]; ok {
		l, err := cfg.BuildLogger()
		if err != nil {
			return nil, err
		}
		return l.Named(name).With(p.fields()...), nil
	}

	return p.base.Named(name), nil
}

func (p *Plugin) fields() []zap.Field {
	return []zap.Field{
		zap.String("version", p.version),
		zap.String("cmd", p.cmd),
	}
}
