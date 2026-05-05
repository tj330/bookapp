package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tj330/bookapp/metadata/internal/repository"
	"github.com/tj330/bookapp/metadata/pkg/model"

	gen "github.com/tj330/bookapp/gen/mock/metadata/repository"
)

func TestController(t *testing.T) {
	errUnexpected := errors.New("unexpected error")

	tests := []struct {
		name       string
		expRepoRes *model.Metadata
		expRepoErr error
		wantRes    *model.Metadata
		wantErr    error
	}{
		{
			name:       "not-found",
			expRepoErr: repository.ErrNotFound,
			wantErr:    ErrNotFound,
		},
		{
			name:       "unexpected error",
			expRepoErr: errUnexpected,
			wantErr:    errUnexpected,
		},
		{
			name:       "success",
			expRepoRes: &model.Metadata{},
			wantRes:    &model.Metadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := gen.NewMockmetadataRepository(ctrl)
			c := New(repoMock)
			ctx := context.Background()
			id := "id"
			repoMock.EXPECT().Get(ctx, id).Return(tt.expRepoRes, tt.expRepoErr)
			res, err := c.Get(ctx, id)
			assert.Equal(t, tt.wantRes, res, tt.name)
			assert.ErrorIs(t, tt.wantErr, err, tt.name)
		})
	}
}
