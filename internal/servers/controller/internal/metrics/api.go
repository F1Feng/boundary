// Package metrics contains the singleton metric vectors and methods to access
// them through the controller code base.  Only exposing the metrics through
// their respective functions ensures they remain singletons and allows
// the code to enforce the appropriate labels are used.
package metrics

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/boundary/globals"
	"github.com/hashicorp/boundary/internal/gen/controller/api/services"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpcpb "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	pathRegex              map[*regexp.Regexp]string
	expectedPathsToMethods map[string][]string
)

func gatherPathInfo(rule *grpcpb.HttpRule, paths map[string][]string) {
	switch r := rule.GetPattern().(type) {
	case *grpcpb.HttpRule_Get:
		paths[r.Get] = append(paths[r.Get], http.MethodGet)
	case *grpcpb.HttpRule_Post:
		paths[r.Post] = append(paths[r.Post], http.MethodPost)
	case *grpcpb.HttpRule_Patch:
		paths[r.Patch] = append(paths[r.Patch], http.MethodPatch)
	case *grpcpb.HttpRule_Put:
		paths[r.Put] = append(paths[r.Put], http.MethodPut)
	case *grpcpb.HttpRule_Delete:
		paths[r.Delete] = append(paths[r.Delete], http.MethodDelete)
	default:
		panic("unknown rule of ")
	}
	for _, additional := range rule.AdditionalBindings {
		gatherPathInfo(additional, paths)
	}
}

func gatherServicePathsAndMethods(fd protoreflect.FileDescriptor, paths map[string][]string) error {
	for j := 0; j < fd.Services().Len(); j++ {
		sd := fd.Services().Get(j)
		for i := 0; i < sd.Methods().Len(); i++ {
			r := sd.Methods().Get(i)
			opts := r.Options().(*descriptorpb.MethodOptions)
			httpRule := proto.GetExtension(opts, grpcpb.E_Http).(*grpcpb.HttpRule)
			if proto.Equal(httpRule, &grpcpb.HttpRule{}) || httpRule == nil {
				return fmt.Errorf("empty or no http rule found on service method %q", r.FullName())
			}
			gatherPathInfo(httpRule, paths)
		}
	}
	return nil
}

// TODO: Auto generate the list of file descriptors so new services created
//  will automatically be covered by the metrics.
func apiPathsAndMethods() map[string][]string {
	fds := []protoreflect.FileDescriptor{
		services.File_controller_api_services_v1_account_service_proto,
		services.File_controller_api_services_v1_auth_method_service_proto,
		services.File_controller_api_services_v1_authtokens_service_proto,
		services.File_controller_api_services_v1_credential_library_service_proto,
		services.File_controller_api_services_v1_credential_store_service_proto,
		services.File_controller_api_services_v1_group_service_proto,
		services.File_controller_api_services_v1_host_catalog_service_proto,
		services.File_controller_api_services_v1_host_service_proto,
		services.File_controller_api_services_v1_host_set_service_proto,
		services.File_controller_api_services_v1_managed_group_service_proto,
		services.File_controller_api_services_v1_role_service_proto,
		services.File_controller_api_services_v1_scope_service_proto,
		services.File_controller_api_services_v1_session_service_proto,
		services.File_controller_api_services_v1_target_service_proto,
		services.File_controller_api_services_v1_user_service_proto,
	}
	paths := make(map[string][]string)
	for _, f := range fds {
		if err := gatherServicePathsAndMethods(f, paths); err != nil {
			panic(err)
		}
	}
	return paths
}

func init() {
	expectedPathsToMethods = apiPathsAndMethods()
	pathRegex = make(map[*regexp.Regexp]string)
	for p := range expectedPathsToMethods {
		pathRegex[buildRegexFromPath(p)] = p
	}
}

func buildRegexFromPath(p string) *regexp.Regexp {
	const idRegexp = "[[:alnum:]]{1,}_[[:alnum:]]{10,}"

	pWithId := strings.Replace(p, "{id}", idRegexp, 1)
	escaped := pWithId
	return regexp.MustCompile(fmt.Sprintf("^%s$", escaped))
}

const (
	invalidPathValue = "invalid"

	labelHttpCode   = "code"
	labelHttpPath   = "path"
	labelHttpMethod = "method"
	apiSubSystem    = "controller_api"
)

