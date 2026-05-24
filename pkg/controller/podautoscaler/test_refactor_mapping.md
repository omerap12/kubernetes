# Test Refactor Mapping: horizontal_test.go

This document maps the old standalone test functions to their new table-driven locations.

## Consolidated Tests

| Old Test (master) | New Test Function | Subtest Name |
|---|---|---|
| `TestScaleUp` | `TestScaleCPU` | "scale up" |
| `TestScaleUpUnreadyLessScale` | `TestScaleCPU` | "scale up with unready pods results in less scale" |
| `TestScaleUpHotCpuLessScale` | `TestScaleCPU` | "scale up with hot cpu pods results in less scale" |
| `TestScaleUpIgnoresFailedPods` | `TestScaleCPU` | "scale up ignores failed pods" |
| `TestScaleUpIgnoresDeletionPods` | `TestScaleCPU` | "scale up ignores deletion pods" |
| `TestScaleUpDeployment` | `TestScaleCPU` | "scale up deployment" |
| `TestScaleUpReplicaSet` | `TestScaleCPU` | "scale up replicaset" |
| `TestScaleUpUnreadyNoScale` | `TestScaleCPU` | "no scale up when all unready pods bring utilization within target" |
| `TestScaleDown` | `TestScaleCPU` | "scale down" |
| `TestScaleDownIncludeUnreadyPods` | `TestScaleCPU` | "scale down includes unready pods" |
| `TestScaleDownIgnoreHotCpuPods` | `TestScaleCPU` | "scale down ignores hot cpu pods" |
| `TestScaleDownIgnoresFailedPods` | `TestScaleCPU` | "scale down ignores failed pods" |
| `TestScaleDownIgnoresDeletionPods` | `TestScaleCPU` | "scale down ignores deletion pods" |
| `TestScaleUpContainer` | `TestScaleContainerResource` | "scale up container resource" |
| `TestScaleDownContainerResource` | `TestScaleContainerResource` | "scale down container resource" |
| `TestScaleUpHotCpuNoScale` | `TestScaleUpUnreadyOrHotCpuNoScale` | "hot CPU pods excluded from scaling" |
| `TestScaleUpCMUnreadyandCpuHot` | `TestScaleUpUnreadyOrHotCpuNoScale` | "unready pods excluded from scaling" |
| `TestScaleUpCM` | `TestScaleUpCM` | "scale up custom metric" |
| `TestScaleUpCMUnreadyAndHotCpuNoLessScale` | `TestScaleUpCM` | "unready and hot cpu pods do not reduce scale for custom metric" / "unready and hot cpu pods scale to max with scaling limited condition" |
| `TestScaleUpCMExternal` | `TestScaleUpCM` | "scale up external metric with value target" |
| `TestScaleUpPerPodCMExternal` | `TestScaleUpCM` | "scale up external metric with per-pod average value target" |
| `TestScaleUpCMObject` | `TestScaleUpCMObject` | "scale up object metric with value target" |
| `TestScaleUpFromZeroCMObject` | `TestScaleUpCMObject` | "scale up from zero with feature gate enabled" / "scale up from zero with feature gate disabled" |
| `TestScaleUpFromZeroIgnoresToleranceCMObject` | `TestScaleUpCMObject` | "scale up from zero ignores tolerance with feature gate enabled" / "scale up from zero ignores tolerance with feature gate disabled" |
| `TestScaleUpPerPodCMObject` | `TestScaleUpCMObject` | "scale up per-pod object metric with average value target" |
| `TestScaleUpOneMetricInvalid` | `TestScaleWithOneInvalidMetric` | "scale up with one invalid metric" |
| `TestScaleUpFromZeroOneMetricInvalid` | `TestScaleWithOneInvalidMetric` | "from zero with invalid metric, FG on" / "from zero with invalid metric, FG off" |
| `TestScaleDownCMObject` | `TestScaleDownCM` | "scale down object metric with value target" |
| `TestScaleDownToZeroCMObject` | `TestScaleDownCM` | "scale down to zero object metric with feature gate enabled" / "scale down to zero object metric with feature gate disabled" |
| `TestScaleDownPerPodCMObject` | `TestScaleDownCM` | "scale down per-pod object metric with average value target" |
| `TestScaleDownCMExternal` | `TestScaleDownCM` | "scale down external metric with value target" |
| `TestScaleDownToZeroCMExternal` | `TestScaleDownCM` | "scale down to zero external metric with feature gate enabled" / "scale down to zero external metric with feature gate disabled" |
| `TestScaleDownPerPodCMExternal` | `TestScaleDownCM` | "scale down per-pod external metric with average value target" |
| `TestScaleUpFromZeroWithoutCondition` | `TestScaleToZeroBehavior` | "from zero without condition, FG on" / "from zero without condition, FG off" |
| `TestScaleUpFromZeroWhenMinReplicasIncreased` | `TestScaleToZeroBehavior` | "minReplicas increased with ScaledToZero condition, FG on" / "minReplicas increased with ScaledToZero condition, FG off" |
| `TestScaledToZeroConditionHandledOnNormalRescale` | `TestScaleToZeroBehavior` | "stale ScaledToZero condition cleared on normal rescale, FG on" / "stale ScaledToZero condition removed on normal rescale, FG off" |
| `TestManualScaleToZeroDisablesHPA` | `TestScaleToZeroBehavior` | "manual scale to zero disables HPA, FG on" / "manual scale to zero disables HPA, FG off" |
| `TestTolerance` | `TestTolerance` | "CPU resource metric within tolerance" |
| `TestToleranceCM` | `TestTolerance` | "pods metric within tolerance" |
| `TestToleranceCMObject` | `TestTolerance` | "object metric with value target within tolerance" |
| `TestToleranceCMExternal` | `TestTolerance` | "external metric with value target within tolerance" |
| `TestTolerancePerPodCMObject` | `TestTolerance` | "per-pod object metric with average value target within tolerance" |
| `TestTolerancePerPodCMExternal` | `TestTolerance` | "per-pod external metric with average value target within tolerance" |
| `TestConfigurableTolerance` | `TestConfigurableTolerance` | "scaling up because of a 1% configurable tolerance" / "no scale-down because of a 90% configurable tolerance" / "no scaling because of the large default tolerance" / "no scaling because the configurable tolerance is ignored when feature gate is disabled" |
| `TestMinReplicas` | `TestReplicaLimits` | "scale down clamped to min replicas" |
| `TestZeroMinReplicasDesiredZero` | `TestReplicaLimits` | "scale down to zero when min replicas is zero" |
| `TestMinReplicasDesiredZero` | `TestReplicaLimits` | "scale down clamped to min when desired is zero" |
| `TestZeroReplicas` | `TestReplicaLimits` | "no scaling when spec replicas is zero" |
| `TestTooFewReplicas` | `TestReplicaLimits` | "scale up to min when below min replicas" |
| `TestTooManyReplicas` | `TestReplicaLimits` | "scale down to max when above max replicas" |
| `TestMaxReplicas` | `TestReplicaLimits` | "scale up clamped to max replicas" |
| `TestScaleUpRCImmediately` | `TestReplicaLimits` | "scale up to min immediately even with recent scale time" |
| `TestScaleDownRCImmediately` | `TestReplicaLimits` | "scale down to max immediately even with recent scale time" |
| `TestSuperfluousMetrics` | `TestMetricsEdgeCases` | "superfluous metrics reported" |
| `TestMissingMetrics` | `TestMetricsEdgeCases` | "missing metrics from some pods" |
| `TestEmptyMetrics` | `TestMetricsEdgeCases` | "empty metrics from all pods" |
| `TestEmptyCPURequest` | `TestMetricsEdgeCases` | "empty CPU requests" |
| `TestMissingReports` | `TestMetricsEdgeCases` | "missing reports from some pods scales down" |
| `TestMoreReplicasThanSpecNoScale` | `TestMetricsEdgeCases` | "more replicas than spec no scale" |
| `TestEventCreated` | `TestEventCreation` | "event created on scale up" |
| `TestEventNotCreated` | `TestEventCreation` | "no event when replicas unchanged" |
| `TestUpscaleCapGreaterThanMaxReplicas` | `TestUpscaleCap` | "capped by scale up rules" / "capped by max replicas when scale up limit exceeds max" |
| `TestConditionInvalidSelectorMissing` | `TestConditionSelectorValidation` | "invalid selector missing" |
| `TestConditionInvalidSelectorUnparsable` | `TestConditionSelectorValidation` | "invalid selector unparsable" |
| `TestConditionNoAmbiguousSelectorWhenNoSelectorOverlapBetweenHPAs` | `TestConditionSelectorValidation` | "no ambiguous selector when no overlap between HPAs" |
| `TestConditionAmbiguousSelectorWhenFullSelectorOverlapBetweenHPAs` | `TestConditionSelectorValidation` | "ambiguous selector when full overlap between HPAs" |
| `TestConditionAmbiguousSelectorWhenPartialSelectorOverlapBetweenHPAs` | `TestConditionSelectorValidation` | "ambiguous selector when partial overlap between HPAs" |
| `TestConditionFailedGetMetrics` | `TestConditionFailedGetMetrics` | "failed get resource metric" / "failed get pods metric" / "failed get object metric" / "failed get external metric" |
| `TestConditionInvalidSourceType` | `TestInvalidMetricSourceType` | "only invalid metric source type" |
| `TestNoScaleDownOneMetricInvalid` | `TestInvalidMetricSourceType` | "no scale down when one metric has invalid source type" |
| `TestConditionFailedGetScale` | `TestConditionFailedScale` | "failed get scale" |
| `TestConditionFailedUpdateScale` | `TestConditionFailedScale` | "failed update scale" |
| `TestNoBackoffUpscaleCM` | `TestScaleTimingBehavior` | "no backoff upscale custom metric only" |
| `TestNoBackoffUpscaleCMNoBackoffCpu` | `TestScaleTimingBehavior` | "no backoff upscale custom metric and CPU" |
| `TestStabilizeDownscale` | `TestScaleTimingBehavior` | "stabilize downscale" |
| `TestScaleUpOneMetricEmpty` | `TestOneMetricEmptyExternalError` | "scale up when one external metric errors" |
| `TestNoScaleDownOneMetricEmpty` | `TestOneMetricEmptyExternalError` | "no scale down when one external metric errors" |
| `TestReconciliationDurationIsRecorded` | `TestReconciliationDurationIsRecorded` | "duration recorded on successful scale up" |
| `TestReconciliationDurationIsRecordedOnError` | `TestReconciliationDurationIsRecorded` | "duration recorded on error" |

## Unchanged Tests

These tests were not modified by the refactor:

- `TestAvoidUnnecessaryUpdates`
- `TestBuildQuantity`
- `TestCalculateScaleDownLimitWithBehaviors`
- `TestCalculateScaleUpLimitWithScalingRules`
- `TestComputedToleranceAlgImplementation`
- `TestConvertDesiredReplicasWithRules`
- `TestDeleteHPAClearsConsistencyStore`
- `TestEnqueueHPAAddsImmediately`
- `TestEnqueueHPARegistersSelectorBeforeQueueAdd`
- `TestHPARescaleWithSuccessfulConflictRetry`
- `TestMultipleHPAs`
- `TestNormalizeDesiredReplicas`
- `TestNormalizeDesiredReplicasWithBehavior`
- `TestReconcileKeyEnsureReadyStaleCache`
- `TestReconciliationsTotalCountMultipleReconciliations`
- `TestScaleDownStabilizeInitialSize`
- `TestScalingWithRules`
- `TestStoreScaleEvents`
- `TestUpdateHPAEnqueueBehavior`
- `TestUpdateHPAFallsBackWhenFeatureDisabled`
- `TestUpdateStatusPopulatesConsistencyStore`
