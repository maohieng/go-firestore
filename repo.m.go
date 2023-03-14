package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"github.com/maohieng/errs"
	"github.com/maohieng/go-repo"
	"google.golang.org/api/iterator"
)

var (
	ErrNoParentFound = errors.New("no parent collection found")
	ErrNotFound      = errors.New("record not found")
)

func NewFirestoreRepository(client *firestore.Client, cllName string) repo.CRUDRepository {
	return &firestoreRepo{client: client, cllName: cllName}
}

type firestoreRepo struct {
	client  *firestore.Client
	cllName string
}

func getDocRef(cll *firestore.CollectionRef, item repo.BaseEntityType) *firestore.DocumentRef {
	var doc *firestore.DocumentRef
	if item.GetId() == "" {
		doc = cll.NewDoc()
		item.SetId(doc.ID)
	} else {
		doc = cll.Doc(item.GetId())
	}

	return doc
}

func (f *firestoreRepo) Create(ctx context.Context, item repo.BaseEntityType) (id string, err error) {
	const op errs.Op = "firestore.Create"
	item.SetActive(true)

	cll := f.client.Collection(f.cllName)

	var doc *firestore.DocumentRef = getDocRef(cll, item)

	_, err = doc.Create(ctx, item)
	if err != nil {
		return "", errs.New(err, op)
	}

	return doc.ID, nil
}

func (f *firestoreRepo) CreateAll(ctx context.Context, items []repo.BaseEntityType) (ids []string, err error) {
	const op errs.Op = "firestore.CreateAll"
	bw := f.client.BulkWriter(ctx)

	es := make([]error, 0, len(items))
	ids = make([]string, 0, len(items))
	for _, item := range items {
		doc := getDocRef(f.client.Collection(f.cllName), item)
		_, e := bw.Create(doc, item)
		// should join errors
		if e == nil {
			ids = append(ids, doc.ID)
		} else {
			es = append(es, e)
		}
	}

	bw.End()

	if len(es) > 0 {
		return nil, errors.Join(es...)
	}

	return ids, nil
}

func (f *firestoreRepo) Update(ctx context.Context, id string, fv map[string]interface{}) error {
	const op errs.Op = "firestore.Update"
	cll := f.client.Collection(f.cllName)
	if cll == nil {
		return errs.New(ErrNoParentFound, op)
	}

	doc := cll.Doc(id)

	// Check existing
	shot, err := doc.Get(ctx)
	if err != nil {
		return errs.New(err, op)
	}
	if !shot.Exists() {
		return errs.New(ErrNotFound, op)
	}

	updates := make([]firestore.Update, 0, len(fv))
	for k, v := range fv {
		up := firestore.Update{
			Path:  k,
			Value: v,
		}
		updates = append(updates, up)
	}

	_, err = doc.Update(ctx, updates)
	if err != nil {
		return errs.New(err, op)
	}

	return nil
}

func (f *firestoreRepo) GetOne(ctx context.Context, id string, item repo.BaseEntityType) error {
	var op errs.Op = "firestore.GetOne"
	shot, err := f.client.Collection(f.cllName).Doc(id).Get(ctx)
	if err != nil {
		return errs.New(err, op)
	}

	if !shot.Exists() {
		return errs.New(ErrNotFound, op)
	}

	err = shot.DataTo(item)
	if err != nil {
		return errs.New(err, op)
	}
	item.SetId(shot.Ref.ID)

	return nil
}

func (f *firestoreRepo) GetAll(ctx context.Context, newItem func() repo.BaseEntityType) ([]repo.BaseEntityType, error) {
	var op errs.Op = "firestore.GetAll"
	shots, err := f.client.Collection(f.cllName).Documents(ctx).GetAll()
	if err != nil {
		return nil, errs.New(err, op)
	}

	results := make([]repo.BaseEntityType, 0, len(shots))
	for _, shot := range shots {
		item := newItem()
		if shot.Exists() {
			results = append(results, item)
			item.SetId(shot.Ref.ID)
		}
	}
	//for {
	//	shot, err := shotItr.Next()
	//	if err == iterator.Done {
	//		break
	//	}
	//
	//	if err != nil {
	//		return nil, errs.New(err, op)
	//	}
	//
	//	item := newItem()
	//	//new(repo.BaseEntityType)
	//	if shot.Exists() {
	//		_ = shot.DataTo(item)
	//		results = append(results, item)
	//
	//	}
	//}

	return results, nil
}

func (f *firestoreRepo) Paginate(ctx context.Context, limit int, startToken string, newItem func() repo.BaseEntityType) (repo.Page, error) {
	var qry firestore.Query

	// TODO to test
	if startToken == "" {
		qry = f.client.Collection(f.cllName).Limit(limit)
	} else {
		dc, err := f.decodeSnapshot(ctx, startToken)
		if err != nil {
			return repo.Page{}, err
		}

		qry = f.client.Collection(f.cllName).StartAt(dc).Limit(limit)
	}

	all, err := qry.Documents(ctx).GetAll()
	if err != nil {
		return repo.Page{}, err
	}

	var lastShot *firestore.DocumentSnapshot
	results := make([]repo.BaseEntityType, 0, len(all))
	for _, shot := range all {
		item := newItem()
		if shot.Exists() {
			results = append(results, item)
			item.SetId(shot.Ref.ID)
			lastShot = shot
		}
	}

	// check if next is Done
	var next string
	it := f.client.Collection(f.cllName).StartAfter(lastShot).Limit(1).Documents(ctx)
	for {
		shot, e := it.Next()
		if e == iterator.Done {
			break
		}

		next = encodedSnapshot(shot)
	}

	return repo.Page{
		Items:     results,
		NextToken: next,
	}, nil
}

func encodedSnapshot(shot *firestore.DocumentSnapshot) string {
	//TODO serialize it
	// https://bitbucket.org/cammobteamweb/bluetrace-kh-server/src/master/jomutils/src/main/java/github/jommobile/cloud/datastore/ObjectifyUtils.java
	return shot.Ref.ID
}

func (f *firestoreRepo) decodeSnapshot(ctx context.Context, token string) (*firestore.DocumentSnapshot, error) {
	//TODO deserialize it
	// https://bitbucket.org/cammobteamweb/bluetrace-kh-server/src/master/jomutils/src/main/java/github/jommobile/cloud/datastore/ObjectifyUtils.java
	shot, err := f.client.Collection(f.cllName).Doc(token).Get(ctx)
	if err != nil {
		return nil, err
	}

	return shot, nil
}

func (f *firestoreRepo) Delete(ctx context.Context, id string) error {
	var op errs.Op = "firestore.Delete"
	_, err := f.client.Collection(f.cllName).Doc(id).Delete(ctx)
	if err != nil {
		return errs.New(err, op)
	}

	return nil
}

func (f *firestoreRepo) SoftDelete(ctx context.Context, id string) error {
	fv := make(map[string]any)
	fv[repo.ActiveFieldName] = false
	return f.Update(ctx, id, fv)
}
