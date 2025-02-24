package engines

import (
	"github.com/bububa/atomic-agents/components/vectordb/engines/chromem"
	"github.com/bububa/atomic-agents/components/vectordb/engines/memory"
	"github.com/bububa/atomic-agents/components/vectordb/engines/milvus"
)

var (
	FromChromem = chromem.New
	FromMemory  = memory.New
	FromMilvus  = milvus.New
)
