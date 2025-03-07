package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/boundary/internal/authtoken"
	authtokenStore "github.com/hashicorp/boundary/internal/authtoken/store"
	cred "github.com/hashicorp/boundary/internal/credential"
	"github.com/hashicorp/boundary/internal/credential/vault"
	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/db/timestamp"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/host/static"
	staticStore "github.com/hashicorp/boundary/internal/host/static/store"
	"github.com/hashicorp/boundary/internal/iam"
	iamStore "github.com/hashicorp/boundary/internal/iam/store"
	"github.com/hashicorp/boundary/internal/kms"
	"github.com/hashicorp/boundary/internal/oplog"
	"github.com/hashicorp/boundary/internal/target"
	"github.com/hashicorp/boundary/internal/target/tcp"
	tcpStore "github.com/hashicorp/boundary/internal/target/tcp/store"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRepository_ListSession(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	const testLimit = 10
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	rw := db.New(conn)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms, WithLimit(testLimit))
	require.NoError(t, err)
	composedOf := TestSessionParams(t, conn, wrapper, iamRepo)

	type args struct {
		opt []Option
	}
	tests := []struct {
		name            string
		createCnt       int
		args            args
		wantCnt         int
		wantErr         bool
		withConnections int
	}{
		{
			name:      "no-limit",
			createCnt: repo.defaultLimit + 1,
			args: args{
				opt: []Option{WithLimit(-1)},
			},
			wantCnt: repo.defaultLimit + 1,
			wantErr: false,
		},
		{
			name:      "default-limit",
			createCnt: repo.defaultLimit + 1,
			args:      args{},
			wantCnt:   repo.defaultLimit,
			wantErr:   false,
		},
		{
			name:      "custom-limit",
			createCnt: repo.defaultLimit + 1,
			args: args{
				opt: []Option{WithLimit(3)},
			},
			wantCnt: 3,
			wantErr: false,
		},
		{
			name:      "withScopeId",
			createCnt: repo.defaultLimit + 1,
			args: args{
				opt: []Option{WithScopeIds([]string{composedOf.ScopeId})},
			},
			wantCnt: repo.defaultLimit,
			wantErr: false,
		},
		{
			name:      "bad-withScopeId",
			createCnt: repo.defaultLimit + 1,
			args: args{
				opt: []Option{WithScopeIds([]string{"o_thisIsNotValid"})},
			},
			wantCnt: 0,
			wantErr: false,
		},
		{
			name:            "multiple-connections",
			createCnt:       repo.defaultLimit + 1,
			args:            args{},
			wantCnt:         repo.defaultLimit,
			wantErr:         false,
			withConnections: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
			testSessions := []*Session{}
			for i := 0; i < tt.createCnt; i++ {
				s := TestSession(t, conn, wrapper, composedOf)
				_ = TestState(t, conn, s.PublicId, StatusActive)
				testSessions = append(testSessions, s)
				for i := 0; i < tt.withConnections; i++ {
					_ = TestConnection(t, conn, s.PublicId, "127.0.0.1", 22, "127.0.0.2", 23, "127.0.0.1")
				}
			}
			assert.Equal(tt.createCnt, len(testSessions))
			got, err := repo.ListSessions(context.Background(), tt.args.opt...)
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tt.wantCnt, len(got))
			for i := 0; i < len(got); i++ {
				assert.Equal(tt.withConnections, len(got[i].Connections))
				for _, c := range got[i].Connections {
					assert.Equal("127.0.0.1", c.ClientTcpAddress)
					assert.Equal(uint32(22), c.ClientTcpPort)
					assert.Equal("127.0.0.2", c.EndpointTcpAddress)
					assert.Equal(uint32(23), c.EndpointTcpPort)
				}
			}
			if tt.wantCnt > 0 {
				assert.Equal(StatusActive, got[0].States[0].Status)
				assert.Equal(StatusPending, got[0].States[1].Status)
			}
		})
	}
	t.Run("withOrder", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
		wantCnt := 5
		for i := 0; i < wantCnt; i++ {
			_ = TestSession(t, conn, wrapper, composedOf)
		}
		got, err := repo.ListSessions(context.Background(), WithOrderByCreateTime(db.AscendingOrderBy))
		require.NoError(err)
		assert.Equal(wantCnt, len(got))

		for i := 0; i < len(got)-1; i++ {
			first := got[i].CreateTime.Timestamp.AsTime()
			second := got[i+1].CreateTime.Timestamp.AsTime()
			assert.True(first.Before(second))
		}
	})
	t.Run("withUserId", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
		wantCnt := 5
		for i := 0; i < wantCnt; i++ {
			_ = TestSession(t, conn, wrapper, composedOf)
		}
		s := TestDefaultSession(t, conn, wrapper, iamRepo)
		got, err := repo.ListSessions(context.Background(), WithUserId(s.UserId))
		require.NoError(err)
		assert.Equal(1, len(got))
		assert.Equal(s.UserId, got[0].UserId)
	})
	t.Run("withUserIdAndwithScopeId", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
		wantCnt := 5
		for i := 0; i < wantCnt; i++ {
			// Scope 1 User 1
			_ = TestSession(t, conn, wrapper, composedOf)
		}
		// Scope 2 User 2
		s := TestDefaultSession(t, conn, wrapper, iamRepo)

		// Scope 1 User 2
		coDiffUser := composedOf
		coDiffUser.AuthTokenId = s.AuthTokenId
		coDiffUser.UserId = s.UserId
		wantS := TestSession(t, conn, wrapper, coDiffUser)

		got, err := repo.ListSessions(context.Background(), WithUserId(coDiffUser.UserId), WithScopeIds([]string{coDiffUser.ScopeId}))
		require.NoError(err)
		assert.Equal(1, len(got))
		assert.Equal(wantS.UserId, got[0].UserId)
		assert.Equal(wantS.ScopeId, got[0].ScopeId)
	})
	t.Run("WithSessionIds", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
		testSessions := []*Session{}
		for i := 0; i < 10; i++ {
			s := TestSession(t, conn, wrapper, composedOf)
			_ = TestState(t, conn, s.PublicId, StatusActive)
			testSessions = append(testSessions, s)
		}
		assert.Equal(10, len(testSessions))
		withIds := []string{testSessions[0].PublicId, testSessions[1].PublicId}
		got, err := repo.ListSessions(context.Background(), WithSessionIds(withIds...), WithOrderByCreateTime(db.AscendingOrderBy))
		require.NoError(err)
		assert.Equal(2, len(got))
		assert.Equal(StatusActive, got[0].States[0].Status)
		assert.Equal(StatusPending, got[0].States[1].Status)
	})
	t.Run("withServerId", func(t *testing.T) {
		assert, require := assert.New(t), require.New(t)
		db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
		for i := 0; i < 6; i++ {
			if i%2 == 0 {
				TestWorker(t, conn, wrapper, WithServerId(fmt.Sprintf("server-%d", i/2)))
			}
			_ = TestSession(t, conn, wrapper, composedOf,
				WithServerId(fmt.Sprintf("server-%d", i/2)),
				WithDbOpts(db.WithSkipVetForWrite(true)),
			)
		}
		got, err := repo.ListSessions(context.Background())
		require.NoError(err)
		assert.Equal(6, len(got))

		for i := 0; i < 3; i++ {
			serverId := fmt.Sprintf("server-%d", i)
			got, err = repo.ListSessions(context.Background(), WithServerId(serverId))
			require.NoError(err)
			assert.Equal(2, len(got))
			for _, item := range got {
				assert.Equal(serverId, item.ServerId)
			}
		}
	})
}

