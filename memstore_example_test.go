package memstore_test

import (
	"context"
	"fmt"

	"github.com/khgame/memstore"
)

type GameUserPackageSlot struct {
	ItemName string
	Quantity int64
}

func (p GameUserPackageSlot) StoreName() string {
	return p.ItemName
}

// Example_InMemStorage the testable example for InMemStorage
func Example_InMemStorage() {
	gameAndUsageSample := "the_rpg_game"
	resourceStore := memstore.NewInMemoryStorage[GameUserPackageSlot](gameAndUsageSample)
	resourceStore.Dumper = createCacheDumper[GameUserPackageSlot]()
	user001, user002 := "user001", "user002"
	err := resourceStore.Set(user001, &GameUserPackageSlot{
		ItemName: "item001",
		Quantity: 1,
	})
	if err != nil {
		panic(err)
	}
	err = resourceStore.Set(user001, &GameUserPackageSlot{
		ItemName: "item002",
		Quantity: 2,
	})
	if err != nil {
		panic(err)
	}
	err = resourceStore.Set(user002, &GameUserPackageSlot{
		ItemName: "item001",
		Quantity: 100,
	})
	if err != nil {
		panic(err)
	}
	resourceStore.Save(context.Background())
	resourceStore2 := memstore.NewInMemoryStorage[GameUserPackageSlot](gameAndUsageSample)
	resourceStore2.Dumper = resourceStore.Dumper
	resourceStore2.Load(context.Background())

	item001, item002 := &GameUserPackageSlot{ItemName: "item001"}, &GameUserPackageSlot{ItemName: "item002"}
	if err = resourceStore2.Get(user001, item001); err != nil {
		panic(err)
	}
	fmt.Println(item001.Quantity)
	if err = resourceStore2.Get(user002, item001); err != nil {
		panic(err)
	}
	fmt.Println(item001.Quantity)
	if err = resourceStore2.Get(user001, item002); err != nil {
		panic(err)
	}
	fmt.Println(item002.Quantity)
	if err = resourceStore2.Get(user002, item002); err != nil {
		panic(err)
	}
	fmt.Println(item002.Quantity)

	// Output:
	// 1
	// 100
	// 2
	// 0
}
