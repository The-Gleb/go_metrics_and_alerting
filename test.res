?   	github.com/The-Gleb/go_metrics_and_alerting/cmd/agent	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/cmd/server	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/cmd/staticlint	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/middleware	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/logger	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/repository	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/repository/file_storage	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/internal/server	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/pkg/client	[no test files]
?   	github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry	[no test files]
=== RUN   Test_getAllMetricHandler_ServeHTTP
=== RUN   Test_getAllMetricHandler_ServeHTTP/normal
    get_all_metrics_test.go:89: {"Gauge":[{"id":"HeapAlloc","type":"gauge","value":3782369280}],"Counter":[{"id":"PollCount","type":"counter","delta":123}]}
--- PASS: Test_getAllMetricHandler_ServeHTTP (0.00s)
    --- PASS: Test_getAllMetricHandler_ServeHTTP/normal (0.00s)
=== RUN   Test_getMetricJSONHandler_ServeHTTP
--- PASS: Test_getMetricJSONHandler_ServeHTTP (0.00s)
=== RUN   Test_getMetricHandler_ServeHTTP
=== RUN   Test_getMetricHandler_ServeHTTP/normal_gauge_test_#1
    get_metric_test.go:83: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/get_metric_test.go:83
        	Error:      	Not equal: 
        	            	expected: 200
        	            	actual  : 400
        	Test:       	Test_getMetricHandler_ServeHTTP/normal_gauge_test_#1
=== RUN   Test_getMetricHandler_ServeHTTP/normal_counter_test_#2
    get_metric_test.go:83: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/get_metric_test.go:83
        	Error:      	Not equal: 
        	            	expected: 200
        	            	actual  : 400
        	Test:       	Test_getMetricHandler_ServeHTTP/normal_counter_test_#2
=== RUN   Test_getMetricHandler_ServeHTTP/neg_counter_test_#3
    get_metric_test.go:83: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/get_metric_test.go:83
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 400
        	Test:       	Test_getMetricHandler_ServeHTTP/neg_counter_test_#3
=== RUN   Test_getMetricHandler_ServeHTTP/wrong_metric_type_test_#4
=== RUN   Test_getMetricHandler_ServeHTTP/empty_params
--- FAIL: Test_getMetricHandler_ServeHTTP (0.00s)
    --- FAIL: Test_getMetricHandler_ServeHTTP/normal_gauge_test_#1 (0.00s)
    --- FAIL: Test_getMetricHandler_ServeHTTP/normal_counter_test_#2 (0.00s)
    --- FAIL: Test_getMetricHandler_ServeHTTP/neg_counter_test_#3 (0.00s)
    --- PASS: Test_getMetricHandler_ServeHTTP/wrong_metric_type_test_#4 (0.00s)
    --- PASS: Test_getMetricHandler_ServeHTTP/empty_params (0.00s)
=== RUN   Test_updateMetricSetHandler_ServeHTTP
=== RUN   Test_updateMetricSetHandler_ServeHTTP/normal
    update_metric_set_test.go:65: 2 metrics updated
=== RUN   Test_updateMetricSetHandler_ServeHTTP/request_with_invalid_body
    update_metric_set_test.go:65: invalid character 's' looking for beginning of value
        
--- PASS: Test_updateMetricSetHandler_ServeHTTP (0.00s)
    --- PASS: Test_updateMetricSetHandler_ServeHTTP/normal (0.00s)
    --- PASS: Test_updateMetricSetHandler_ServeHTTP/request_with_invalid_body (0.00s)
=== RUN   Test_updateMetricHandler_ServeHTTP
=== RUN   Test_updateMetricHandler_ServeHTTP/normal_gauge_test_#1
    update_metric_test.go:113: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/update_metric_test.go:113
        	Error:      	Not equal: 
        	            	expected: "23.23"
        	            	actual  : "0xc00028e1a8"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-23.23
        	            	+0xc00028e1a8
        	Test:       	Test_updateMetricHandler_ServeHTTP/normal_gauge_test_#1
=== RUN   Test_updateMetricHandler_ServeHTTP/first_add_counter_test_#2
    update_metric_test.go:113: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/update_metric_test.go:113
        	Error:      	Not equal: 
        	            	expected: "23"
        	            	actual  : "0xc00028e1f0"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-23
        	            	+0xc00028e1f0
        	Test:       	Test_updateMetricHandler_ServeHTTP/first_add_counter_test_#2