func TestRepository_ListSessions_Multiple_Scopes(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	rw := db.New(conn)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)

	db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")

	const numPerScope = 10
	var projs []string
	for i := 0; i < numPerScope; i++ {
		composedOf := TestSessionParams(t, conn, wrapper, iamRepo)
		projs = append(projs, composedOf.ScopeId)
		s := TestSession(t, conn, wrapper, composedOf)
		_ = TestState(t, conn, s.PublicId, StatusActive)
	}

	got, err := repo.ListSessions(context.Background(), WithScopeIds(projs))
	require.NoError(t, err)
	assert.Equal(t, len(projs), len(got))
}

func TestRepository_CreateSession(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)

	type args struct {
		composedOf ComposedOf
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantIsError errors.Code
	}{
		{
			name: "valid",
			args: args{
				composedOf: TestSessionParams(t, conn, wrapper, iamRepo),
			},
			wantErr: false,
		},
		{
			name: "valid-with-credentials",
			args: args{
				composedOf: testSessionCredentialParams(t, conn, wrapper, iamRepo),
			},
			wantErr: false,
		},
		{
			name: "empty-userId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.UserId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name: "empty-hostId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.HostId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name: "empty-targetId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.TargetId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name: "empty-hostSetId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.HostSetId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name: "empty-authTokenId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.AuthTokenId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name: "empty-scopeId",
			args: args{
				composedOf: func() ComposedOf {
					c := TestSessionParams(t, conn, wrapper, iamRepo)
					c.ScopeId = ""
					return c
				}(),
			},
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			s := &Session{
				UserId:             tt.args.composedOf.UserId,
				HostId:             tt.args.composedOf.HostId,
				TargetId:           tt.args.composedOf.TargetId,
				HostSetId:          tt.args.composedOf.HostSetId,
				AuthTokenId:        tt.args.composedOf.AuthTokenId,
				ScopeId:            tt.args.composedOf.ScopeId,
				Endpoint:           "tcp://127.0.0.1:22",
				ExpirationTime:     tt.args.composedOf.ExpirationTime,
				ConnectionLimit:    tt.args.composedOf.ConnectionLimit,
				DynamicCredentials: tt.args.composedOf.DynamicCredentials,
			}
			ses, privKey, err := repo.CreateSession(context.Background(), wrapper, s)
			if tt.wantErr {
				assert.Error(err)
				assert.Nil(ses)
				assert.True(errors.Match(errors.T(tt.wantIsError), err))
				return
			}
			require.NoError(err)
			assert.NotNil(ses)
			assert.NotNil(privKey)
			assert.NotNil(ses.States)
			assert.NotNil(ses.CreateTime)
			assert.NotNil(ses.States[0].StartTime)
			assert.Equal(ses.States[0].Status, StatusPending)
			keyId, err := wrapper.KeyId(context.Background())
			require.NoError(err)
			assert.Equal(keyId, ses.KeyId)
			assert.Len(ses.DynamicCredentials, len(s.DynamicCredentials))
			foundSession, _, err := repo.LookupSession(context.Background(), ses.PublicId)
			assert.NoError(err)
			assert.Equal(keyId, foundSession.KeyId)

			// Account for slight offsets in nanos
			assert.True(foundSession.ExpirationTime.Timestamp.AsTime().Sub(ses.ExpirationTime.Timestamp.AsTime()) < time.Second)
			ses.ExpirationTime = foundSession.ExpirationTime

			assert.Equal(ses, foundSession)

			err = db.TestVerifyOplog(t, rw, ses.PublicId, db.WithOperation(oplog.OpType_OP_TYPE_CREATE), db.WithCreateNotBefore(10*time.Second))
			assert.Error(err)

			require.Equal(1, len(foundSession.States))
			assert.Equal(foundSession.States[0].Status, StatusPending)
			assert.Equal(s.DynamicCredentials, foundSession.DynamicCredentials)
			for _, cred := range foundSession.DynamicCredentials {
				assert.Empty(cred.CredentialId)
				assert.NotEmpty(cred.SessionId)
				assert.NotEmpty(cred.LibraryId)
				assert.NotEmpty(cred.CredentialPurpose)
			}
		})
	}
}

