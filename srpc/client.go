package srpc

import (
	"context"
	"fmt"
	"math"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	kitextracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"

	"github.com/ssgohq/goten-core/logx"
)

// ClientBuilder helps construct Kitex client with common options.
type ClientBuilder struct {
	config  *ClientConfig
	options []client.Option
}

// NewClientBuilder creates a new client builder with the given configuration.
func NewClientBuilder(config *ClientConfig) *ClientBuilder {
	config.SetDefaults()
	return &ClientBuilder{
		config:  config,
		options: make([]client.Option, 0),
	}
}

// Build returns all configured client options ready to pass to the generated
// Kitex client NewClient function.
//
// Example:
//
//	builder := srpc.NewClientBuilder(&config)
//	cli, err := userservice.NewClient("user-rpc", builder.Build()...)
func (b *ClientBuilder) Build() []client.Option {
	opts := make([]client.Option, 0, 10)

	// 1. Service discovery or direct endpoints
	if resolver := b.buildResolver(); resolver != nil {
		opts = append(opts, client.WithResolver(resolver))
	} else if len(b.config.Endpoints) > 0 {
		opts = append(opts, client.WithHostPorts(b.config.Endpoints...))
	}

	// 2. Timeouts
	if b.config.Timeout.RPC > 0 {
		opts = append(opts, client.WithRPCTimeout(b.config.Timeout.RPC))
	}
	if b.config.Timeout.Connect > 0 {
		opts = append(opts, client.WithConnectTimeout(b.config.Timeout.Connect))
	}

	// 3. Load balancer
	if lb := b.buildLoadBalancer(); lb != nil {
		opts = append(opts, client.WithLoadBalancer(lb))
	}

	// 4. Retry policy
	if b.config.Retry.Enabled {
		opts = append(opts, b.buildRetryPolicy())
	}

	// 5. Circuit breaker
	if b.config.CircuitBreaker.Enabled {
		opts = append(opts, b.buildCircuitBreaker())
	}

	// 6. Connection pool (long connections)
	// Note: Long connection pooling is handled internally by Kitex
	// based on transport protocol

	// 7. OpenTelemetry tracing middleware
	// This propagates trace context from incoming requests to outgoing RPC calls
	opts = append(opts, client.WithSuite(kitextracing.NewClientSuite()))

	// 8. User-provided options
	opts = append(opts, b.options...)

	return opts
}

// WithOption adds a custom client option.
func (b *ClientBuilder) WithOption(opt client.Option) *ClientBuilder {
	b.options = append(b.options, opt)
	return b
}

// WithMiddleware adds a middleware to the client.
func (b *ClientBuilder) WithMiddleware(mw endpoint.Middleware) *ClientBuilder {
	b.options = append(b.options, client.WithMiddleware(mw))
	return b
}

// buildResolver creates a service resolver based on configuration.
func (b *ClientBuilder) buildResolver() discovery.Resolver {
	switch b.config.Discovery.Type {
	case "consul":
		return b.buildConsulResolver()
	case "etcd":
		return b.buildEtcdResolver()
	default:
		return nil
	}
}

// buildConsulResolver creates a Consul resolver.
func (b *ClientBuilder) buildConsulResolver() discovery.Resolver {
	cfg := b.config.Discovery.Consul

	r, err := consul.NewConsulResolver(cfg.Address)
	if err != nil {
		logx.Errorw("Failed to create Consul resolver", "address", cfg.Address, "error", err)
		return nil
	}

	logx.Debugw("Consul resolver created", "address", cfg.Address)
	return r
}

// buildEtcdResolver creates an etcd resolver.
func (b *ClientBuilder) buildEtcdResolver() discovery.Resolver {
	// TODO: Implement etcd resolver when needed
	logx.Warnw("Etcd resolver is not yet implemented")
	return nil
}

// buildLoadBalancer creates a load balancer based on configuration.
func (b *ClientBuilder) buildLoadBalancer() loadbalance.Loadbalancer {
	switch b.config.LoadBalancer {
	case "roundrobin":
		return loadbalance.NewWeightedRoundRobinBalancer()
	case "random":
		return loadbalance.NewWeightedRandomBalancer()
	case "consistenthash":
		return loadbalance.NewConsistBalancer(
			loadbalance.NewConsistentHashOption(
				func(ctx context.Context, req interface{}) string {
					// Default key function uses RPC method
					if ri := rpcinfo.GetRPCInfo(ctx); ri != nil {
						return ri.Invocation().MethodName()
					}
					return ""
				},
			),
		)
	default:
		return loadbalance.NewWeightedRoundRobinBalancer()
	}
}

// buildRetryPolicy creates a retry policy based on configuration.
func (b *ClientBuilder) buildRetryPolicy() client.Option {
	fp := retry.NewFailurePolicy()
	fp.WithMaxRetryTimes(b.config.Retry.MaxRetries)

	if b.config.Retry.Delay > 0 {
		fp.WithFixedBackOff(int(b.config.Retry.Delay.Milliseconds()))
	}
	if b.config.Retry.MaxDelay > 0 {
		maxDelayMs := b.config.Retry.MaxDelay.Milliseconds()
		if maxDelayMs > 0 && maxDelayMs <= math.MaxUint32 {
			fp.WithMaxDurationMS(uint32(maxDelayMs))
		}
	}

	return client.WithFailureRetry(fp)
}

