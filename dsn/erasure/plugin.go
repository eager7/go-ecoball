package erasure

import (
	"github.com/ipfs/go-ipfs/core/coredag"
	"github.com/ipfs/go-ipfs/plugin"
	//"gx/ipfs/QmYVNvtQkeZ6AKSwDrjQTs432QtL6umrrK41EBq3cu7iSP/go-cid"
	"gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
)

// Plugins is exported list of plugins that will be loaded
var Plugins = []plugin.Plugin{
	&erasurePlugin{},
}

type erasurePlugin struct{}

var _ plugin.PluginIPLD = (*erasurePlugin)(nil)

func (*erasurePlugin) Name() string {
	return "ecoball-erasure"
}

func (*erasurePlugin) Version() string {
	return "0.0.1"
}

func (*erasurePlugin) Init() error {
	return nil
}

func (*erasurePlugin) RegisterBlockDecoders(dec format.BlockDecoder) error {
	//dec.Register(cid.Erasure, DecodeProtobufBlock)
	return nil
}

func (*erasurePlugin) RegisterInputEncParsers(iec coredag.InputEncParsers) error {
	return nil
}
