package kubectl

import (
	"context"
	"io"
	"strconv"
	"sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var colorCodes = []color.Color{
	color.LightRed,
	color.LightGreen,
	color.LightYellow,
	color.LightBlue,
	color.LightPurple,
	color.Red,
	color.Green,
	color.Yellow,
	color.Blue,
	color.Purple,
	color.Cyan,
}

type prefixedWriter struct {
	output io.Writer
	prefix string
	color  color.Color
}

// LogAggregator wraps an io.Writer and applies
// a prefix to every line written
type LogAggregator struct {
	output  io.Writer
	writers []*prefixedWriter
	lock    *sync.Mutex
}

// New creates a LogAggregator
func NewLogAggregator(output io.Writer) *LogAggregator {
	return &LogAggregator{
		output:  output,
		writers: []*prefixedWriter{},
		lock:    &sync.Mutex{},
	}
}

// GetOutput creates a new io.Writer that will have automatically
// prefix and color it's output.
func (p *LogAggregator) GetOutput(prefix string) io.Writer {
	p.lock.Lock()
	defer p.lock.Unlock()
	writer := &prefixedWriter{
		output: p.output,
		prefix: prefix,
		color:  colorCodes[len(p.writers)%len(colorCodes)],
	}
	p.writers = append(p.writers, writer)
	return writer
}

func (p prefixedWriter) Write(content []byte) (int, error) {
	n, err := p.color.Fprintf(p.output, "[%s] ", p.prefix)
	if err != nil {
		return n, errors.Wrap(err, "error writing prefix to output")
	}
	return p.output.Write(content)
}

// RolloutStatus monitors the rollout status of depoyments in the manifest list
func (k *CLI) RolloutStatus(ctx context.Context, out io.Writer, manifests ManifestList) error {
	decodedManifests := manifests.decodeManifests()
	logs := NewLogAggregator(out)

	var wg sync.WaitGroup

	for _, obj := range decodedManifests {
		wg.Add(1)
		go func(obj runtime.Object) {
			defer wg.Done()
			switch o := obj.(type) {
			case *v1.StatefulSet:
				major, _ := strconv.Atoi(k.Version().Major)
				minor, _ := strconv.Atoi(k.Version().Minor)
				if major == 1 && minor <= 11 {
					// https://github.com/kubernetes/kubernetes/issues/68573
					// Prior to 1.12 the rollout status of stateful sets was
					// buggy and would not terminate.
					color.Default.Fprintln(out, "StatefulSet rollout status is not supported pre kubernetes 1.12")
					break
				}
			case *v1.DaemonSet:
				k.Run(ctx, nil, logs.GetOutput(o.Name), "rollout", make([]string, 0), "status", o.TypeMeta.Kind, o.Name)
			case *v1.Deployment:
				k.Run(ctx, nil, logs.GetOutput(o.Name), "rollout", make([]string, 0), "status", o.TypeMeta.Kind, o.Name)
				break
			default:
				break
			}
		}(obj)
	}

	wg.Wait()
	return nil
}

// decodeManifests into a slice of k8s objects
func (l *ManifestList) decodeManifests() []runtime.Object {
	var decoded []runtime.Object

	for _, manifest := range *l {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(manifest, nil, nil)
		if err != nil {
			errors.Wrap(err, "Error while decoding manifest")
		} else {
			decoded = append(decoded, obj)
		}
	}

	return decoded
}
