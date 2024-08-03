// Go-Service/src/test/usecase/skeleton_usecase_test.go
package usecase_test

import (
	"Go-Service/src/main/application/usecase"
	"Go-Service/src/main/domain/entity"
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

func TestSkeletonUseCase_GetSkeletonByID(t *testing.T) {
	repo := &MockSkeletonRepository{
		skeletons: map[string]*entity.Skeleton{
			"1": {ID: "1", Name: "Test Skeleton"},
		},
	}
	uc := usecase.SkeletonUseCase{SkeletonRepo: repo}

	skeleton, err := uc.GetSkeletonByID("1")
	assert.Nil(t, err)
	assert.NotNil(t, skeleton)
	assert.Equal(t, "Test Skeleton", skeleton.Name)

	skeleton, err = uc.GetSkeletonByID("2")
	assert.NotNil(t, err)
	assert.Nil(t, skeleton)
}

func TestSkeletonUseCase_CreateSkeleton(t *testing.T) {
	repo := &MockSkeletonRepository{skeletons: make(map[string]*entity.Skeleton)}
	uc := usecase.SkeletonUseCase{SkeletonRepo: repo}

	newSkeleton := &entity.Skeleton{ID: "1", Name: "New Skeleton"}
	err := uc.CreateSkeleton(newSkeleton)
	assert.Nil(t, err)

	skeleton, err := uc.GetSkeletonByID("1")
	assert.Nil(t, err)
	assert.NotNil(t, skeleton)
	assert.Equal(t, "New Skeleton", skeleton.Name)

	err = uc.CreateSkeleton(newSkeleton)
	assert.NotNil(t, err)
}
