package query

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/zitadel/zitadel/internal/domain"
)

var (
	prepareUserAuthMethodsStmt = `SELECT projections.user_auth_methods4.token_id,` +
		` projections.user_auth_methods4.creation_date,` +
		` projections.user_auth_methods4.change_date,` +
		` projections.user_auth_methods4.resource_owner,` +
		` projections.user_auth_methods4.user_id,` +
		` projections.user_auth_methods4.sequence,` +
		` projections.user_auth_methods4.name,` +
		` projections.user_auth_methods4.state,` +
		` projections.user_auth_methods4.method_type,` +
		` COUNT(*) OVER ()` +
		` FROM projections.user_auth_methods4` +
		` AS OF SYSTEM TIME '-1 ms'`
	prepareUserAuthMethodsCols = []string{
		"token_id",
		"creation_date",
		"change_date",
		"resource_owner",
		"user_id",
		"sequence",
		"name",
		"state",
		"method_type",
		"count",
	}
	prepareActiveAuthMethodTypesStmt = `SELECT projections.users8_notifications.password_set,` +
		` auth_method_types.method_type,` +
		` user_idps_count.count` +
		` FROM projections.users8` +
		` LEFT JOIN projections.users8_notifications ON projections.users8.id = projections.users8_notifications.user_id AND projections.users8.instance_id = projections.users8_notifications.instance_id` +
		` LEFT JOIN (SELECT DISTINCT(auth_method_types.method_type), auth_method_types.user_id, auth_method_types.instance_id FROM projections.user_auth_methods4 AS auth_method_types` +
		` WHERE auth_method_types.state = $1) AS auth_method_types` +
		` ON auth_method_types.user_id = projections.users8.id AND auth_method_types.instance_id = projections.users8.instance_id` +
		` LEFT JOIN (SELECT user_idps_count.user_id, user_idps_count.instance_id, COUNT(user_idps_count.user_id) AS count FROM projections.idp_user_links3 AS user_idps_count` +
		` GROUP BY user_idps_count.user_id, user_idps_count.instance_id) AS user_idps_count` +
		` ON user_idps_count.user_id = projections.users8.id AND user_idps_count.instance_id = projections.users8.instance_id` +
		` AS OF SYSTEM TIME '-1 ms`
	prepareActiveAuthMethodTypesCols = []string{
		"password_set",
		"method_type",
		"idps_count",
	}
)