=== RUN   Test_updateMetricHandler_ServeHTTP/second_add_counter_test_#3
    update_metric_test.go:113: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler/update_metric_test.go:113
        	Error:      	Not equal: 
        	            	expected: "30"
        	            	actual  : "0xc000015258"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-30
        	            	+0xc000015258
        	Test:       	Test_updateMetricHandler_ServeHTTP/second_add_counter_test_#3
=== RUN   Test_updateMetricHandler_ServeHTTP/name_and_value_not_sent_-_test_#4
=== RUN   Test_updateMetricHandler_ServeHTTP/value_not_sent_-_test_#5
=== RUN   Test_updateMetricHandler_ServeHTTP/wrong_metric_type-_test_#5
=== RUN   Test_updateMetricHandler_ServeHTTP/incorrect_metric_value_type_-_test_#6
=== RUN   Test_updateMetricHandler_ServeHTTP/empty_metric_name_-_test_#7
--- FAIL: Test_updateMetricHandler_ServeHTTP (0.00s)
    --- FAIL: Test_updateMetricHandler_ServeHTTP/normal_gauge_test_#1 (0.00s)
    --- FAIL: Test_updateMetricHandler_ServeHTTP/first_add_counter_test_#2 (0.00s)
    --- FAIL: Test_updateMetricHandler_ServeHTTP/second_add_counter_test_#3 (0.00s)
    --- PASS: Test_updateMetricHandler_ServeHTTP/name_and_value_not_sent_-_test_#4 (0.00s)
    --- PASS: Test_updateMetricHandler_ServeHTTP/value_not_sent_-_test_#5 (0.00s)
    --- PASS: Test_updateMetricHandler_ServeHTTP/wrong_metric_type-_test_#5 (0.00s)
    --- PASS: Test_updateMetricHandler_ServeHTTP/incorrect_metric_value_type_-_test_#6 (0.00s)
    --- PASS: Test_updateMetricHandler_ServeHTTP/empty_metric_name_-_test_#7 (0.00s)
