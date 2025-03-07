package session_test

import (
	"context"
	"testing"

	"github.com/hashicorp/boundary/internal/authtoken"
	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/host/static"
	"github.com/hashicorp/boundary/internal/iam"
	"github.com/hashicorp/boundary/internal/kms"
	"github.com/hashicorp/boundary/internal/servers"
	"github.com/hashicorp/boundary/internal/session"
	"github.com/hashicorp/boundary/internal/target"
	"github.com/hashicorp/boundary/internal/target/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerStatusReport(t *testing.T) {
	ctx := context.Background()
	conn, _ := db.TestSetup(t, "postgres")
	rw := db.New(conn)
	wrapper := db.TestWrapper(t)
	kms := kms.TestKms(t, conn, wrapper)
	org, prj := iam.TestScopes(t, iam.TestRepo(t, conn, wrapper))

	serverRepo, _ := servers.NewRepository(rw, rw, kms)
	serverRepo.UpsertServer(ctx, &servers.Server{
		PrivateId: "test_controller1",
		Type:      "controller",
		Address:   "127.0.0.1",
	})
	serverRepo.UpsertServer(ctx, &servers.Server{
		PrivateId: "test_worker1",
		Type:      "worker",
		Address:   "127.0.0.1",
	})

	repo, err := session.NewRepository(rw, rw, kms)
	require.NoError(t, err)
	connRepo, err := session.NewConnectionRepository(ctx, rw, rw, kms, session.WithWorkerStateDelay(0))
	require.NoError(t, err)

	at := authtoken.TestAuthToken(t, conn, kms, org.GetPublicId())
	uId := at.GetIamUserId()
	hc := static.TestCatalogs(t, conn, prj.GetPublicId(), 1)[0]
	hs := static.TestSets(t, conn, hc.GetPublicId(), 1)[0]
	h := static.TestHosts(t, conn, hc.GetPublicId(), 1)[0]
	static.TestSetMembers(t, conn, hs.GetPublicId(), []*static.Host{h})
	tar := tcp.TestTarget(
		ctx,
		t, conn, prj.GetPublicId(), "test",
		target.WithHostSources([]string{hs.GetPublicId()}),
		target.WithSessionConnectionLimit(-1),
	)

	type testCase struct {
		worker              *servers.Server
		req                 []session.StateReport
		want                []session.StateReport
		orphanedConnections []string
	}
	cases := []struct {
		name   string
		caseFn func(t *testing.T) testCase
	}{
		{
			name: "No Sessions",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				return testCase{
					worker: worker,
					req:    []session.StateReport{},
					want:   []session.StateReport{},
				}
			},
		},
		{
			name: "No Sessions already canceled",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				require.NoError(t, err)

				_, _, err = connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)

				_, err = repo.CancelSession(ctx, sess.PublicId, sess.Version)
				require.NoError(t, err)

				return testCase{
					worker: worker,
					req:    []session.StateReport{},
					want:   []session.StateReport{},
				}
			},
		},
		{
			name: "Still Active",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				require.NoError(t, err)

				connection, _, err := connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)
				return testCase{
					worker: worker,
					req: []session.StateReport{
						{
							SessionId:     sess.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection.PublicId},
						},
					},
					want: []session.StateReport{},
				}
			},
		},
		{
			name: "SessionClosed",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				connection, _, err := connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)
				_, err = repo.CancelSession(ctx, sess.PublicId, sess.Version)
				require.NoError(t, err)

				return testCase{
					worker: worker,
					req: []session.StateReport{
						{
							SessionId:     sess.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection.PublicId},
						},
					},
					want: []session.StateReport{
						{
							SessionId: sess.PublicId,
							Status:    session.StatusCanceling,
						},
					},
				}
			},
		},
		{
			name: "MultipleSessionsClosed",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				connection, _, err := connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)
				_, err = repo.CancelSession(ctx, sess.PublicId, sess.Version)
				require.NoError(t, err)

				sess2 := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu2 := session.TestTofu(t)
				sess2, _, err = repo.ActivateSession(ctx, sess2.PublicId, sess2.Version, worker.PrivateId, worker.Type, tofu2)
				require.NoError(t, err)
				connection2, _, err := connRepo.AuthorizeConnection(ctx, sess2.PublicId, worker.PrivateId)
				require.NoError(t, err)
				_, err = repo.CancelSession(ctx, sess2.PublicId, sess2.Version)
				require.NoError(t, err)

				return testCase{
					worker: worker,
					req: []session.StateReport{
						{
							SessionId:     sess.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection.PublicId},
						},
						{
							SessionId:     sess2.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection2.PublicId},
						},
					},
					want: []session.StateReport{
						{
							SessionId: sess.PublicId,
							Status:    session.StatusCanceling,
						},
						{
							SessionId: sess2.PublicId,
							Status:    session.StatusCanceling,
						},
					},
				}
			},
		},
		{
			name: "OrphanedConnection",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				connection, _, err := connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)

				sess2 := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu2 := session.TestTofu(t)
				sess2, _, err = repo.ActivateSession(ctx, sess2.PublicId, sess2.Version, worker.PrivateId, worker.Type, tofu2)
				require.NoError(t, err)
				connection2, _, err := connRepo.AuthorizeConnection(ctx, sess2.PublicId, worker.PrivateId)
				require.NoError(t, err)
				require.NotEqual(t, connection.PublicId, connection2.PublicId)

				return testCase{
					worker: worker,
					req: []session.StateReport{
						{
							SessionId:     sess2.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection2.PublicId},
						},
					},
					want:                []session.StateReport{},
					orphanedConnections: []string{connection.PublicId},
				}
			},
		},
		{
			name: "MultipleSessionsAndOrphanedConnections",
			caseFn: func(t *testing.T) testCase {
				worker := session.TestWorker(t, conn, wrapper)
				sess := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu := session.TestTofu(t)
				sess, _, err = repo.ActivateSession(ctx, sess.PublicId, sess.Version, worker.PrivateId, worker.Type, tofu)
				require.NoError(t, err)
				connection, _, err := connRepo.AuthorizeConnection(ctx, sess.PublicId, worker.PrivateId)
				require.NoError(t, err)
				_, err = repo.CancelSession(ctx, sess.PublicId, sess.Version)
				require.NoError(t, err)

				sess2 := session.TestSession(t, conn, wrapper, session.ComposedOf{
					UserId:          uId,
					HostId:          h.GetPublicId(),
					TargetId:        tar.GetPublicId(),
					HostSetId:       hs.GetPublicId(),
					AuthTokenId:     at.GetPublicId(),
					ScopeId:         prj.GetPublicId(),
					Endpoint:        "tcp://127.0.0.1:22",
					ConnectionLimit: 10,
				})
				tofu2 := session.TestTofu(t)
				sess2, _, err = repo.ActivateSession(ctx, sess2.PublicId, sess2.Version, worker.PrivateId, worker.Type, tofu2)
				require.NoError(t, err)
				connection2, _, err := connRepo.AuthorizeConnection(ctx, sess2.PublicId, worker.PrivateId)
				require.NoError(t, err)
				connection3, _, err := connRepo.AuthorizeConnection(ctx, sess2.PublicId, worker.PrivateId)
				require.NoError(t, err)
				_, err = repo.CancelSession(ctx, sess2.PublicId, sess2.Version)
				require.NoError(t, err)

				return testCase{
					worker: worker,
					req: []session.StateReport{
						{
							SessionId:     sess.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection.PublicId},
						},
						{
							SessionId:     sess2.PublicId,
							Status:        session.StatusActive,
							ConnectionIds: []string{connection2.PublicId},
						},
					},
					want: []session.StateReport{
						{
							SessionId: sess.PublicId,
							Status:    session.StatusCanceling,
						},
						{
							SessionId: sess2.PublicId,
							Status:    session.StatusCanceling,
						},
					},
					orphanedConnections: []string{connection3.PublicId},
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)

			tc := tt.caseFn(t)

			got, err := session.WorkerStatusReport(ctx, repo, connRepo, tc.worker.PrivateId, tc.req)
			require.NoError(err)
			assert.ElementsMatch(tc.want, got)
			for _, dc := range tc.orphanedConnections {
				gotConn, states, err := connRepo.LookupConnection(ctx, dc)
				require.NoError(err)
				assert.Equal(session.ConnectionSystemError, session.ClosedReason(gotConn.ClosedReason))
				assert.Equal(2, len(states))
				assert.Nil(states[0].EndTime)
				assert.Equal(session.StatusClosed, states[0].Status)
			}
		})
	}
}