// buildCircuitBreaker creates a circuit breaker based on configuration.
func (b *ClientBuilder) buildCircuitBreaker() client.Option {
	cbSuite := circuitbreak.NewCBSuite(func(ri rpcinfo.RPCInfo) string {
		// Key by service/method for per-method circuit breaking
		return ri.To().ServiceName() + "/" + ri.To().Method()
	})

	return client.WithCircuitBreaker(cbSuite)
}

// BuildClient is a convenience function that creates options for a client.
//
// Example:
//
//	opts := srpc.BuildClient(&config)
//	cli, err := userservice.NewClient("user-rpc", opts...)
func BuildClient(config *ClientConfig) []client.Option {
	return NewClientBuilder(config).Build()
}

// DirectClient creates client options for direct connection to specified endpoints.
// This bypasses service discovery.
//
// Example:
//
//	opts := srpc.DirectClient([]string{"localhost:8888"})
//	cli, err := userservice.NewClient("user-rpc", opts...)
func DirectClient(endpoints []string) []client.Option {
	config := &ClientConfig{
		Endpoints: endpoints,
	}
	config.SetDefaults()
	return NewClientBuilder(config).Build()
}

// ConsulClient creates client options for Consul-based service discovery.
//
// Example:
//
//	opts := srpc.ConsulClient("localhost:8500")
//	cli, err := userservice.NewClient("user-rpc", opts...)
func ConsulClient(consulAddr string) []client.Option {
	config := &ClientConfig{
		Discovery: DiscoveryConfig{
			Type: "consul",
			Consul: ConsulConfig{
				Address: consulAddr,
			},
		},
	}
	config.SetDefaults()
	return NewClientBuilder(config).Build()
}

// WithRetry returns a client option to enable retry with the specified max attempts.
func WithRetry(maxRetries int) client.Option {
	fp := retry.NewFailurePolicy()
	fp.WithMaxRetryTimes(maxRetries)
	return client.WithFailureRetry(fp)
}

// WithCircuitBreaker returns a client option to enable circuit breaker.
func WithCircuitBreaker() client.Option {
	cbSuite := circuitbreak.NewCBSuite(func(ri rpcinfo.RPCInfo) string {
		return ri.To().ServiceName() + "/" + ri.To().Method()
	})
	return client.WithCircuitBreaker(cbSuite)
}

// WithLoadBalancer returns a client option for the specified load balancer type.
// Supported types: "roundrobin", "random", "consistenthash"
func WithLoadBalancer(lbType string) client.Option {
	var lb loadbalance.Loadbalancer
	switch lbType {
	case "roundrobin":
		lb = loadbalance.NewWeightedRoundRobinBalancer()
	case "random":
		lb = loadbalance.NewWeightedRandomBalancer()
	default:
		lb = loadbalance.NewWeightedRoundRobinBalancer()
	}
	return client.WithLoadBalancer(lb)
}

// MustNewClient creates a new RPC client using the provided factory function and configuration.
// This is the recommended way to create RPC clients in service context.
// Panics if client creation fails, which is appropriate during service initialization.
//
// The generic type T should be the Kitex generated Client interface (e.g., userservice.Client).
// The newClientFn should be the Kitex generated NewClient function (e.g., userservice.NewClient).
//
// Example:
//
//	// In config.go - type-safe config fields
//	type Config struct {
//	    UserRpc srpc.ClientConfig `yaml:"UserRpc"`
//	}
//
//	// In service_context.go
//	type ServiceContext struct {
//	    Config  config.Config
//	    UserRpc userservice.Client
//	}
//
//	func NewServiceContext(c config.Config) *ServiceContext {
//	    return &ServiceContext{
//	        Config:  c,
//	        UserRpc: srpc.MustNewClient(userservice.NewClient, &c.UserRpc),
//	    }
//	}
func MustNewClient[T any](
	newClientFn func(string, ...client.Option) (T, error),
	cfg *ClientConfig,
) T {
	if cfg == nil {
		panic("srpc.MustNewClient: config is nil")
	}
	cfg.SetDefaults()
	opts := BuildClient(cfg)
	cli, err := newClientFn(cfg.ServiceName, opts...)
	if err != nil {
		panic(fmt.Sprintf("srpc.MustNewClient: failed to create client for %s: %v", cfg.ServiceName, err))
	}
	logx.Infow("RPC client created",
		"serviceName", cfg.ServiceName,
		"discoveryType", cfg.Discovery.Type,
	)
	return cli
}

// NewClientWithConfig creates a new RPC client using the provided factory function and configuration.
// Returns an error if client creation fails, which is useful when you need to handle errors gracefully.
//
// Example:
//
//	userRpc, err := srpc.NewClientWithConfig(userservice.NewClient, &c.UserRpc)
//	if err != nil {
//	    return nil, fmt.Errorf("failed to create user rpc client: %w", err)
//	}
func NewClientWithConfig[T any](
	newClientFn func(string, ...client.Option) (T, error),
	cfg *ClientConfig,
) (T, error) {
	var zero T
	if cfg == nil {
		return zero, fmt.Errorf("srpc.NewClientWithConfig: config is nil")
	}
	cfg.SetDefaults()
	opts := BuildClient(cfg)
	cli, err := newClientFn(cfg.ServiceName, opts...)
	if err != nil {
		return zero, fmt.Errorf("srpc.NewClientWithConfig: failed to create client for %s: %w", cfg.ServiceName, err)
	}
	logx.Infow("RPC client created",
		"serviceName", cfg.ServiceName,
		"discoveryType", cfg.Discovery.Type,
	)
	return cli, nil
}