=== RUN   Example_getAllMetricHandler_ServeHTTP
--- PASS: Example_getAllMetricHandler_ServeHTTP (0.00s)
=== RUN   Example_getMetricJSONHandler_ServeHTTP
--- FAIL: Example_getMetricJSONHandler_ServeHTTP (0.00s)
got:
400
invalid entity.GetMetricDTO: 1 problems
want:
200
{"id":"Alloc","type":"gauge","value":3782369280}
=== RUN   Example_getMetricHandler_ServeHTTP
--- FAIL: Example_getMetricHandler_ServeHTTP (0.00s)
got:
400
invalid metric struct, some fields are empty, but they shouldn`t
want:
200
123.4
=== RUN   Example_updateMetricJSONHandler_ServeHTTP
--- PASS: Example_updateMetricJSONHandler_ServeHTTP (0.00s)
=== RUN   Example_updateMetricSetHandler_ServeHTTP
--- PASS: Example_updateMetricSetHandler_ServeHTTP (0.00s)
=== RUN   Example_updateMetricHandler_ServeHTTP
--- FAIL: Example_updateMetricHandler_ServeHTTP (0.00s)
got:
200
0xc000308368
want:
200
12.12
FAIL
FAIL	github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler	0.007s
=== RUN   Test_metricService_UpdateMetric
=== RUN   Test_metricService_UpdateMetric/positive_add_gauge
=== RUN   Test_metricService_UpdateMetric/positive_update_gauge
=== RUN   Test_metricService_UpdateMetric/positive_add_counter
=== RUN   Test_metricService_UpdateMetric/positive_update_counter
=== RUN   Test_metricService_UpdateMetric/negative,_gauge,_empty_metric.Value
=== RUN   Test_metricService_UpdateMetric/negative,_counter,_empty_metric.Delta
--- PASS: Test_metricService_UpdateMetric (0.00s)
    --- PASS: Test_metricService_UpdateMetric/positive_add_gauge (0.00s)
    --- PASS: Test_metricService_UpdateMetric/positive_update_gauge (0.00s)
    --- PASS: Test_metricService_UpdateMetric/positive_add_counter (0.00s)
    --- PASS: Test_metricService_UpdateMetric/positive_update_counter (0.00s)
    --- PASS: Test_metricService_UpdateMetric/negative,_gauge,_empty_metric.Value (0.00s)
    --- PASS: Test_metricService_UpdateMetric/negative,_counter,_empty_metric.Delta (0.00s)
=== RUN   Test_metricService_UpdateMetricSet
=== RUN   Test_metricService_UpdateMetricSet/first_insert
=== RUN   Test_metricService_UpdateMetricSet/update_metrics
=== RUN   Test_metricService_UpdateMetricSet/invalid_metric_struct
--- PASS: Test_metricService_UpdateMetricSet (0.00s)
    --- PASS: Test_metricService_UpdateMetricSet/first_insert (0.00s)
    --- PASS: Test_metricService_UpdateMetricSet/update_metrics (0.00s)
    --- PASS: Test_metricService_UpdateMetricSet/invalid_metric_struct (0.00s)
=== RUN   Test_metricService_GetMetric
=== RUN   Test_metricService_GetMetric/pos_gauge_test_#1
    metric_test.go:202: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/domain/service/metric_test.go:202
        	Error:      	Not equal: 
        	            	expected: entity.Metric{ID:"Alloc", MType:"gauge", Delta:(*int64)(nil), Value:(*float64)(0xc000014c48)}
        	            	actual  : entity.Metric{ID:"", MType:"", Delta:(*int64)(nil), Value:(*float64)(nil)}
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1,6 +1,6 @@
        	            	 (entity.Metric) {
        	            	- ID: (string) (len=5) "Alloc",
        	            	- MType: (string) (len=5) "gauge",
        	            	+ ID: (string) "",
        	            	+ MType: (string) "",
        	            	  Delta: (*int64)(<nil>),
        	            	- Value: (*float64)(12345)
        	            	+ Value: (*float64)(<nil>)
        	            	 }
        	Test:       	Test_metricService_GetMetric/pos_gauge_test_#1
=== RUN   Test_metricService_GetMetric/pos_counter_test_#2
    metric_test.go:202: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/domain/service/metric_test.go:202
        	Error:      	Not equal: 
        	            	expected: entity.Metric{ID:"PollCount", MType:"counter", Delta:(*int64)(0xc000014c50), Value:(*float64)(nil)}
        	            	actual  : entity.Metric{ID:"", MType:"", Delta:(*int64)(nil), Value:(*float64)(nil)}
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1,5 +1,5 @@
        	            	 (entity.Metric) {
        	            	- ID: (string) (len=9) "PollCount",
        	            	- MType: (string) (len=7) "counter",
        	            	- Delta: (*int64)(123),
        	            	+ ID: (string) "",
        	            	+ MType: (string) "",
        	            	+ Delta: (*int64)(<nil>),
        	            	  Value: (*float64)(<nil>)
        	Test:       	Test_metricService_GetMetric/pos_counter_test_#2
=== RUN   Test_metricService_GetMetric/neg,_metric_not_found
    metric_test.go:199: 
        	Error Trace:	/home/bp/go/src/praktikum/github.com/reviews/go_24/The-Gleb/go_metrics_and_alerting/internal/domain/service/metric_test.go:199
        	Error:      	Target error should be in err chain:
        	            	expected: "metric name not found"
        	            	in chain: "invalid metric struct, some fields are empty, but they shouldn`t"
        	Test:       	Test_metricService_GetMetric/neg,_metric_not_found
=== RUN   Test_metricService_GetMetric/invalid_type
--- FAIL: Test_metricService_GetMetric (0.00s)
    --- FAIL: Test_metricService_GetMetric/pos_gauge_test_#1 (0.00s)
    --- FAIL: Test_metricService_GetMetric/pos_counter_test_#2 (0.00s)
    --- FAIL: Test_metricService_GetMetric/neg,_metric_not_found (0.00s)
    --- PASS: Test_metricService_GetMetric/invalid_type (0.00s)
FAIL
FAIL	github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service	0.002s
=== RUN   Test_getAllMetricsUsecase_GetAllMetricsJSON
=== RUN   Test_getAllMetricsUsecase_GetAllMetricsJSON/positive
--- PASS: Test_getAllMetricsUsecase_GetAllMetricsJSON (0.00s)
    --- PASS: Test_getAllMetricsUsecase_GetAllMetricsJSON/positive (0.00s)
PASS
ok  	github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase	(cached)
=== RUN   TestDB_UpdateMetricSet
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_UpdateMetricSet/first_insert
=== RUN   TestDB_UpdateMetricSet/update_metrics
=== RUN   TestDB_UpdateMetricSet/invalid_metric_struct
--- PASS: TestDB_UpdateMetricSet (0.06s)
    --- PASS: TestDB_UpdateMetricSet/first_insert (0.00s)
    --- PASS: TestDB_UpdateMetricSet/update_metrics (0.00s)
    --- PASS: TestDB_UpdateMetricSet/invalid_metric_struct (0.00s)