func TestRepository_updateState(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)

	tests := []struct {
		name                   string
		session                *Session
		newStatus              Status
		overrideSessionId      *string
		overrideSessionVersion *uint32
		wantStateCnt           int
		wantErr                bool
		wantIsError            errors.Code
	}{
		{
			name:         "canceling",
			session:      TestDefaultSession(t, conn, wrapper, iamRepo),
			newStatus:    StatusCanceling,
			wantStateCnt: 2,
			wantErr:      false,
		},
		{
			name: "closed",
			session: func() *Session {
				s := TestDefaultSession(t, conn, wrapper, iamRepo)
				_ = TestState(t, conn, s.PublicId, StatusActive)
				return s
			}(),
			newStatus:    StatusTerminated,
			wantStateCnt: 3,
			wantErr:      false,
		},
		{
			name:      "bad-version",
			session:   TestDefaultSession(t, conn, wrapper, iamRepo),
			newStatus: StatusCanceling,
			overrideSessionVersion: func() *uint32 {
				v := uint32(22)
				return &v
			}(),
			wantErr: true,
		},
		{
			name:      "empty-version",
			session:   TestDefaultSession(t, conn, wrapper, iamRepo),
			newStatus: StatusCanceling,
			overrideSessionVersion: func() *uint32 {
				v := uint32(0)
				return &v
			}(),
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name:      "bad-sessionId",
			session:   TestDefaultSession(t, conn, wrapper, iamRepo),
			newStatus: StatusCanceling,
			overrideSessionId: func() *string {
				s := "s_thisIsNotValid"
				return &s
			}(),
			wantErr: true,
		},
		{
			name:      "empty-session",
			session:   TestDefaultSession(t, conn, wrapper, iamRepo),
			newStatus: StatusCanceling,
			overrideSessionId: func() *string {
				s := ""
				return &s
			}(),
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			var id string
			var version uint32
			switch {
			case tt.overrideSessionId != nil:
				id = *tt.overrideSessionId
			default:
				id = tt.session.PublicId
			}
			switch {
			case tt.overrideSessionVersion != nil:
				version = *tt.overrideSessionVersion
			default:
				version = tt.session.Version
			}

			s, ss, err := repo.updateState(context.Background(), id, version, tt.newStatus)
			if tt.wantErr {
				require.Error(err)
				assert.Truef(errors.Match(errors.T(tt.wantIsError), err), "unexpected error %s", err.Error())
				return
			}
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(ss)
			assert.Equal(tt.wantStateCnt, len(ss))
			assert.Equal(tt.newStatus, ss[0].Status)
		})
	}
}

