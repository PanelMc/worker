package io

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func hclEnconde(cfg interface{}) []byte {
	block := gohcl.EncodeAsBlock(cfg, "")

	f := hclwrite.NewEmptyFile()
	*f.Body() = *block.Body()

	return f.Bytes()
}

func hclDecode(src []byte, cfg interface{}) error {
	return hclsimple.Decode("TODO.hcl", src, nil, cfg)
}
