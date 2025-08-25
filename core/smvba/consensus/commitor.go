package consensus

import (
	"bft/mvba/logger"
	"bft/mvba/mempool"
)

type Committor struct {
	Mempool  *mempool.Mempool
	Index    int64
	Blocks   map[int64]*ConsensusBlock
	commitCh chan *ConsensusBlock
	callBack chan<- struct{}
}

func NewCommittor(callBack chan<- struct{}, pool *mempool.Mempool) *Committor {
	c := &Committor{
		Mempool:  pool,
		Index:    0,
		Blocks:   map[int64]*ConsensusBlock{},
		commitCh: make(chan *ConsensusBlock),
		callBack: callBack,
	}
	go c.run()
	return c
}

func (c *Committor) Commit(block *ConsensusBlock) {
	if block.Epoch < c.Index {
		return
	}
	c.Blocks[block.Epoch] = block
	for {
		if b, ok := c.Blocks[c.Index]; ok {
			c.commitCh <- b
			delete(c.Blocks, c.Index)
			c.Index++
		} else {
			break
		}
	}
}

func (c *Committor) run() {
	for block := range c.commitCh {
		//logger.Info.Printf("commit ConsensusBlock epoch %d node %d the length of the payload is %d\n", block.Epoch, block.Proposer, len(block.PayLoads))
		for _, payload := range block.PayLoads {
			if smallblock, err := c.Mempool.GetBlock(payload); err == nil {
				if smallblock.Batch.Txs != nil {
					logger.Info.Printf("commit Block node %d batch_id %d \n", smallblock.Proposer, smallblock.Batch.ID)
				}
			} else {
				//阻塞提交，等待收到payload
				logger.Error.Printf("get key error\n")
			}
		}
		logger.Info.Printf("commit ConsensusBlock epoch %d node %d\n", block.Epoch, block.Proposer)
		c.callBack <- struct{}{}
	}
}