func TestRepository_TerminateCompletedSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	connRepo, err := NewConnectionRepository(ctx, rw, rw, kms)
	require.NoError(t, err)

	setupFn := func(limit int32, expireIn time.Duration, leaveOpen bool) *Session {
		require.NotEqualf(t, int32(0), limit, "setupFn: limit cannot be zero")
		exp := timestamppb.New(time.Now().Add(expireIn))
		composedOf := TestSessionParams(t, conn, wrapper, iamRepo)
		composedOf.ConnectionLimit = limit
		composedOf.ExpirationTime = &timestamp.Timestamp{Timestamp: exp}
		s := TestSession(t, conn, wrapper, composedOf)

		srv := TestWorker(t, conn, wrapper)
		tofu := TestTofu(t)
		s, _, err = repo.ActivateSession(context.Background(), s.PublicId, s.Version, srv.PrivateId, srv.Type, tofu)
		require.NoError(t, err)
		c := TestConnection(t, conn, s.PublicId, "127.0.0.1", 22, "127.0.0.1", 222, "127.0.0.1")
		if !leaveOpen {
			cw := CloseWith{
				ConnectionId: c.PublicId,
				BytesUp:      1,
				BytesDown:    1,
				ClosedReason: ConnectionClosedByUser,
			}
			_, err = connRepo.closeConnections(context.Background(), []CloseWith{cw})
			require.NoError(t, err)
		}
		return s
	}

	type testArgs struct {
		sessions   []*Session
		wantTermed map[string]TerminationReason
	}
	tests := []struct {
		name    string
		setup   func() testArgs
		wantErr bool
	}{
		{
			name: "sessions-with-closed-connections",
			setup: func() testArgs {
				cnt := 1
				wantTermed := map[string]TerminationReason{}
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					// make one with closed connections
					s := setupFn(1, time.Hour+1, false)
					wantTermed[s.PublicId] = ConnectionLimit
					sessions = append(sessions, s)

					// make one with connection left open
					s2 := setupFn(1, time.Hour+1, true)
					sessions = append(sessions, s2)
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: wantTermed,
				}
			},
		},
		{
			name: "sessions-with-open-and-closed-connections",
			setup: func() testArgs {
				cnt := 5
				wantTermed := map[string]TerminationReason{}
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					// make one with closed connections
					s := setupFn(2, time.Hour+1, false)
					_ = TestConnection(t, conn, s.PublicId, "127.0.0.1", 22, "127.0.0.1", 222, "127.0.0.1")
					sessions = append(sessions, s)
					wantTermed[s.PublicId] = ConnectionLimit
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: nil,
				}
			},
		},
		{
			name: "sessions-with-no-connections",
			setup: func() testArgs {
				cnt := 5
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					s := TestDefaultSession(t, conn, wrapper, iamRepo)
					sessions = append(sessions, s)
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: nil,
				}
			},
		},
		{
			name: "sessions-with-open-connections",
			setup: func() testArgs {
				cnt := 5
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					s := setupFn(1, time.Hour+1, true)
					sessions = append(sessions, s)
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: nil,
				}
			},
		},
		{
			name: "expired-sessions",
			setup: func() testArgs {
				cnt := 5
				wantTermed := map[string]TerminationReason{}
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					// make one with closed connections
					s := setupFn(1, time.Millisecond+1, false)
					// make one with connection left open
					s2 := setupFn(1, time.Millisecond+1, true)
					sessions = append(sessions, s, s2)
					wantTermed[s.PublicId] = TimedOut
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: wantTermed,
				}
			},
		},
		{
			name: "canceled-sessions-with-closed-connections",
			setup: func() testArgs {
				cnt := 1
				wantTermed := map[string]TerminationReason{}
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					// make one with limit of 3 and closed connections
					s := setupFn(3, time.Hour+1, false)
					wantTermed[s.PublicId] = SessionCanceled
					sessions = append(sessions, s)

					// make one with connection left open
					s2 := setupFn(1, time.Hour+1, true)
					sessions = append(sessions, s2)

					// now cancel the sessions
					for _, ses := range sessions {
						_, err := repo.CancelSession(context.Background(), ses.PublicId, ses.Version)
						require.NoError(t, err)
					}
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: wantTermed,
				}
			},
		},
		{
			name: "sessions-with-unlimited-connections",
			setup: func() testArgs {
				cnt := 5
				wantTermed := map[string]TerminationReason{}
				sessions := make([]*Session, 0, 5)
				for i := 0; i < cnt; i++ {
					// make one with unlimited connections
					s := setupFn(-1, time.Hour+1, false)
					// make one with limit of one all connections closed
					s2 := setupFn(1, time.Hour+1, false)
					sessions = append(sessions, s, s2)
					wantTermed[s2.PublicId] = ConnectionLimit
				}
				return testArgs{
					sessions:   sessions,
					wantTermed: wantTermed,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			db.TestDeleteWhere(t, conn, func() interface{} { i := AllocSession(); return &i }(), "1=1")
			args := tt.setup()

			got, err := repo.TerminateCompletedSessions(context.Background())
			if tt.wantErr {
				require.Error(err)
				return
			}
			assert.NoError(err)
			t.Logf("terminated: %d", got)
			var foundTerminated int
			for _, ses := range args.sessions {
				found, _, err := repo.LookupSession(context.Background(), ses.PublicId)
				require.NoError(err)
				_, shouldBeTerminated := args.wantTermed[found.PublicId]
				if shouldBeTerminated {
					if found.TerminationReason != "" {
						foundTerminated += 1
					}
					assert.Equal(args.wantTermed[found.PublicId].String(), found.TerminationReason)
					t.Logf("terminated %s has a connection limit of %d", found.PublicId, found.ConnectionLimit)
					conn, err := connRepo.ListConnectionsBySessionId(context.Background(), found.PublicId)
					require.NoError(err)
					for _, sc := range conn {
						c, cs, err := connRepo.LookupConnection(context.Background(), sc.PublicId)
						require.NoError(err)
						assert.NotEmpty(c.ClosedReason)
						for _, s := range cs {
							t.Logf("%s session %s connection state %s at %s", found.PublicId, s.ConnectionId, s.Status, s.EndTime)
						}
					}
				} else {
					t.Logf("not terminated %s has a connection limit of %d", found.PublicId, found.ConnectionLimit)
					assert.Equal("", found.TerminationReason)
					conn, err := connRepo.ListConnectionsBySessionId(context.Background(), found.PublicId)
					require.NoError(err)
					for _, sc := range conn {
						cs, err := fetchConnectionStates(context.Background(), rw, sc.PublicId)
						require.NoError(err)
						for _, s := range cs {
							t.Logf("%s session %s connection state %s at %s", found.PublicId, s.ConnectionId, s.Status, s.EndTime)
						}
					}
				}
			}
			assert.Equal(len(args.wantTermed), foundTerminated)
		})
	}
}

