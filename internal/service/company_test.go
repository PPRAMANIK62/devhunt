package service_test

import (
	"context"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository/mocks"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompanyService_Create_AlreadyExists(t *testing.T) {
	repo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "existing", UserID: userID}, nil
		},
	}
	svc := service.NewCompanyService(repo)

	_, err := svc.Create(context.Background(), "user-1", models.CreateCompanyRequest{
		Name: "Acme",
		Slug: "acme",
	})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeConflict, appErr.Type)
}

func TestCompanyService_Create_Success(t *testing.T) {
	repo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return nil, apperr.NotFound("company")
		},
		CreateFn: func(ctx context.Context, company *models.Company) error {
			company.ID = "company-new"
			return nil
		},
	}
	svc := service.NewCompanyService(repo)

	company, err := svc.Create(context.Background(), "user-1", models.CreateCompanyRequest{
		Name: "Acme",
		Slug: "acme",
	})

	require.NoError(t, err)
	assert.Equal(t, "company-new", company.ID)
	assert.Equal(t, "user-1", company.UserID)
	assert.Equal(t, "acme", company.Slug)
}

func TestCompanyService_GetMine_NotFound(t *testing.T) {
	repo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return nil, apperr.NotFound("company")
		},
	}
	svc := service.NewCompanyService(repo)

	_, err := svc.GetMine(context.Background(), "user-no-company")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestCompanyService_GetByID_NotFound(t *testing.T) {
	repo := &mocks.CompanyRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Company, error) {
			return nil, apperr.NotFound("company")
		},
	}
	svc := service.NewCompanyService(repo)

	_, err := svc.GetByID(context.Background(), "missing-id")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestCompanyService_Update_Success(t *testing.T) {
	name := "Updated"
	slug := "updated"
	var capturedFields map[string]any

	repo := &mocks.CompanyRepo{
		UpdateFn: func(ctx context.Context, userID string, fields map[string]any) (*models.Company, error) {
			capturedFields = fields
			return &models.Company{ID: "c1", Name: "Updated", Slug: "updated"}, nil
		},
	}
	svc := service.NewCompanyService(repo)

	company, err := svc.Update(context.Background(), "user-1", models.UpdateCompanyRequest{
		Name: &name,
		Slug: &slug,
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated", company.Name)
	assert.Equal(t, "Updated", capturedFields["name"])
	assert.Equal(t, "updated", capturedFields["slug"])
	// Description and Website not provided — should not appear in fields
	_, hasDesc := capturedFields["description"]
	assert.False(t, hasDesc)
}

func TestCompanyService_Delete_Success(t *testing.T) {
	deleted := false

	repo := &mocks.CompanyRepo{
		DeleteFn: func(ctx context.Context, userID string) error {
			deleted = true
			return nil
		},
	}
	svc := service.NewCompanyService(repo)

	err := svc.Delete(context.Background(), "user-1")

	require.NoError(t, err)
	assert.True(t, deleted)
}
