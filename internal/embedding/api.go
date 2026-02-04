package embedding

type Api interface {
	Compute(text string) ([]float32, error)
}
