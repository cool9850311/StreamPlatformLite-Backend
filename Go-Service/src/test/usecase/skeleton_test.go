// Go-Service/src/test/usecase/skeleton_test.go
package usecase_test

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity"
	errors_enum "Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/entity/role"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockSkeletonRepository struct {
	skeletons map[string]*entity.Skeleton
}

func (m *MockSkeletonRepository) GetByID(id string) (*entity.Skeleton, error) {
	if skeleton, exists := m.skeletons[id]; exists {
		return skeleton, nil
	}
	return nil, errors.New("skeleton not found")
}

func (m *MockSkeletonRepository) Create(skeleton *entity.Skeleton) error {
	if _, exists := m.skeletons[skeleton.ID]; exists {
		return errors.New("skeleton already exists")
	}
	m.skeletons[skeleton.ID] = skeleton
	return nil
}

// MockLogger implementation
type MockLogger struct{}

func (m *MockLogger) Panic(ctx context.Context, msg string) {}
func (m *MockLogger) Fatal(ctx context.Context, msg string) {}
func (m *MockLogger) Error(ctx context.Context, msg string) {}
func (m *MockLogger) Warn(ctx context.Context, msg string)  {}
func (m *MockLogger) Info(ctx context.Context, msg string)  {}
func (m *MockLogger) Debug(ctx context.Context, msg string) {}
func (m *MockLogger) Trace(ctx context.Context, msg string) {}

func TestSkeletonUseCase_GetSkeletonByID(t *testing.T) {
	repo := &MockSkeletonRepository{
		skeletons: map[string]*entity.Skeleton{
			"1": {ID: "1", Name: "Test Skeleton"},
		},
	}
	logger := &MockLogger{}
	uc := usecase.SkeletonUseCase{
		SkeletonRepo: repo,
		Log:          logger,
	}

	ctx := context.Background()
	skeleton, err := uc.GetSkeletonByID(ctx, "1", role.User)
	assert.Nil(t, err)
	assert.NotNil(t, skeleton)
	assert.Equal(t, "Test Skeleton", skeleton.Name)

	skeleton, err = uc.GetSkeletonByID(ctx, "2", role.User)
	assert.NotNil(t, err)
	assert.Nil(t, skeleton)
}

func TestSkeletonUseCase_GetSkeletonByIDFail(t *testing.T) {
	repo := &MockSkeletonRepository{
		skeletons: map[string]*entity.Skeleton{
			"1": {ID: "1", Name: "Test Skeleton"},
		},
	}
	logger := &MockLogger{}
	uc := usecase.SkeletonUseCase{
		SkeletonRepo: repo,
		Log:          logger,
	}

	ctx := context.Background()
	skeleton, err := uc.GetSkeletonByID(ctx, "1", role.Guest)
	assert.ErrorIs(t, err, errors_enum.ErrUnauthorized)
	assert.Nil(t, skeleton)
}

func TestSkeletonUseCase_CreateSkeleton(t *testing.T) {
	repo := &MockSkeletonRepository{skeletons: make(map[string]*entity.Skeleton)}
	logger := &MockLogger{}
	uc := usecase.SkeletonUseCase{
		SkeletonRepo: repo,
		Log:          logger,
	}

	ctx := context.Background()
	newSkeleton := &entity.Skeleton{ID: "1", Name: "New Skeleton"}
	err := uc.CreateSkeleton(ctx, newSkeleton, role.User)
	assert.Nil(t, err)

	skeleton, err := uc.GetSkeletonByID(ctx, "1", role.User)
	assert.Nil(t, err)
	assert.NotNil(t, skeleton)
	assert.Equal(t, "New Skeleton", skeleton.Name)

	err = uc.CreateSkeleton(ctx, newSkeleton, role.User)
	assert.NotNil(t, err)
}

func TestSkeletonUseCase_CreateSkeletonFail(t *testing.T) {
	repo := &MockSkeletonRepository{skeletons: make(map[string]*entity.Skeleton)}
	logger := &MockLogger{}
	uc := usecase.SkeletonUseCase{
		SkeletonRepo: repo,
		Log:          logger,
	}

	ctx := context.Background()
	newSkeleton := &entity.Skeleton{ID: "1", Name: "New Skeleton"}
	err := uc.CreateSkeleton(ctx, newSkeleton, role.Guest)
	assert.ErrorIs(t, err, errors_enum.ErrUnauthorized)
}
