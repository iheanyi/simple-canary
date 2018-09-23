//go:generate gorunpkg github.com/99designs/gqlgen

package app

import (
	context "context"
	time "time"

	dbpkg "github.com/iheanyi/simple-canary/internal/db"
)

type Resolver struct {
	db dbpkg.CanaryStore
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) TestInstance() TestInstanceResolver {
	return &testInstanceResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Tests(ctx context.Context) ([]dbpkg.TestInstance, error) {
	tests, err := r.db.ListTests()
	return tests, err
}

func (r *queryResolver) Test(ctx context.Context, id string) (*dbpkg.TestInstance, error) {
	test, err := r.db.FindTestByID(id)
	return test, err
}

func (r *queryResolver) OngoingTests(ctx context.Context) ([]dbpkg.TestInstance, error) {
	tests, err := r.db.ListOngoingTests()
	return tests, err
}

type testInstanceResolver struct{ *Resolver }

func (r *testInstanceResolver) ID(ctx context.Context, obj *dbpkg.TestInstance) (string, error) {
	return obj.TestID, nil
}
func (r *testInstanceResolver) Name(ctx context.Context, obj *dbpkg.TestInstance) (string, error) {
	return obj.TestName, nil
}
func (r *testInstanceResolver) StartAt(ctx context.Context, obj *dbpkg.TestInstance) (time.Time, error) {
	return obj.StartAt, nil
}
func (r *testInstanceResolver) EndAt(ctx context.Context, obj *dbpkg.TestInstance) (*time.Time, error) {
	return &obj.EndAt, nil
}
func (r *testInstanceResolver) FailCause(ctx context.Context, obj *dbpkg.TestInstance) (*string, error) {
	return &obj.FailCause, nil
}
