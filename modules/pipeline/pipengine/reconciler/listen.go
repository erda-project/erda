package reconciler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/strutil"
)

// Listen watch incoming pipelines which need to be scheduled from etcd.
func (r *Reconciler) Listen() {
	rlog.Infof("start listen")
	for {
		_ = r.js.IncludeWatch().Watch(context.Background(), etcdReconcilerWatchPrefix, true, true, true, nil,
			func(key string, _ interface{}, t storetypes.ChangeType) (_ error) {

				// async reconcile, non-blocking, so we can watch subsequent incoming pipelines
				go func() {
					rlog.Infof("watched a key change, key: %s, changeType: %s", key, t.String())

					// parse pipelineID
					pipelineID, err := parsePipelineIDFromWatchedKey(key)
					if err != nil {
						rlog.Errorf("failed to parse pipelineID from watched key, key: %s, err: %v", key, err)
						return
					}

					// construct context for pipeline reconciler
					pCtx := makeContextForPipelineReconcile(pipelineID)

					// reconciler
					reconcileErr := r.reconcile(pCtx, pipelineID)

					// if reconcile failed, put and wait next reconcile
					if reconcileErr != nil {
						rlog.PErrorf(pipelineID, "failed to reconcile pipeline, err: %v", err)
						// add to reconciler again, wait next reconcile
						time.Sleep(time.Second * 5)
						r.Add(pipelineID)
						return
					}
				}()

				return nil
			},
		)
	}
}

// makeContextForPipelineReconcile
func makeContextForPipelineReconcile(pipelineID uint64) context.Context {
	pCtx := context.WithValue(context.Background(), ctxKeyPipelineID, pipelineID)
	exitCh := make(chan struct{})
	pCtx = context.WithValue(pCtx, ctxKeyPipelineExitCh, exitCh)
	pCtx, pCancel := context.WithCancel(pCtx)
	pCtx = context.WithValue(pCtx, ctxKeyPipelineExitChCancelFunc, pCancel)
	return pCtx
}

// parsePipelineIDFromWatchedKey get pipelineID from watched key.
func parsePipelineIDFromWatchedKey(key string) (uint64, error) {
	pipelineIDStr := strutil.TrimPrefixes(key, etcdReconcilerWatchPrefix)
	return strconv.ParseUint(pipelineIDStr, 10, 64)
}

// makePipelineWatchedKey construct etcd watched key by pipelineID.
func makePipelineWatchedKey(pipelineID uint64) string {
	return fmt.Sprintf("%s%d", etcdReconcilerWatchPrefix, pipelineID)
}
