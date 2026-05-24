# Test Infrastructure Refactoring: horizontal_test.go

## Overview

The core change replaces a complex, stateful, concurrency-heavy test harness with a simple synchronous approach:

- **Before**: `testCase.prepareTestClient()` → `testCase.setupController()` → `testCase.runTest()` (starts full controller loop)
- **After**: `newFakeHorizontalClient()` → `newHorizontalSetup()` → direct `reconcileAutoscaler()` call

---

## The Old Architecture: `testCase` + `runTest`

### `testCase` struct (~50 fields, embeds `sync.Mutex`)

The struct mixed three different concerns into one object:

1. **Input data** — scenario configuration (`reportedLevels`, `CPUTarget`, `specReplicas`, etc.)
2. **Mutable runtime state** — modified by reactors during execution (`scaleUpdated`, `statusUpdated`, `eventCreated`, `processed` channel)
3. **Expected outcomes** — used in assertions (`expectedDesiredReplicas`, `expectedConditions`, `verifyCPUCurrent`, etc.)

The mutex was necessary because fake client reactors run on controller goroutines and mutate the same struct that the test thread reads for assertions.

### `prepareTestClient` (~500 lines)

This was the most complex part. It set up reactors for all five fake clients (Clientset, MetricsClient, CustomMetricsClient, ExternalMetricsClient, ScaleClient), with critical characteristics:

**The HPA object was constructed inside a reactor**, not ahead of time:
```go
fakeClient.AddReactor("list", "horizontalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
    tc.Lock()
    defer tc.Unlock()
    // ... builds the HPA from tc fields here ...
    return true, obj, nil
})
```

This meant the HPA only existed when the controller's informer listed it — you couldn't inspect the HPA independently.

**Assertions were embedded inside reactors** (not after the test action):
```go
fakeScaleClient.AddReactor("update", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
    tc.Lock()
    defer tc.Unlock()
    replicas := action.(core.UpdateAction).GetObject().(*autoscalingv1.Scale).Spec.Replicas
    assert.Equal(t, tc.expectedDesiredReplicas, replicas)  // assertion buried in reactor!
    tc.scaleUpdated = true
    return true, obj, nil
})
```

Similarly for the HPA status update reactor:
```go
fakeClient.AddReactor("update", "horizontalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
    // ... assertions about desired replicas, CPU utilization, conditions ...
    tc.statusUpdated = true
    tc.processed <- obj.Name  // signal that reconciliation completed
    return true, obj, nil
})
```

**Every reactor acquired a lock**, because the controller's goroutines call into these reactors concurrently:
```go
fakeScaleClient.AddReactor("get", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
    tc.Lock()
    defer tc.Unlock()
    // ...
})
```

**Duplicate reactors** existed for each resource type (replicationcontrollers, deployments, replicasets) even though they did the same thing — returning/asserting the same scale object.

### `runTest` / `runTestWithController`

```go
func (tc *testCase) runTest(t *testing.T) {
    hpaController, informerFactory := tc.setupController(t)
    tc.runTestWithController(t, hpaController, informerFactory)
}
```

`runTestWithController` started the full controller loop with 5 workers and used channel-based polling to detect completion:

```go
func (tc *testCase) runTestWithController(t *testing.T, hpaController *HorizontalController, informerFactory informers.SharedInformerFactory) {
    ctx, cancel := context.WithCancel(context.Background())
    informerFactory.Start(ctx.Done())

    var wg sync.WaitGroup
    wg.Go(func() {
        hpaController.Run(ctx, 5)  // start 5 workers
    })
    defer wg.Wait()
    defer cancel()

    // ... wait for tc.processed channel or timeout ...
    tc.verifyResults(ctx, t)
}
```

The waiting logic had two modes:
- **Event verification**: sleep for 2 seconds, draining the channel
- **Normal**: poll the `tc.processed` channel with 100ms timeouts, waiting for N reconciliations

### Problems with this approach

1. **Race conditions**: Shared mutable state accessed from controller goroutines and the test goroutine, requiring pervasive locking.
2. **Hidden assertions**: Key checks are buried inside reactor callbacks, not at the test level — failures point at reactor code, not the test case.
3. **Slow**: Every test starts 5 workers, waits for informer sync, then polls a channel. Event tests sleep 2 full seconds.
4. **Opaque flow**: To understand what a test checks, you must trace through `prepareTestClient` (500 lines), `setupController`, `runTestWithController`, and `verifyResults`.
5. **Fragile**: The `processed` channel signal comes from the status-update reactor, so if the controller path doesn't update status, the test hangs until timeout.

---

## The New Architecture: `horizontalScenario` + `reconcileAutoscaler`

### `horizontalScenario` struct (~20 fields, no mutex)

Pure input data — only describes the state of the fake cluster:

```go
type horizontalScenario struct {
    minReplicas    int32
    maxReplicas    int32
    specReplicas   int32
    statusReplicas int32
    CPUTarget      int32
    reportedLevels      []uint64
    reportedCPURequests []resource.Quantity
    metricsTarget       []autoscalingv2.MetricSpec
    // ... pod state fields ...
    resource      *fakeResource
    lastScaleTime *metav1.Time
}
```

No mutable state, no expected outcomes, no mutex.

### `buildHPA` (standalone function)

