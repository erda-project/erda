package cachehelpers

import (
    "context"
    "errors"
    "testing"

    clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
    modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
    providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
    "github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
    "github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

// mockManager implements cachetypes.Manager for testing
type mockManager struct {
    listAll   map[cachetypes.ItemType]struct{
        total uint64
        value any
        err   error
    }
    getByID   map[cachetypes.ItemType]map[string]struct{
        value any
        err   error
    }
    calls struct {
        listAll []cachetypes.ItemType
        getByID []struct{ itemType cachetypes.ItemType; id string }
    }
}

func newMockManager() *mockManager {
    return &mockManager{
        listAll: make(map[cachetypes.ItemType]struct{
            total uint64
            value any
            err   error
        }),
        getByID: make(map[cachetypes.ItemType]map[string]struct{
            value any
            err   error
        }),
    }
}

func (m *mockManager) ListAll(ctx context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
    m.calls.listAll = append(m.calls.listAll, itemType)
    if res, ok := m.listAll[itemType]; ok {
        return res.total, res.value, res.err
    }
    return 0, nil, errors.New("unexpected ListAll call")
}

func (m *mockManager) GetByID(ctx context.Context, itemType cachetypes.ItemType, id string) (any, error) {
    m.calls.getByID = append(m.calls.getByID, struct{ itemType cachetypes.ItemType; id string }{itemType: itemType, id: id})
    if byType, ok := m.getByID[itemType]; ok {
        if res, ok := byType[id]; ok {
            return res.value, res.err
        }
    }
    return nil, errors.New("unexpected GetByID call")
}

func (m *mockManager) TriggerRefresh(ctx context.Context, _ ...cachetypes.ItemType) {}

func withCacheManager(t *testing.T, m cachetypes.Manager) context.Context {
    t.Helper()
    ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
    ctxhelper.PutCacheManager(ctx, m)
    return ctx
}

func Test_listAllClientBelongedModels_Success(t *testing.T) {
    mm := newMockManager()
    models := []*modelpb.Model{
        {Id: "m1", ClientId: "c1", ProviderId: "p1"},
        {Id: "m2", ClientId: "c1", ProviderId: "p2"},
        {Id: "m3", ClientId: "c2", ProviderId: "p1"},
    }
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{value: models}

    ctx := withCacheManager(t, mm)
    got, err := _listAllClientBelongedModels(ctx, "c1")
    if err != nil {
        t.Fatalf("_listAllClientBelongedModels error: %v", err)
    }
    if len(got) != 2 {
        t.Fatalf("expected 2 models, got %d", len(got))
    }
    if got[0].Id != "m1" || got[1].Id != "m2" {
        t.Fatalf("unexpected models order/ids: %v, %v", got[0].Id, got[1].Id)
    }
}

func Test_listAllClientBelongedModels_Error(t *testing.T) {
    mm := newMockManager()
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{err: errors.New("list error")}
    ctx := withCacheManager(t, mm)
    if _, err := _listAllClientBelongedModels(ctx, "c1"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}

func Test_listAllClientAssignedModels_NoMatch(t *testing.T) {
    mm := newMockManager()
    relations := []*clientmodelrelationpb.ClientModelRelation{
        {ClientId: "other", ModelId: "mx"},
    }
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: relations}
    ctx := withCacheManager(t, mm)
    got, err := _listAllClientAssignedModels(ctx, "c1")
    if err != nil {
        t.Fatalf("_listAllClientAssignedModels error: %v", err)
    }
    if len(got) != 0 {
        t.Fatalf("expected 0 models, got %d", len(got))
    }
}

func Test_listAllClientAssignedModels_ListError(t *testing.T) {
    mm := newMockManager()
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{err: errors.New("relations list error")}
    ctx := withCacheManager(t, mm)
    if _, err := _listAllClientAssignedModels(ctx, "c1"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}

func Test_listAllClientAssignedModels_GetModelError(t *testing.T) {
    mm := newMockManager()
    relations := []*clientmodelrelationpb.ClientModelRelation{
        {ClientId: "c1", ModelId: "m1"},
    }
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: relations}
    // return error on fetching the model
    mm.getByID[cachetypes.ItemTypeModel] = map[string]struct{ value any; err error }{
        "m1": {value: nil, err: errors.New("model not found")},
    }
    ctx := withCacheManager(t, mm)
    if _, err := _listAllClientAssignedModels(ctx, "c1"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}

func TestListAllClientModels_Success_WithDedupProviderFetch(t *testing.T) {
    mm := newMockManager()
    // All models in cache (used by belonged function)
    models := []*modelpb.Model{
        {Id: "m1", ClientId: "c1", ProviderId: "p1"}, // belonged
        {Id: "m2", ClientId: "c1", ProviderId: "p2"}, // belonged
        {Id: "m3", ClientId: "c2", ProviderId: "p1"}, // for assignment
    }
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{value: models}
    // Relations include one assigned model m3 for c1
    relations := []*clientmodelrelationpb.ClientModelRelation{
        {ClientId: "c1", ModelId: "m3"},
        {ClientId: "other", ModelId: "mx"},
    }
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: relations}
    // GetByID for models (only m3 is used by assigned list)
    mm.getByID[cachetypes.ItemTypeModel] = map[string]struct{ value any; err error }{
        "m3": {value: models[2]},
    }
    // Providers
    p1 := &providerpb.ServiceProvider{Id: "p1"}
    p2 := &providerpb.ServiceProvider{Id: "p2"}
    mm.getByID[cachetypes.ItemTypeProvider] = map[string]struct{ value any; err error }{
        "p1": {value: p1},
        "p2": {value: p2},
    }

    ctx := withCacheManager(t, mm)
    got, err := ListAllClientModels(ctx, "c1")
    if err != nil {
        t.Fatalf("ListAllClientModels error: %v", err)
    }
    if len(got) != 3 {
        t.Fatalf("expected 3 models, got %d", len(got))
    }
    // Verify providers are set correctly
    for _, mwp := range got {
        if mwp.Provider == nil || mwp.Provider.Id == "" {
            t.Fatalf("provider not attached for model %s", mwp.Id)
        }
        if mwp.Provider.Id != mwp.ProviderId {
            t.Fatalf("provider id mismatch, got provider %s for model providerId %s", mwp.Provider.Id, mwp.ProviderId)
        }
    }
    // Ensure provider p1 fetched only once though used by two models (m1 and m3)
    var p1FetchCount int
    for _, c := range mm.calls.getByID {
        if c.itemType == cachetypes.ItemTypeProvider && c.id == "p1" {
            p1FetchCount++
        }
    }
    if p1FetchCount != 1 {
        t.Fatalf("expected provider p1 fetch once, got %d", p1FetchCount)
    }
}

func TestListAllClientModels_ErrorOnProviderFetch(t *testing.T) {
    mm := newMockManager()
    models := []*modelpb.Model{
        {Id: "m1", ClientId: "c1", ProviderId: "pX"},
    }
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{value: models}
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: []*clientmodelrelationpb.ClientModelRelation{}}
    // provider fetch err
    mm.getByID[cachetypes.ItemTypeProvider] = map[string]struct{ value any; err error }{
        "pX": {err: errors.New("provider not found")},
    }
    ctx := withCacheManager(t, mm)
    if _, err := ListAllClientModels(ctx, "c1"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}

func TestGetOneClientModel_Success(t *testing.T) {
    mm := newMockManager()
    models := []*modelpb.Model{
        {Id: "m1", ClientId: "c1", ProviderId: "p1"},
        {Id: "m2", ClientId: "c1", ProviderId: "p2"},
    }
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{value: models}
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: []*clientmodelrelationpb.ClientModelRelation{}}
    mm.getByID[cachetypes.ItemTypeProvider] = map[string]struct{ value any; err error }{
        "p1": {value: &providerpb.ServiceProvider{Id: "p1"}},
        "p2": {value: &providerpb.ServiceProvider{Id: "p2"}},
    }
    ctx := withCacheManager(t, mm)
    got, err := GetOneClientModel(ctx, "c1", "m2")
    if err != nil {
        t.Fatalf("GetOneClientModel error: %v", err)
    }
    if got == nil || got.Model == nil || got.Provider == nil {
        t.Fatalf("expected non-nil model with provider")
    }
    if got.Id != "m2" || got.Provider.Id != "p2" {
        t.Fatalf("unexpected result, model=%s provider=%s", got.Id, got.Provider.Id)
    }
}

func TestGetOneClientModel_NotFound(t *testing.T) {
    mm := newMockManager()
    models := []*modelpb.Model{
        {Id: "m1", ClientId: "c1", ProviderId: "p1"},
    }
    mm.listAll[cachetypes.ItemTypeModel] = struct{
        total uint64
        value any
        err   error
    }{value: models}
    mm.listAll[cachetypes.ItemTypeClientModelRelation] = struct{
        total uint64
        value any
        err   error
    }{value: []*clientmodelrelationpb.ClientModelRelation{}}
    mm.getByID[cachetypes.ItemTypeProvider] = map[string]struct{ value any; err error }{
        "p1": {value: &providerpb.ServiceProvider{Id: "p1"}},
    }
    ctx := withCacheManager(t, mm)
    if _, err := GetOneClientModel(ctx, "c1", "not-exist"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}