func Test_UserAuthMethodPrepares(t *testing.T) {
	type want struct {
		sqlExpectations sqlExpectation
		err             checkErr
	}
	tests := []struct {
		name    string
		prepare interface{}
		want    want
		object  interface{}
	}{
		{
			name:    "prepareUserAuthMethodsQuery no result",
			prepare: prepareUserAuthMethodsQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareUserAuthMethodsStmt),
					nil,
					nil,
				),
			},
			object: &AuthMethods{AuthMethods: []*AuthMethod{}},
		},
		{
			name:    "prepareUserAuthMethodsQuery one result",
			prepare: prepareUserAuthMethodsQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareUserAuthMethodsStmt),
					prepareUserAuthMethodsCols,
					[][]driver.Value{
						{
							"token_id",
							testNow,
							testNow,
							"ro",
							"user_id",
							uint64(20211108),
							"name",
							domain.MFAStateReady,
							domain.UserAuthMethodTypeU2F,
						},
					},
				),
			},
			object: &AuthMethods{
				SearchResponse: SearchResponse{
					Count: 1,
				},
				AuthMethods: []*AuthMethod{
					{
						TokenID:       "token_id",
						CreationDate:  testNow,
						ChangeDate:    testNow,
						ResourceOwner: "ro",
						UserID:        "user_id",
						Sequence:      20211108,
						Name:          "name",
						State:         domain.MFAStateReady,
						Type:          domain.UserAuthMethodTypeU2F,
					},
				},
			},
		},
		{
			name:    "prepareUserAuthMethodsQuery multiple result",
			prepare: prepareUserAuthMethodsQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareUserAuthMethodsStmt),
					prepareUserAuthMethodsCols,
					[][]driver.Value{
						{
							"token_id",
							testNow,
							testNow,
							"ro",
							"user_id",
							uint64(20211108),
							"name",
							domain.MFAStateReady,
							domain.UserAuthMethodTypeU2F,
						},
						{
							"token_id-2",
							testNow,
							testNow,
							"ro",
							"user_id",
							uint64(20211108),
							"name-2",
							domain.MFAStateReady,
							domain.UserAuthMethodTypePasswordless,
						},
					},
				),
			},
			object: &AuthMethods{
				SearchResponse: SearchResponse{
					Count: 2,
				},
				AuthMethods: []*AuthMethod{
					{
						TokenID:       "token_id",
						CreationDate:  testNow,
						ChangeDate:    testNow,
						ResourceOwner: "ro",
						UserID:        "user_id",
						Sequence:      20211108,
						Name:          "name",
						State:         domain.MFAStateReady,
						Type:          domain.UserAuthMethodTypeU2F,
					},
					{
						TokenID:       "token_id-2",
						CreationDate:  testNow,
						ChangeDate:    testNow,
						ResourceOwner: "ro",
						UserID:        "user_id",
						Sequence:      20211108,
						Name:          "name-2",
						State:         domain.MFAStateReady,
						Type:          domain.UserAuthMethodTypePasswordless,
					},
				},
			},
		},
		{
			name:    "prepareUserAuthMethodsQuery sql err",
			prepare: prepareUserAuthMethodsQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(prepareUserAuthMethodsStmt),
					sql.ErrConnDone,
				),
				err: func(err error) (error, bool) {
					if !errors.Is(err, sql.ErrConnDone) {
						return fmt.Errorf("err should be sql.ErrConnDone got: %w", err), false
					}
					return nil, true
				},
			},
			object: nil,
		},
		{
			name:    "prepareActiveUserAuthMethodTypesQuery no result",
			prepare: prepareActiveUserAuthMethodTypesQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareActiveAuthMethodTypesStmt),
					nil,
					nil,
				),
			},
			object: &AuthMethodTypes{AuthMethodTypes: []domain.UserAuthMethodType{}},
		},
		{
			name:    "prepareActiveUserAuthMethodTypesQuery one second factor",
			prepare: prepareActiveUserAuthMethodTypesQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareActiveAuthMethodTypesStmt),
					prepareActiveAuthMethodTypesCols,
					[][]driver.Value{
						{
							true,
							domain.UserAuthMethodTypePasswordless,
							1,
						},
					},
				),
			},
			object: &AuthMethodTypes{
				SearchResponse: SearchResponse{
					Count: 3,
				},
				AuthMethodTypes: []domain.UserAuthMethodType{
					domain.UserAuthMethodTypePasswordless,
					domain.UserAuthMethodTypePassword,
					domain.UserAuthMethodTypeIDP,
				},
			},
		},
		{
			name:    "prepareActiveUserAuthMethodTypesQuery multiple second factors",
			prepare: prepareActiveUserAuthMethodTypesQuery,
			want: want{
				sqlExpectations: mockQueries(
					regexp.QuoteMeta(prepareActiveAuthMethodTypesStmt),
					prepareActiveAuthMethodTypesCols,
					[][]driver.Value{
						{
							true,
							domain.UserAuthMethodTypePasswordless,
							1,
						},
						{
							true,
							domain.UserAuthMethodTypeOTP,
							1,
						},
					},
				),
			},
			object: &AuthMethodTypes{
				SearchResponse: SearchResponse{
					Count: 4,
				},
				AuthMethodTypes: []domain.UserAuthMethodType{
					domain.UserAuthMethodTypePasswordless,
					domain.UserAuthMethodTypeOTP,
					domain.UserAuthMethodTypePassword,
					domain.UserAuthMethodTypeIDP,
				},
			},
		},
		{
			name:    "prepareActiveUserAuthMethodTypesQuery sql err",
			prepare: prepareActiveUserAuthMethodTypesQuery,
			want: want{
				sqlExpectations: mockQueryErr(
					regexp.QuoteMeta(prepareActiveAuthMethodTypesStmt),
					sql.ErrConnDone,
				),
				err: func(err error) (error, bool) {
					if !errors.Is(err, sql.ErrConnDone) {
						return fmt.Errorf("err should be sql.ErrConnDone got: %w", err), false
					}
					return nil, true
				},
			},
			object: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPrepare(t, tt.prepare, tt.object, tt.want.sqlExpectations, tt.want.err, defaultPrepareArgs...)
		})
	}
}