The HPA is constructed explicitly, up front, and returned for direct use:

```go
func buildHPA(t *testing.T, s *horizontalScenario) *autoscalingv2.HorizontalPodAutoscaler {
    // ... builds HPA from scenario fields ...
    return hpa
}
```

This replaces the HPA being hidden inside a `list` reactor.

### `newFakeHorizontalClient`

Sets up the same five fake clients, but with critical simplifications:

1. **No mutex** — reactors just return data; no shared mutable state.
2. **No assertions inside reactors** — reactors only return fake data.
3. **No signal channels** — no `tc.processed` channel to coordinate.
4. **Data comes from the scenario struct** (read-only), not a locked mutable struct.

```go
func newFakeHorizontalClient(t *testing.T, s *horizontalScenario) (...) {
    fakeClient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
        // just builds pods from s.reportedLevels, s.reportedCPURequests, etc.
        // no lock, no assertions, no state mutation
        return true, obj, nil
    })
    // ...
}
```

### `newHorizontalSetup`

Creates the controller and returns a `horizontalSetup` struct with all clients accessible:

```go
type horizontalSetup struct {
    controller      *HorizontalController
    informerFactory informers.SharedInformerFactory
    scaleClient     *scalefake.FakeScaleClient
    testClient      *fake.Clientset
    eventClient     *fake.Clientset
    metricsClient   *metricsfake.Clientset
    cmClient        *cmfake.FakeCustomMetricsClient
    emClient        *emfake.FakeExternalMetricsClient
    ctx             context.Context
}
```

### Direct `reconcileAutoscaler` call (replaces `runTest`)

Instead of starting the full controller loop:

```go
setup := newHorizontalSetup(t, &tt.fixture)
hpa := buildHPA(t, &tt.fixture)
key := fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)

err := setup.controller.reconcileAutoscaler(setup.ctx, hpa, key)
require.NoError(t, err)
```

This is a synchronous, single-threaded call. No goroutines, no channels, no timeouts.

### Assertions are at the test level

After `reconcileAutoscaler` returns, tests inspect the fake client's recorded actions:

```go
scaleUpdated := false
for _, action := range setup.scaleClient.Actions() {
    if action.GetVerb() == "update" {
        scaleUpdated = true
        scale := action.(core.UpdateAction).GetObject().(*autoscalingv1.Scale)
        assert.Equal(t, tt.expectedDesiredReplicas, scale.Spec.Replicas)
    }
}
assert.Equal(t, tt.expectedScaleUpdated, scaleUpdated)
```

---

## Side-by-Side: How a test looked before vs. after

### Before (one test per scenario)

```go
func TestScaleUp(t *testing.T) {
    tc := testCase{
        minReplicas:             2,
        maxReplicas:             6,
        specReplicas:            3,
        statusReplicas:          3,
        expectedDesiredReplicas: 5,
        CPUTarget:               30,
        verifyCPUCurrent:        true,
        reportedLevels:          []uint64{300, 500, 700},
        reportedCPURequests:     []resource.Quantity{...},
        useMetricsAPI:           true,
        expectedReportedReconciliationActionLabel: monitor.ActionLabelScaleUp,
        expectedReportedReconciliationErrorLabel:  monitor.ErrorLabelNone,
        expectedReportedMetricComputationActionLabels: map[...]{...},
        expectedReportedMetricComputationErrorLabels:  map[...]{...},
    }
    tc.runTest(t)
}
```

### After (table-driven)

```go
func TestScaleCPU(t *testing.T) {
    tests := []struct {
        name                    string
        fixture                 horizontalScenario
        expectedDesiredReplicas int32
        expectedScaleUpdated    bool
        expectedActionLabel     monitor.ActionLabel
    }{
        {
            name: "scale up",
            fixture: horizontalScenario{
                minReplicas:         2,
                maxReplicas:         6,
                specReplicas:        3,
                statusReplicas:      3,
                CPUTarget:           30,
                reportedLevels:      []uint64{300, 500, 700},
                reportedCPURequests: []resource.Quantity{...},
            },
            expectedDesiredReplicas: 5,
            expectedScaleUpdated:    true,
            expectedActionLabel:     monitor.ActionLabelScaleUp,
        },
        // ... more cases ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            setup := newHorizontalSetup(t, &tt.fixture)
            hpa := buildHPA(t, &tt.fixture)
            key := fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)

            err := setup.controller.reconcileAutoscaler(setup.ctx, hpa, key)
            require.NoError(t, err)

            // explicit assertions here ...
        })
    }
}
```

---

## Summary of Key Improvements

| Aspect | Before | After |
|--------|--------|-------|
| Execution | Full controller loop (5 workers) | Direct `reconcileAutoscaler` call |
| Concurrency | Mutex-protected shared state across goroutines | Single-threaded, no locks |
| HPA construction | Hidden inside `list` reactor | Explicit `buildHPA` function |
| Assertions | Scattered across reactor callbacks + `verifyResults` | Inline at test level after the call |
| Synchronization | Channel polling + sleep timeouts | None needed |
| Test structure | One top-level `func Test*` per scenario | Table-driven subtests |
| Adding a test | Copy entire `func Test*` + fill ~20 fields including expected metric labels | Add a struct literal to the `tests` slice |
