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
)

const (
        //mainnet数据库路径
        dbpath = "/home/lzt/Documents/mygo/src/github.com/lzt/neb/data.db"
	maxRoutine = 10000
)

//获取整个区块链上所有交易单数量
func getTotalTXs(db *storage.RocksStorage) (uint64, error) {
        var currentH uint64
        txChan := make(chan uint64)
	sema := make(chan struct{}, maxRoutine)
        var wg sync.WaitGroup

        if block, err := getTailBlock(db); err != nil {
                return 0, err
        } else {
                currentH = block.Height()
		fmt.Println(currentH)
        }

        for i := uint64(1); i <= currentH; i++ {
                wg.Add(1)
		sema <- struct{}{}
                go func(h uint64) {
                        defer wg.Done()

			b, err := getBlockByHeight(h, db)
			if err != nil {
				log.Println(err)
				log.Println(h)
				txChan <- 0
			} else {
				txChan <- uint64(len(b.Transactions()))

			}
			<- sema
                }(i)
        }

        go func() {
                wg.Wait()
                close(txChan)
        }()

        var total uint64
        for num := range txChan {
                total += num
        }

        return total, nil
}

//通过hash获取某个区块
func getBlockByHash(hash []byte, db *storage.RocksStorage) (*core.Block, error) {
        if b, err := db.Get(hash); err != nil {
                return nil, err
        } else {
                pbBlock := new(corepb.Block)
                block := new(core.Block)
                if err := proto.Unmarshal(b, pbBlock); err != nil {
			return nil, err
                }
                if err := block.FromProto(pbBlock); err != nil {
                        return nil, err
                }
                return block, nil
        }
}

//通过高度得到某个区块
func getBlockByHeight(h uint64, db *storage.RocksStorage) (*core.Block, error) {
	if hash, err := db.Get(byteutils.FromUint64(h)); err != nil {
                return nil, err
        } else {
                if b, err := getBlockByHash(hash, db); err != nil {
                        return nil, err
                } else {
                        return b, nil
                }
        }
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

func getTXsByDay(y,m,d int, db *storage.RocksStorage) (*core.Transactions, error) {
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
	fmt.Println(t.Unix())
	b, _ := getTailBlock(db)
	fmt.Println(time.Unix(b.Timestamp(), int64(0)))
	return nil, nil
}

func main() {
        db, err := storage.NewRocksStorage(dbpath)

        if err != nil {
                fmt.Println(dbpath)
                log.Fatal(err)
        }
        defer db.Close()

        // if block, err := getBlockByHeight(uint64(256977), db); err != nil {
        //         log.Fatal(err)
        // } else {
        //         fmt.Println(block)
        //         for _, tx := range block.Transactions() {
        //                 fmt.Println(tx)
        //         }
        // }

        // if num, err := getTotalTXs(db); err != nil {
        //         log.Fatal(err)
	// } else {
        //         fmt.Println(num)
        // }

	getTXsByDay(2018, 5, 5, db)
}
