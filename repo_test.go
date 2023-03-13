package firestore

//
//type Menu struct {
//	repo.BaseEntity
//	Name string `firestore:"name" json:"name"`
//}
//
//var (
//	ctx  context.Context
//	repo repo.CRUDRepository
//)
//
//func TestMain(m *testing.M) {
//	os.Exit(func() int {
//		ctx = context.Background()
//		fApp := firebase.InitAppDefault(ctx)
//		fStore, err := fApp.Firestore(ctx)
//		if err != nil {
//			log.Println("cmd", "init", "firebase", "firestore", "err", err)
//			os.Exit(1)
//		}
//		defer func(client *firestore.Client) {
//			err := client.Close()
//			if err != nil {
//				log.Println("cmd", "exit", "closing", "firestore", "err", err)
//			}
//		}(fStore)
//
//		repo = NewFirestoreRepository(fStore, "test_menus")
//
//		return m.Run()
//	}())
//}
//
//func TestCreate(t *testing.T) {
//	id, err := repo.Create(ctx, &Menu{
//		BaseEntity: entity.BaseEntity{
//			Active: true,
//			Id:     "",
//		},
//		Name: "Test menu 2",
//	})
//
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//
//	t.Logf("Created: %s", id)
//}
//
//func TestUpdate(t *testing.T) {
//	fv := make(map[string]any, 0)
//	fv["name"] = "Update test"
//	err := repo.Update(ctx, "KAXKJcF6ou12SNiv7028", fv)
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//}
//
//func TestGetOne(t *testing.T) {
//	item := &Menu{}
//	err := repo.GetOne(ctx, "KAXKJcF6ou12SNiv7028", item)
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//
//	itemJson, _ := json.Marshal(item)
//	t.Logf("GetOne: %s", string(itemJson))
//}
//
//func TestGetAll(t *testing.T) {
//	results, err := repo.GetAll(ctx, func() entity.BaseEntityType {
//		return &Menu{}
//	})
//
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//
//	itemJson, _ := json.Marshal(results)
//	t.Logf("GetOne: %s", string(itemJson))
//}
//
//func TestDelete(t *testing.T) {
//	err := repo.Delete(ctx, "KAXKJcF6ou12SNiv7028")
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//}
//
//func TestSoftDelete(t *testing.T) {
//	err := repo.SoftDelete(ctx, "mMibQfBSB5NVJVyvFvDG")
//	if err != nil {
//		t.Fatalf("Expected success, got %v", err)
//	}
//}
