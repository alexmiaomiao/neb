package main

import (
        "fmt"
        "log"
	"time"
        "sync"

        "github.com/nebulasio/go-nebulas/storage"
        "github.com/nebulasio/go-nebulas/core"
        "github.com/nebulasio/go-nebulas/util/byteutils"
        "github.com/gogo/protobuf/proto"
        "github.com/nebulasio/go-nebulas/core/pb"
//        "github.com/nebulasio/go-nebulas/rpc"
//	"github.com/nebulasio/go-nebulas/rpc/pb"
)

const (
        //mainnet数据库路径
        dbpath = "/home/lzt/Documents/mygo/src/github.com/alexmiaomiao/neb/data.db"
	maxRoutine = 5000
	rpcaddr = "13.251.33.39:8684"
)

//获取整个区块链上所有交易单数量
func getTotalTXs(db *storage.RocksStorage) (uint64, error) {
        var currentH uint64
        txChan := make(chan int)
	sema := make(chan struct{}, maxRoutine)
	res := make(chan uint64)
        var wg sync.WaitGroup

        if block, err := getTailBlock(db); err != nil {
                return 0, err
        } else {
                currentH = block.Height()
		fmt.Println(currentH)
        }

	wg.Add(int(currentH))
	go func() {
                wg.Wait()
                close(txChan)
        }()

	go func() {
		var total uint64
		for num := range txChan {
			total += uint64(num)
		}
		res <- total
	}()


        for i := uint64(1); i <= currentH; i++ {
		sema <- struct{}{}
                go func(h uint64) {
                        defer wg.Done()

			b, err := getBlockByHeight(h, db)
			if err != nil {
				log.Println(err)
				log.Println(h)
				txChan <- 0
			} else {
				txChan <- len(b.Transactions())

			}
			<- sema
                }(i)
        }

        return  <- res, nil
}

//func rpcGetBlockByHash(hash []byte) (*rpcpb.)

//通过hash获取某个区块
func getBlockByHash(hash []byte, db *storage.RocksStorage) (*core.Block, error) {
        value, err := db.Get(hash)
	if err != nil {
                return nil, err
        } 
	pbBlock := new(corepb.Block)
	block := new(core.Block)
	if err = proto.Unmarshal(value, pbBlock); err != nil {
		return nil, err
	}
	fmt.Println(pbBlock.GetHeader())
	//fmt.Println(pbBlock)
	if err = block.FromProto(pbBlock); err != nil {
		fmt.Println(block)
		return nil, err
	}
	return block, nil
}

//通过高度得到某个区块
func getBlockByHeight(h uint64, db *storage.RocksStorage) (*core.Block, error) {
	hash, err := db.Get(byteutils.FromUint64(h))
	if err != nil {
                return nil, err
        }
	return getBlockByHash(hash, db)
}

//得到最后一个区块
func getTailBlock(db *storage.RocksStorage) (*core.Block, error) {
        if hash, err := db.Get([]byte(core.Tail)); err != nil {
                return nil, err
        } else {
                if b, err := getBlockByHash(hash, db); err != nil {
                        return nil, err
                } else {
                        return b, nil
		}
        }
}

func heightInPeriod(ts, te time.Time, hs, he uint64, db *storage.RocksStorage) (uint64, error) {
	hm := (hs + he) / 2
        block, err := getBlockByHeight(hm, db)
	if err != nil {
		fmt.Println(hm, err)
                return 0, err
        }
	bt := block.Timestamp()
	if bt < ts.Unix() {
		return heightInPeriod(ts, te, hm+1, he, db)
	}
	if bt > te.Unix() {
		return heightInPeriod(ts, te, hs, hm-1, db)
	}
	return hm, nil
}

func getTXsByDay(y,m,d int, db *storage.RocksStorage) (*core.Transactions, error) {
	ts := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
	te := ts.Add(24*time.Hour)
	
        block, err := getTailBlock(db)
	if err != nil {
                return nil, err
        }
	
	hh, err := heightInPeriod(ts, te, 2, block.Height(), db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hh)

	return nil, nil
}

func main() {
        db, err := storage.NewRocksStorage(dbpath)

        if err != nil {
                fmt.Println(dbpath)
                log.Fatal(err)
        }
        defer db.Close()

	//224013
        if block, err := getBlockByHeight(uint64(224013), db); err != nil {
                log.Fatal(err)
        } else {
                fmt.Println(block)
                for _, tx := range block.Transactions() {
                        fmt.Println(tx)
                }
        }

        // if num, err := getTotalTXs(db); err != nil {
        //         log.Fatal(err)
	// } else {
        //         fmt.Println(num)
        // }

	//getTXsByDay(2018, 5, 5, db)
}