var (
	msgSizeBuckets = prometheus.ExponentialBuckets(100, 10, 8)

	// httpRequestLatency collects measurements of how long it takes
	// the boundary system to reply to a request to the controller api
	// from the time that boundary received the request.
	httpRequestLatency prometheus.ObserverVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: globals.MetricNamespace,
			Subsystem: apiSubSystem,
			Name:      "http_request_duration_seconds",
			Help:      "Histogram of latencies for HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{labelHttpCode, labelHttpPath, labelHttpMethod},
	)

	// httpRequestSize collections measurements of how large each request
	// to the boundary controller api is.
	httpRequestSize prometheus.ObserverVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: globals.MetricNamespace,
			Subsystem: apiSubSystem,
			Name:      "http_request_size_bytes",
			Help:      "Histogram of request sizes for HTTP requests.",
			// 100 bytes, 1kb, 10kb, 100kb, 1mb, 10mb, 100mb, 1gb
			Buckets: msgSizeBuckets,
		},
		[]string{labelHttpCode, labelHttpPath, labelHttpMethod},
	)

	// httpRequestSize collections measurements of how large each rresponse
	// from the boundary controller api is.
	httpResponseSize prometheus.ObserverVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: globals.MetricNamespace,
			Subsystem: apiSubSystem,
			Name:      "http_response_size_bytes",
			Help:      "Histogram of response sizes for HTTP responses.",
			// 100 bytes, 1kb, 10kb, 100kb, 1mb, 10mb, 100mb, 1gb
			Buckets: msgSizeBuckets,
		},
		[]string{labelHttpCode, labelHttpPath, labelHttpMethod},
	)
)

var universalStatusCodes = []int{
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusBadRequest,

	http.StatusInternalServerError,
	http.StatusGatewayTimeout,
}

// Codes which are only currently used in the authentication flow
var authenticationStatusCodes = []int{
	http.StatusAccepted,
	http.StatusFound,
}

var expectedStatusCodesPerMethod = map[string][]int{
	http.MethodGet: append(universalStatusCodes,
		http.StatusOK),
	http.MethodPost: append(universalStatusCodes,
		append(authenticationStatusCodes, http.StatusOK)...),
	http.MethodPatch: append(universalStatusCodes,
		http.StatusOK),

	// delete methods always returns no content instead of a StatusOK
	http.MethodDelete: append(universalStatusCodes,
		http.StatusNoContent),

	http.MethodOptions: {
		http.StatusNoContent,
		http.StatusForbidden,
		http.StatusMethodNotAllowed,
	},
}

// ApiMetricHandler provides a metric handler which measures
// 1. The response size
// 2. The request size
// 3. The request latency
// and attaches status code, method, and path labels for each of these
// measurements.
func ApiMetricHandler(wrapped http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p := invalidPathValue
		for r, ep := range pathRegex {
			if r.Match([]byte(req.URL.Path)) {
				p = ep
				break
			}
		}
		l := prometheus.Labels{
			labelHttpPath: p,
		}
		promhttp.InstrumentHandlerDuration(
			httpRequestLatency.MustCurryWith(l),
			promhttp.InstrumentHandlerRequestSize(
				httpResponseSize.MustCurryWith(l),
				promhttp.InstrumentHandlerResponseSize(
					httpResponseSize.MustCurryWith(l),
					wrapped,
				),
			),
		).ServeHTTP(rw, req)
	})
}

// RegisterMetrics registers the controller metrics and initializes them to 0
// for all possible label combinations.
func RegisterMetrics(r prometheus.Registerer) {
	r.MustRegister(httpResponseSize, httpRequestSize, httpRequestLatency)

	for p, methods := range expectedPathsToMethods {
		for _, m := range methods {
			for _, sc := range expectedStatusCodesPerMethod[m] {
				l := prometheus.Labels{labelHttpCode: strconv.Itoa(sc), labelHttpPath: p, labelHttpMethod: strings.ToLower(m)}
				httpResponseSize.With(l)
				httpRequestSize.With(l)
				httpRequestLatency.With(l)
			}
		}
	}

	// When an invalid path is found, any method is possible both we expect
	// an error response.
	p := invalidPathValue
	for m := range expectedStatusCodesPerMethod {
		for _, sc := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
			l := prometheus.Labels{labelHttpCode: strconv.Itoa(sc), labelHttpPath: p, labelHttpMethod: strings.ToLower(m)}
			httpResponseSize.With(l)
			httpRequestSize.With(l)
			httpRequestLatency.With(l)
		}
	}
}