func TestRepository_CancelSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	connRepo, err := NewConnectionRepository(ctx, rw, rw, kms)
	require.NoError(t, err)
	setupFn := func() *Session {
		session := TestDefaultSession(t, conn, wrapper, iamRepo)
		_ = TestConnection(t, conn, session.PublicId, "127.0.0.1", 22, "127.0.0.1", 2222, "127.0.0.1")
		return session
	}
	tests := []struct {
		name                   string
		session                *Session
		overrideSessionId      *string
		overrideSessionVersion *uint32
		wantErr                bool
		wantIsError            errors.Code
		wantStatus             Status
	}{
		{
			name:       "valid",
			session:    setupFn(),
			wantStatus: StatusCanceling,
		},
		{
			name: "already-terminated",
			session: func() *Session {
				session := TestDefaultSession(t, conn, wrapper, iamRepo)
				c := TestConnection(t, conn, session.PublicId, "127.0.0.1", 22, "127.0.0.1", 2222, "127.0.0.1")
				cw := CloseWith{
					ConnectionId: c.PublicId,
					BytesUp:      1,
					BytesDown:    1,
					ClosedReason: ConnectionClosedByUser,
				}
				_, err = CloseConnections(ctx, repo, connRepo, []CloseWith{cw})
				require.NoError(t, err)
				s, _, err := repo.LookupSession(ctx, session.PublicId)
				require.NoError(t, err)
				assert.Equal(t, StatusTerminated, s.States[0].Status)
				// The two transactions to cancel connections and terminate the session will result in version being 2, not 1
				session.Version = s.Version
				return session
			}(),
			wantStatus: StatusTerminated,
		},
		{
			name:    "bad-session-id",
			session: setupFn(),
			overrideSessionId: func() *string {
				id, err := newId()
				require.NoError(t, err)
				return &id
			}(),
			wantErr:    true,
			wantStatus: StatusCanceling,
		},
		{
			name:    "missing-session-id",
			session: setupFn(),
			overrideSessionId: func() *string {
				id := ""
				return &id
			}(),
			wantStatus:  StatusCanceling,
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name:    "bad-version-id",
			session: setupFn(),
			overrideSessionVersion: func() *uint32 {
				v := uint32(101)
				return &v
			}(),
			wantStatus: StatusCanceling,
			wantErr:    true,
		},
		{
			name:    "missing-version-id",
			session: setupFn(),
			overrideSessionVersion: func() *uint32 {
				v := uint32(0)
				return &v
			}(),
			wantStatus:  StatusCanceling,
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			var id string
			var version uint32
			switch {
			case tt.overrideSessionId != nil:
				id = *tt.overrideSessionId
			default:
				id = tt.session.PublicId
			}
			switch {
			case tt.overrideSessionVersion != nil:
				version = *tt.overrideSessionVersion
			default:
				version = tt.session.Version
			}
			s, err := repo.CancelSession(ctx, id, version)
			if tt.wantErr {
				require.Error(err)
				assert.Truef(errors.Match(errors.T(tt.wantIsError), err), "unexpected error %s", err.Error())
				return
			}
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(s.States)
			assert.Equal(tt.wantStatus, s.States[0].Status)

			stateCnt := len(s.States)
			origStartTime := s.States[0].StartTime
			// check idempontency
			s2, err := repo.CancelSession(context.Background(), id, version+1)
			require.NoError(err)
			require.NotNil(s2)
			require.NotNil(s2.States)
			assert.Equal(stateCnt, len(s2.States))
			assert.Equal(tt.wantStatus, s.States[0].Status)
			assert.Equal(origStartTime, s2.States[0].StartTime)
		})
	}
}

