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

func (f *firestoreRepo) Create(ctx context.Context, item repo.BaseEntityType) (id string, err error) {
	const op errs.Op = "firestore.Create"
	item.SetActive(true)

	cll := f.client.Collection(f.cllName)

	var doc *firestore.DocumentRef
	if item.GetId() == "" {
		doc = cll.NewDoc()
		item.SetId(doc.ID)
	} else {
		doc = cll.Doc(item.GetId())
	}

	_, err = doc.Create(ctx, item)
	if err != nil {
		return "", errs.New(err, op)
	}

	return doc.ID, nil
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
	docItr := f.client.Collection(f.cllName).Documents(ctx)

	results := make([]repo.BaseEntityType, 0)
	for {
		shot, err := docItr.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, errs.New(err, op)
		}

		item := newItem()
		//new(repo.BaseEntityType)
		if shot.Exists() {
			_ = shot.DataTo(item)
			results = append(results, item)

		}
	}

	return results, nil
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