=== RUN   TestDB_GetAllMetrics
    storage_test.go:125: before new metric db 1
    storage_test.go:130: before new metric db
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_GetAllMetrics/positive
--- PASS: TestDB_GetAllMetrics (0.03s)
    --- PASS: TestDB_GetAllMetrics/positive (0.00s)
=== RUN   TestDB_UpdateGauge
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_UpdateGauge/first_insert
=== RUN   TestDB_UpdateGauge/update_metrics
--- PASS: TestDB_UpdateGauge (0.09s)
    --- PASS: TestDB_UpdateGauge/first_insert (0.00s)
    --- PASS: TestDB_UpdateGauge/update_metrics (0.00s)
=== RUN   TestDB_UpdateCounter
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_UpdateCounter/first_insert
=== RUN   TestDB_UpdateCounter/update_metrics
--- PASS: TestDB_UpdateCounter (0.03s)
    --- PASS: TestDB_UpdateCounter/first_insert (0.00s)
    --- PASS: TestDB_UpdateCounter/update_metrics (0.00s)
=== RUN   TestDB_GetGauge
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_GetGauge/positive
=== RUN   TestDB_GetGauge/metric_doesn`t_exists
--- PASS: TestDB_GetGauge (0.03s)
    --- PASS: TestDB_GetGauge/positive (0.00s)
    --- PASS: TestDB_GetGauge/metric_doesn`t_exists (0.00s)
=== RUN   TestDB_GetCounter
2024/03/12 20:43:02 in new metrics db
2024/03/12 20:43:02 walking dir:  {0xa21c60}
2024/03/12 20:43:02 .
2024/03/12 20:43:02 migration
2024/03/12 20:43:02 migration/000001_init_schema.down.sql
2024/03/12 20:43:02 migration/000001_init_schema.up.sql
2024/03/12 20:43:02 ERROR no change
=== RUN   TestDB_GetCounter/positive
=== RUN   TestDB_GetCounter/metric_doesn`t_exists
--- PASS: TestDB_GetCounter (0.14s)
    --- PASS: TestDB_GetCounter/positive (0.00s)
    --- PASS: TestDB_GetCounter/metric_doesn`t_exists (0.00s)
PASS
ok  	github.com/The-Gleb/go_metrics_and_alerting/internal/repository/database	(cached)
=== RUN   Test_storage_GetMetric
=== RUN   Test_storage_GetMetric/pos_gauge_test_#1
=== RUN   Test_storage_GetMetric/pos_counter_test_#2
=== RUN   Test_storage_GetMetric/neg_gauge_test_#3
=== RUN   Test_storage_GetMetric/neg_bad_request_test_#4
--- PASS: Test_storage_GetMetric (0.00s)
    --- PASS: Test_storage_GetMetric/pos_gauge_test_#1 (0.00s)
    --- PASS: Test_storage_GetMetric/pos_counter_test_#2 (0.00s)
    --- PASS: Test_storage_GetMetric/neg_gauge_test_#3 (0.00s)
    --- PASS: Test_storage_GetMetric/neg_bad_request_test_#4 (0.00s)
=== RUN   Test_storage_UpdateMetric
=== RUN   Test_storage_UpdateMetric/pos_new_counter_test_#1
=== RUN   Test_storage_UpdateMetric/pos_update_counter_test_#2
=== RUN   Test_storage_UpdateMetric/pos_new_gauge_test_#3
=== RUN   Test_storage_UpdateMetric/pos_update_gauge_test_#4
=== RUN   Test_storage_UpdateMetric/neg_gauge_test_#5
=== RUN   Test_storage_UpdateMetric/neg_counter_test_#6
=== RUN   Test_storage_UpdateMetric/wrong_metric_type_test_#7
--- PASS: Test_storage_UpdateMetric (0.00s)
    --- PASS: Test_storage_UpdateMetric/pos_new_counter_test_#1 (0.00s)
    --- PASS: Test_storage_UpdateMetric/pos_update_counter_test_#2 (0.00s)
    --- PASS: Test_storage_UpdateMetric/pos_new_gauge_test_#3 (0.00s)
    --- PASS: Test_storage_UpdateMetric/pos_update_gauge_test_#4 (0.00s)
    --- PASS: Test_storage_UpdateMetric/neg_gauge_test_#5 (0.00s)
    --- PASS: Test_storage_UpdateMetric/neg_counter_test_#6 (0.00s)
    --- PASS: Test_storage_UpdateMetric/wrong_metric_type_test_#7 (0.00s)
=== RUN   Test_storage_GetAllMetrics
--- PASS: Test_storage_GetAllMetrics (0.00s)
PASS
ok  	github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory	(cached)
FAIL
