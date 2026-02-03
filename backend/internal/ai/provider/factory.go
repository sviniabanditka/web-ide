package provider

import "context"

type ProviderType string

const (
	ProviderMiniMax ProviderType = "minimax"
	ProviderOpenAI  ProviderType = "openai"
	ProviderOllama  ProviderType = "ollama"
)

type Factory struct {
	providers map[ProviderType]func() Provider
}

func NewFactory() *Factory {
	f := &Factory{
		providers: make(map[ProviderType]func() Provider),
	}
	f.providers[ProviderMiniMax] = func() Provider { return NewMiniMax() }
	return f
}

func (f *Factory) Register(t ProviderType, fn func() Provider) {
	f.providers[t] = fn
}

func (f *Factory) Create(t ProviderType) Provider {
	if fn, ok := f.providers[t]; ok {
		return fn()
	}
	return nil
}

func Complete(ctx context.Context, providerType ProviderType, messages []Message, cfg Config) (*Response, error) {
	factory := NewFactory()

	var p Provider
	switch providerType {
	case ProviderMiniMax:
		p = factory.Create(ProviderMiniMax)
	case ProviderOllama:
		p = NewMiniMax() // Placeholder - would use Ollama
	case ProviderOpenAI:
		p = NewMiniMax() // Placeholder - would use OpenAI
	default:
		p = factory.Create(ProviderMiniMax)
	}

	if p == nil {
		return nil, nil
	}

	return p.Complete(ctx, messages, cfg)
}
