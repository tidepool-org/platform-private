package events

import (
	"context"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	lognull "github.com/tidepool-org/platform/log/null"
)

// NewPeriodicDecorator builds a periodic-event-running SaramaEventsRunner appropriate to
// the runtime environment.
//
// This provides abstraction for SaramaEventsRunner to not need to know about the runtime
// environment (Kubernetes, testing, or otherwise).
func NewPeriodicDecorator(leaseName string, period time.Duration,
	task func(context.Context) error, runner SaramaEventsRunner, logger log.Logger) SaramaEventsRunner {

	logRunner := logDecorator{logger: logger, runner: runner}
	solo := NewPeriodicDecoratorSolo(logRunner, period, task)

	switch detectRuntimeEnv() {
	case "kubernetes":
		return NewPeriodicDecoratorK8s(leaseName, solo)
	default:
		return solo
	}
}

func detectRuntimeEnv() string {
	if os.Getenv("POD_NAME") != "" && os.Getenv("POD_NAMESPACE") != "" {
		return "kubernetes"
	}
	return ""
}

// PeriodicDecoratorK8s decorates a [SaramaEventsRunner] to run periodic tasks.
//
// These tasks should run on only one instance. However, in production multiple instances of
// the data service will be running. A Kubernetes Lease[^1] facilitates leader election so
// tasks are run on only one of the running instances, and new code deployments won't
// interrupt the processes.
//
// [^1]: https://kubernetes.io/docs/concepts/architecture/leases/
type PeriodicDecoratorK8s struct {
	LeaseName string
	SaramaEventsRunner
	identity  string
	namespace string
	stop      chan struct{}
	periodicDecoratorLog
}

func NewPeriodicDecoratorK8s(leaseName string, runner SaramaEventsRunner) *PeriodicDecoratorK8s {
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	identity := os.Getenv("POD_NAME")
	if identity == "" {
		identity = "default"
	}
	fields := log.Fields{"identity": identity, "lease": leaseName}
	return &PeriodicDecoratorK8s{
		LeaseName:            leaseName,
		SaramaEventsRunner:   runner,
		identity:             identity,
		namespace:            namespace,
		periodicDecoratorLog: periodicDecoratorLog(fields),
	}
}

func (d *PeriodicDecoratorK8s) Run(ctx context.Context) (err error) {
	elector, err := d.buildLeaderElector(ctx)
	if err != nil {
		return err
	}

	done := ctx.Done()
	for {
		select {
		case <-done:
			return nil
		default:
			// elector.Run will return if canceled, or if it was, but no longer is
			// leader. In the latter case, we should just loop around again.
			elector.Run(ctx)
			d.log(ctx).Infof("leader election looping")
		}
	}
}

func (d *PeriodicDecoratorK8s) buildLeaderElector(ctx context.Context) (*leaderelection.LeaderElector, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "getting k8s cluster config")
	}
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "getting k8s client")
	}
	leaseLock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      d.LeaseName + "-periodic-decorator",
			Namespace: d.namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: d.identity,
		},
	}
	lec := leaderelection.LeaderElectionConfig{
		Lock: leaseLock,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: d.onStartedLeading,
			OnStoppedLeading: d.onStoppedLeadingFunc(ctx),
			OnNewLeader:      d.onNewLeaderFunc(ctx),
		},
	}
	leaderElector, err := leaderelection.NewLeaderElector(lec)
	if err != nil {
		return nil, errors.Wrap(err, "building leader elector")
	}
	lec.WatchDog.SetLeaderElection(leaderElector)

	return leaderElector, nil
}

func (d *PeriodicDecoratorK8s) onStartedLeading(ctx context.Context) {
	errors := make(chan error)
	defer close(errors)
	go func() { errors <- d.SaramaEventsRunner.Run(ctx) }()

	d.log(ctx).Infof("is now the leader")
	d.stop = make(chan struct{})
	for {
		select {
		case <-d.stop:
			return
		case err := <-errors:
			d.log(ctx).WithError(err).Warn("running periodic job")
		}
	}
}

func (d *PeriodicDecoratorK8s) onStoppedLeadingFunc(ctx context.Context) func() {
	return func() {
		if d.stop != nil {
			close(d.stop)
			d.stop = nil
		}
		d.log(ctx).Infof("I'm no longer the leader.")
	}
}

func (d *PeriodicDecoratorK8s) onNewLeaderFunc(ctx context.Context) func(string) {
	return func(id string) {
		if id == d.identity {
			d.log(ctx).Infof("I'm the new leader in town!")
		} else {
			d.log(ctx).Infof("The leader is dead. All hail the leader %q!", id)
		}
	}
}

type PeriodicDecoratorSolo struct {
	Period time.Duration
	Task   func(context.Context) error
	SaramaEventsRunner
	periodicDecoratorLog
}

func NewPeriodicDecoratorSolo(runner SaramaEventsRunner, period time.Duration,
	task func(context.Context) error) *PeriodicDecoratorSolo {

	return &PeriodicDecoratorSolo{
		Period:               period,
		Task:                 task,
		SaramaEventsRunner:   runner,
		periodicDecoratorLog: periodicDecoratorLog(log.Fields{"identity": "solo"}),
	}
}

func (d *PeriodicDecoratorSolo) Run(ctx context.Context) error {
	errors := make(chan error)
	go func() { defer close(errors); errors <- d.SaramaEventsRunner.Run(ctx) }()
	done := ctx.Done()
	nextRun := time.Now()
	for {
		select {
		case err := <-errors:
			return err
		case <-done:
			return nil
		case last := <-time.After(time.Until(nextRun)):
			d.log(ctx).Info("periodic events: Solo")
			if err := d.Task(ctx); err != nil {
				return err
			}
			nextRun = last.Add(d.Period)
		}
	}
}

// periodicDecoratorLog provides consistent logging as a mixin.
type periodicDecoratorLog log.Fields

// log returns a custom [log.Logger] from ctx.
//
// If no logger is found, it returns a [lognull.NullLogger].
func (d periodicDecoratorLog) log(ctx context.Context) log.Logger {
	if ctxLog := log.LoggerFromContext(ctx); ctxLog != nil {
		return ctxLog.WithFields(log.Fields(d))
	}
	return lognull.NewLogger()
}

// logDecorator injects a [log.Logger] into [SaramaEventRunner.Run].
type logDecorator struct {
	logger log.Logger
	runner SaramaEventsRunner
}

func (i logDecorator) Run(ctx context.Context) error {
	return i.runner.Run(log.NewContextWithLogger(ctx, i.logger))
}