func TestRepository_CancelSessionViaFKNull(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)
	setupFn := func() *Session {
		session := TestDefaultSession(t, conn, wrapper, iamRepo)
		_ = TestConnection(t, conn, session.PublicId, "127.0.0.1", 22, "127.0.0.1", 2222, "127.0.0.1")
		return session
	}
	type cancelFk struct {
		s      *Session
		fkType interface{}
	}
	tests := []struct {
		name     string
		cancelFk cancelFk
	}{
		{
			name: "UserId",
			cancelFk: func() cancelFk {
				s := setupFn()
				t := &iam.User{
					User: &iamStore.User{
						PublicId: s.UserId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "Host",
			cancelFk: func() cancelFk {
				s := setupFn()

				t := &static.Host{
					Host: &staticStore.Host{
						PublicId: s.HostId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "Target",
			cancelFk: func() cancelFk {
				s := setupFn()

				t := &tcp.Target{
					Target: &tcpStore.Target{
						PublicId: s.TargetId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "HostSet",
			cancelFk: func() cancelFk {
				s := setupFn()

				t := &static.HostSet{
					HostSet: &staticStore.HostSet{
						PublicId: s.HostSetId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "AuthToken",
			cancelFk: func() cancelFk {
				s := setupFn()

				t := &authtoken.AuthToken{
					AuthToken: &authtokenStore.AuthToken{
						PublicId: s.AuthTokenId,
					},
				}
				// override the table name so we can delete this thing, since
				// it's default table name is a non-writable view.
				t.SetTableName("auth_token")
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "Scope",
			cancelFk: func() cancelFk {
				s := setupFn()

				t := &iam.Scope{
					Scope: &iamStore.Scope{
						PublicId: s.ScopeId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
		{
			name: "canceled-only-once",
			cancelFk: func() cancelFk {
				s := setupFn()
				var err error
				s, err = repo.CancelSession(context.Background(), s.PublicId, s.Version)
				require.NoError(t, err)
				require.Equal(t, StatusCanceling, s.States[0].Status)

				t := &static.Host{
					Host: &staticStore.Host{
						PublicId: s.HostId,
					},
				}
				return cancelFk{
					s:      s,
					fkType: t,
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			s, _, err := repo.LookupSession(context.Background(), tt.cancelFk.s.PublicId)
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(s.States)

			rowsDeleted, err := rw.Delete(context.Background(), tt.cancelFk.fkType)
			if err != nil {
				var pgError *pgconn.PgError
				if errors.As(err, &pgError) {
					t.Log(pgError.Message)
					t.Log(pgError.Detail)
					t.Log(pgError.Where)
					t.Log(pgError.ConstraintName)
					t.Log(pgError.TableName)
				}
			}
			require.NoError(err)
			require.Equal(1, rowsDeleted)

			s, _, err = repo.LookupSession(context.Background(), tt.cancelFk.s.PublicId)
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(s.States)
			assert.Equal(StatusCanceling, s.States[0].Status)
			canceledCnt := 0
			for _, ss := range s.States {
				if ss.Status == StatusCanceling {
					canceledCnt += 1
				}
			}
			assert.Equal(1, canceledCnt)
		})
	}
}

func TestRepository_ActivateSession(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)
	worker := TestWorker(t, conn, wrapper)

	tofu := TestTofu(t)
	tests := []struct {
		name                   string
		session                *Session
		overrideSessionId      *string
		overrideSessionVersion *uint32
		wantErr                bool
		wantIsError            errors.Code
	}{
		{
			name:    "valid",
			session: TestDefaultSession(t, conn, wrapper, iamRepo),
			wantErr: false,
		},
		{
			name: "already-active",
			session: func() *Session {
				s := TestDefaultSession(t, conn, wrapper, iamRepo)
				activeSession, _, err := repo.ActivateSession(context.Background(), s.PublicId, s.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				return activeSession
			}(),
			wantErr: true,
		},
		{
			name:    "bad-session-id",
			session: TestDefaultSession(t, conn, wrapper, iamRepo),
			overrideSessionId: func() *string {
				id, err := newId()
				require.NoError(t, err)
				return &id
			}(),
			wantErr: true,
		},
		{
			name:    "bad-session-version",
			session: TestDefaultSession(t, conn, wrapper, iamRepo),
			overrideSessionVersion: func() *uint32 {
				v := uint32(100)
				return &v
			}(),
			wantErr: true,
		},
		{
			name:    "empty-session-id",
			session: TestDefaultSession(t, conn, wrapper, iamRepo),
			overrideSessionId: func() *string {
				id := ""
				return &id
			}(),
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
		{
			name:    "empty-session-version",
			session: TestDefaultSession(t, conn, wrapper, iamRepo),
			overrideSessionVersion: func() *uint32 {
				v := uint32(0)
				return &v
			}(),
			wantErr:     true,
			wantIsError: errors.InvalidParameter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			var id string
			var version uint32
			switch {
			case tt.overrideSessionId != nil:
				id = *tt.overrideSessionId
			default:
				id = tt.session.PublicId
			}
			switch {
			case tt.overrideSessionVersion != nil:
				version = *tt.overrideSessionVersion
			default:
				version = tt.session.Version
			}
			s, ss, err := repo.ActivateSession(context.Background(), id, version, worker.PrivateId, worker.Type, tofu)
			if tt.wantErr {
				require.Error(err)
				assert.Truef(errors.Match(errors.T(tt.wantIsError), err), "unexpected error %s", err.Error())
				return
			}
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(ss)
			assert.Equal(tofu, s.TofuToken)
			assert.Equal(2, len(ss))
			assert.Equal(StatusActive, ss[0].Status)
		})
		t.Run("already active", func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			session := TestDefaultSession(t, conn, wrapper, iamRepo)
			s, ss, err := repo.ActivateSession(context.Background(), session.PublicId, 1, worker.PrivateId, worker.Type, tofu)
			require.NoError(err)
			require.NotNil(s)
			require.NotNil(ss)
			assert.Equal(2, len(ss))
			assert.Equal(StatusActive, ss[0].Status)

			_, _, err = repo.ActivateSession(context.Background(), session.PublicId, 1, worker.PrivateId, worker.Type, tofu)
			require.Error(err)

			_, _, err = repo.ActivateSession(context.Background(), session.PublicId, 2, worker.PrivateId, worker.Type, tofu)
			require.Error(err)
		})
	}
}

func TestRepository_DeleteSession(t *testing.T) {
	t.Parallel()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	iamRepo := iam.TestRepo(t, conn, wrapper)
	kms := kms.TestKms(t, conn, wrapper)
	repo, err := NewRepository(rw, rw, kms)
	require.NoError(t, err)

	type args struct {
		session *Session
		opt     []Option
	}
	tests := []struct {
		name            string
		args            args
		wantRowsDeleted int
		wantErr         bool
		wantErrMsg      string
	}{
		{
			name: "valid",
			args: args{
				session: TestDefaultSession(t, conn, wrapper, iamRepo),
			},
			wantRowsDeleted: 1,
			wantErr:         false,
		},
		{
			name: "no-public-id",
			args: args{
				session: func() *Session {
					s := AllocSession()
					return &s
				}(),
			},
			wantRowsDeleted: 0,
			wantErr:         true,
			wantErrMsg:      "session.(Repository).DeleteSession: missing public id: parameter violation: error #100",
		},
		{
			name: "not-found",
			args: args{
				session: func() *Session {
					s := TestDefaultSession(t, conn, wrapper, iamRepo)
					id, err := newId()
					require.NoError(t, err)
					s.PublicId = id
					return s
				}(),
			},
			wantRowsDeleted: 0,
			wantErr:         true,
			wantErrMsg:      "db.LookupById: record not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			deletedRows, err := repo.DeleteSession(context.Background(), tt.args.session.PublicId, tt.args.opt...)
			if tt.wantErr {
				assert.Error(err)
				assert.Equal(0, deletedRows)
				assert.Contains(err.Error(), tt.wantErrMsg)
				err = db.TestVerifyOplog(t, rw, tt.args.session.PublicId, db.WithOperation(oplog.OpType_OP_TYPE_DELETE), db.WithCreateNotBefore(10*time.Second))
				assert.Error(err)
				assert.True(errors.IsNotFoundError(err))
				return
			}
			assert.NoError(err)
			assert.Equal(tt.wantRowsDeleted, deletedRows)
			foundSession, _, err := repo.LookupSession(context.Background(), tt.args.session.PublicId)
			assert.NoError(err)
			assert.Nil(foundSession)

			err = db.TestVerifyOplog(t, rw, tt.args.session.PublicId, db.WithOperation(oplog.OpType_OP_TYPE_DELETE), db.WithCreateNotBefore(10*time.Second))
			assert.Error(err)
		})
	}
}

func testSessionCredentialParams(t *testing.T, conn *db.DB, wrapper wrapping.Wrapper, iamRepo *iam.Repository) ComposedOf {
	t.Helper()
	params := TestSessionParams(t, conn, wrapper, iamRepo)
	require := require.New(t)
	rw := db.New(conn)

	ctx := context.Background()

	kms := kms.TestKms(t, conn, wrapper)
	targetRepo, err := target.NewRepository(rw, rw, kms)
	require.NoError(err)
	tar, _, _, err := targetRepo.LookupTarget(ctx, params.TargetId)
	require.NoError(err)
	require.NotNil(tar)

	stores := vault.TestCredentialStores(t, conn, wrapper, params.ScopeId, 1)
	libIds := vault.TestCredentialLibraries(t, conn, wrapper, stores[0].GetPublicId(), 2)
	libs := []*target.CredentialLibrary{
		target.TestNewCredentialLibrary(tar.GetPublicId(), libIds[0].GetPublicId(), cred.ApplicationPurpose),
		target.TestNewCredentialLibrary(tar.GetPublicId(), libIds[1].GetPublicId(), cred.ApplicationPurpose),
	}

	_, _, _, err = targetRepo.AddTargetCredentialSources(ctx, tar.GetPublicId(), tar.GetVersion(), libs)
	require.NoError(err)
	creds := []*DynamicCredential{
		NewDynamicCredential(libIds[0].GetPublicId(), cred.ApplicationPurpose),
		NewDynamicCredential(libIds[1].GetPublicId(), cred.ApplicationPurpose),
	}
	params.DynamicCredentials = creds
	return params
}
