package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"github.com/maohieng/go-firebase"
	"github.com/maohieng/go-repo"
	"log"
	"os"
	"testing"
)

type Menu struct {
	repo.BaseEntity
	Name string `firestore:"name" json:"name"`
}

var (
	ctx      context.Context
	crudrepo *firestoreRepo
)

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx = context.Background()
		fApp := firebase.InitAppDefault(ctx)
		fStore, err := fApp.Firestore(ctx)
		if err != nil {
			log.Println("cmd", "init", "firebase", "firestore", "err", err)
			os.Exit(1)
		}
		defer func(client *firestore.Client) {
			err := client.Close()
			if err != nil {
				log.Println("cmd", "exit", "closing", "firestore", "err", err)
			}
		}(fStore)

		crudrepo = &firestoreRepo{client: fStore, cllName: "test_menus"}

		return m.Run()
	}())
}

func TestCreate(t *testing.T) {
	id, err := crudrepo.Create(ctx, &Menu{
		BaseEntity: repo.BaseEntity{
			Active: true,
			Id:     "test_id_1",
		},
		Name: "test 1",
	})

	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	t.Logf("Created: %s", id)
}

func TestUpdate(t *testing.T) {
	fv := make(map[string]any, 0)
	fv["name"] = "Update test again"
	err := crudrepo.Update(ctx, "WjnvkMFKfN7I4UbK5NQM", fv)
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}
}

func TestGetOne(t *testing.T) {
	item := &Menu{}
	err := crudrepo.GetOne(ctx, "WjnvkMFKfN7I4UbK5NQM", item)
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	itemJson, _ := json.Marshal(item)
	t.Logf("GetOne: %s", string(itemJson))
}

func TestGetAll(t *testing.T) {
	results, err := crudrepo.GetAll(ctx, func() repo.BaseEntityType {
		return &Menu{}
	})

	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	itemJson, _ := json.Marshal(results)
	t.Logf("GetAll: %s", string(itemJson))
}

func TestCreateAll(t *testing.T) {
	menus := []*Menu{
		{
			BaseEntity: repo.BaseEntity{
				Active: true,
				Id:     "",
			},
			Name: "all in 1",
		},
		{
			BaseEntity: repo.BaseEntity{
				Active: false,
				Id:     "",
			},
			Name: "all in 2",
		},
	}

	entities := make([]repo.BaseEntityType, 0, len(menus))
	for _, menu := range menus {
		entities = append(entities, menu)
	}

	ids, err := crudrepo.CreateAll(ctx, entities)
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	t.Logf("CreateAll: %v", ids)
}

func TestPaginate(t *testing.T) {
	page, err := crudrepo.Paginate(ctx, 3, "", func() repo.BaseEntityType {
		return &Menu{}
	})
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	itemJson, _ := json.Marshal(page)
	t.Logf("Page 1: %s", string(itemJson))

	page, err = crudrepo.Paginate(ctx, 3, page.NextToken, func() repo.BaseEntityType {
		return &Menu{}
	})
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	itemJson, _ = json.Marshal(page)
	t.Logf("Page 2: %s", string(itemJson))

	page, err = crudrepo.Paginate(ctx, 3, page.NextToken, func() repo.BaseEntityType {
		return &Menu{}
	})
	if err != nil {
		t.Fatalf("Expected success, got %v", err)
	}

	itemJson, _ = json.Marshal(page)
	t.Logf("Page 3: %s", string(itemJson))
}

//func TestDelete(t *testing.T) {
//	err := crudrepo.Delete(ctx, "KAXKJcF6ou12SNiv7028")
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//}
//
//func TestSoftDelete(t *testing.T) {
//	err := crudrepo.SoftDelete(ctx, "mMibQfBSB5NVJVyvFvDG")
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//}
